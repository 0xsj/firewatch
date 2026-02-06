package ioc

import (
	"net"
	"net/url"
	"strings"

	"github.com/0xsj/firewatch/internal/storage/models"
	"github.com/0xsj/firewatch/pkg/crypto"
	"github.com/0xsj/firewatch/pkg/netutil"
	"github.com/0xsj/firewatch/pkg/timeutil"
)

// Extractor pulls IOCs from honeypot events.
type Extractor struct{}

// NewExtractor creates an IOC extractor.
func NewExtractor() *Extractor {
	return &Extractor{}
}

// FromEvent extracts all IOCs from a single event.
func (x *Extractor) FromEvent(event *models.Event) []*models.IOC {
	var iocs []*models.IOC
	now := timeutil.FormatRFC3339(timeutil.NowUTC())

	// Source IP is always an IOC
	if event.SourceIP != "" && netutil.IsValidIP(event.SourceIP) {
		iocs = append(iocs, &models.IOC{
			ID:        crypto.UUID4(),
			Type:      models.IOCTypeIP,
			Value:     netutil.NormalizeIP(event.SourceIP),
			FirstSeen: event.Timestamp,
			LastSeen:  now,
			Source:    event.ID,
			Severity:  event.Severity,
			Tags:      tagsForEvent(event),
		})
	}

	// Extract domains/URLs from headers
	if referer, ok := event.Headers["referer"]; ok && referer != "" {
		if u, err := url.Parse(referer); err == nil && u.Host != "" {
			host := stripPort(u.Host)
			if net.ParseIP(host) == nil {
				// It's a domain, not an IP
				iocs = append(iocs, &models.IOC{
					ID:        crypto.UUID4(),
					Type:      models.IOCTypeDomain,
					Value:     strings.ToLower(host),
					FirstSeen: event.Timestamp,
					LastSeen:  now,
					Source:    event.ID,
					Severity:  event.Severity,
				})
			}
			iocs = append(iocs, &models.IOC{
				ID:        crypto.UUID4(),
				Type:      models.IOCTypeURL,
				Value:     referer,
				FirstSeen: event.Timestamp,
				LastSeen:  now,
				Source:    event.ID,
				Severity:  event.Severity,
			})
		}
	}

	// Extract forwarded IPs (X-Forwarded-For chain)
	if xff, ok := event.Headers["x-forwarded-for"]; ok && xff != "" {
		for _, raw := range strings.Split(xff, ",") {
			ip := strings.TrimSpace(raw)
			if netutil.IsValidIP(ip) && ip != event.SourceIP {
				iocs = append(iocs, &models.IOC{
					ID:        crypto.UUID4(),
					Type:      models.IOCTypeIP,
					Value:     netutil.NormalizeIP(ip),
					FirstSeen: event.Timestamp,
					LastSeen:  now,
					Source:    event.ID,
					Severity:  "low",
					Tags:      []string{"xff-chain"},
				})
			}
		}
	}

	return iocs
}

// FromEvents extracts and deduplicates IOCs across multiple events.
func (x *Extractor) FromEvents(events []*models.Event) []*models.IOC {
	// Key: type+value → merged IOC
	seen := make(map[string]*models.IOC)

	for _, event := range events {
		for _, ioc := range x.FromEvent(event) {
			key := string(ioc.Type) + ":" + ioc.Value
			if existing, ok := seen[key]; ok {
				// Merge: update last_seen, keep earliest first_seen
				existing.LastSeen = ioc.LastSeen
				if ioc.FirstSeen < existing.FirstSeen {
					existing.FirstSeen = ioc.FirstSeen
				}
				// Keep highest severity
				if severityRank[ioc.Severity] > severityRank[existing.Severity] {
					existing.Severity = ioc.Severity
				}
				// Merge tags
				existing.Tags = mergeTags(existing.Tags, ioc.Tags)
			} else {
				seen[key] = ioc
			}
		}
	}

	result := make([]*models.IOC, 0, len(seen))
	for _, ioc := range seen {
		result = append(result, ioc)
	}
	return result
}

// tagsForEvent generates classification tags based on event data.
func tagsForEvent(event *models.Event) []string {
	var tags []string
	tags = append(tags, event.Module)
	for _, sig := range event.Signatures {
		tags = append(tags, sig)
	}
	return tags
}

// stripPort removes the port from a host:port string.
func stripPort(host string) string {
	h, _, err := net.SplitHostPort(host)
	if err != nil {
		return host // No port present
	}
	return h
}

// mergeTags combines two tag slices, deduplicating.
func mergeTags(a, b []string) []string {
	seen := make(map[string]struct{}, len(a)+len(b))
	for _, t := range a {
		seen[t] = struct{}{}
	}
	for _, t := range b {
		seen[t] = struct{}{}
	}
	result := make([]string, 0, len(seen))
	for t := range seen {
		result = append(result, t)
	}
	return result
}

var severityRank = map[string]int{
	"info":     0,
	"low":      1,
	"medium":   2,
	"high":     3,
	"critical": 4,
}
