package netutil

import (
	"context"
	"net"
	"time"
)

// DefaultDNSTimeout is the default timeout for DNS lookups.
const DefaultDNSTimeout = 5 * time.Second

// ReverseLookup performs a reverse DNS lookup for the given IP address.
// Returns the first hostname found, or empty string on failure.
func ReverseLookup(ip string) string {
	names, _ := ReverseLookupAll(ip)
	if len(names) > 0 {
		return names[0]
	}
	return ""
}

// ReverseLookupAll performs a reverse DNS lookup and returns all hostnames.
func ReverseLookupAll(ip string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultDNSTimeout)
	defer cancel()
	return ReverseLookupContext(ctx, ip)
}

// ReverseLookupContext performs a reverse DNS lookup using the provided context.
func ReverseLookupContext(ctx context.Context, ip string) ([]string, error) {
	resolver := net.DefaultResolver
	names, err := resolver.LookupAddr(ctx, ip)
	if err != nil {
		return nil, err
	}

	// Strip trailing dots from FQDNs
	for i, name := range names {
		if len(name) > 0 && name[len(name)-1] == '.' {
			names[i] = name[:len(name)-1]
		}
	}
	return names, nil
}
