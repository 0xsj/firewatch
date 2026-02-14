package middleware

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/0xsj/firewatch/internal/storage"
	"github.com/0xsj/firewatch/internal/storage/models"
	"github.com/0xsj/firewatch/pkg/crypto"
	"github.com/0xsj/firewatch/pkg/httputil"
	"github.com/0xsj/firewatch/pkg/timeutil"
)

// IPFilterConfig holds parsed allowlist and blocklist entries.
type IPFilterConfig struct {
	Allowlist []*net.IPNet
	Blocklist []*net.IPNet
}

// ParseIPFilter parses string lists of IPs and CIDRs into IPFilterConfig.
func ParseIPFilter(allowlist, blocklist []string) (*IPFilterConfig, error) {
	cfg := &IPFilterConfig{}

	for _, entry := range allowlist {
		ipNet, err := parseIPOrCIDR(entry)
		if err != nil {
			return nil, fmt.Errorf("allowlist entry %q: %w", entry, err)
		}
		cfg.Allowlist = append(cfg.Allowlist, ipNet)
	}

	for _, entry := range blocklist {
		ipNet, err := parseIPOrCIDR(entry)
		if err != nil {
			return nil, fmt.Errorf("blocklist entry %q: %w", entry, err)
		}
		cfg.Blocklist = append(cfg.Blocklist, ipNet)
	}

	return cfg, nil
}

// LoadIPListFile reads IPs/CIDRs from a file (one per line, # comments).
func LoadIPListFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening IP list file: %w", err)
	}
	defer f.Close()

	var entries []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		entries = append(entries, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading IP list file: %w", err)
	}
	return entries, nil
}

// IsAllowed checks if an IP is in the allowlist.
func (c *IPFilterConfig) IsAllowed(ip net.IP) bool {
	for _, ipNet := range c.Allowlist {
		if ipNet.Contains(ip) {
			return true
		}
	}
	return false
}

// IsBlocked checks if an IP is in the blocklist.
func (c *IPFilterConfig) IsBlocked(ip net.IP) bool {
	for _, ipNet := range c.Blocklist {
		if ipNet.Contains(ip) {
			return true
		}
	}
	return false
}

// IPFilter returns middleware that blocks/allows IPs based on the filter config.
// Allowed IPs always pass through. Blocked IPs get a 403 and an event recorded.
// IPs not in either list pass through normally.
func IPFilter(cfg *IPFilterConfig, store storage.Store, logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg == nil || (len(cfg.Allowlist) == 0 && len(cfg.Blocklist) == 0) {
				next.ServeHTTP(w, r)
				return
			}

			ipStr := httputil.ClientIP(r)
			ip := net.ParseIP(ipStr)
			if ip == nil {
				next.ServeHTTP(w, r)
				return
			}

			// Allowlist takes precedence
			if cfg.IsAllowed(ip) {
				next.ServeHTTP(w, r)
				return
			}

			if cfg.IsBlocked(ip) {
				logger.Warn("blocked IP",
					"ip", ipStr,
					"path", r.URL.Path,
					"request_id", RequestID(r.Context()),
				)

				event := &models.Event{
					ID:         crypto.UUID4(),
					Timestamp:  timeutil.FormatRFC3339(timeutil.NowUTC()),
					RequestID:  RequestID(r.Context()),
					SourceIP:   ipStr,
					Module:     "ip_filter",
					Method:     r.Method,
					Path:       r.URL.Path,
					Query:      r.URL.RawQuery,
					Headers:    httputil.HeaderMap(r.Header),
					UserAgent:  r.UserAgent(),
					Severity:   "high",
					Signatures: []string{"ip-blocklist-match"},
				}

				if err := store.SaveEvent(r.Context(), event); err != nil {
					logger.Error("failed to save IP filter event",
						"error", err,
						"event_id", event.ID,
					)
				}

				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// parseIPOrCIDR parses a string as a CIDR range or single IP.
func parseIPOrCIDR(s string) (*net.IPNet, error) {
	if strings.Contains(s, "/") {
		_, ipNet, err := net.ParseCIDR(s)
		if err != nil {
			return nil, err
		}
		return ipNet, nil
	}

	ip := net.ParseIP(s)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", s)
	}

	if ip.To4() != nil {
		return &net.IPNet{IP: ip, Mask: net.CIDRMask(32, 32)}, nil
	}
	return &net.IPNet{IP: ip, Mask: net.CIDRMask(128, 128)}, nil
}
