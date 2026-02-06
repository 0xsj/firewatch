package validate

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// Severity levels recognized by Firewatch.
var validSeverities = map[string]bool{
	"critical": true,
	"high":     true,
	"medium":   true,
	"low":      true,
	"info":     true,
}

// IP returns an error if raw is not a valid IPv4 or IPv6 address.
func IP(raw string) error {
	if net.ParseIP(raw) == nil {
		return fmt.Errorf("invalid IP address: %q", raw)
	}
	return nil
}

// CIDR returns an error if raw is not a valid CIDR notation.
func CIDR(raw string) error {
	if _, _, err := net.ParseCIDR(raw); err != nil {
		return fmt.Errorf("invalid CIDR: %q", raw)
	}
	return nil
}

// URL returns an error if raw is not a valid absolute URL.
func URL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid URL: %q", raw)
	}
	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("URL missing scheme or host: %q", raw)
	}
	return nil
}

// Port returns an error if port is outside the valid range.
func Port(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("port out of range (1-65535): %d", port)
	}
	return nil
}

// Severity returns an error if s is not a recognized severity level.
func Severity(s string) error {
	if !validSeverities[strings.ToLower(s)] {
		return fmt.Errorf("invalid severity %q: must be critical, high, medium, low, or info", s)
	}
	return nil
}

// NonEmpty returns an error if s is empty or whitespace-only.
func NonEmpty(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s must not be empty", field)
	}
	return nil
}

// OneOf returns an error if value is not in the allowed set.
func OneOf(field, value string, allowed []string) error {
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return fmt.Errorf("%s must be one of %v, got %q", field, allowed, value)
}
