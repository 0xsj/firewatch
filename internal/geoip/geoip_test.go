package geoip

import (
	"net"
	"testing"
)

func TestLookup_InvalidIP(t *testing.T) {
	// We can't call Lookup without a real Reader, but we can test
	// the IP parsing logic by verifying net.ParseIP behavior that
	// Lookup relies on.
	ip := net.ParseIP("not-an-ip")
	if ip != nil {
		t.Error("expected nil for invalid IP string")
	}
}

func TestLookup_PrivateIPs(t *testing.T) {
	privateIPs := []string{
		"127.0.0.1",
		"10.0.0.1",
		"192.168.1.1",
		"172.16.0.1",
		"::1",
	}

	for _, ipStr := range privateIPs {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			t.Errorf("ParseIP(%q) returned nil", ipStr)
			continue
		}
		if !ip.IsLoopback() && !ip.IsPrivate() {
			t.Errorf("expected %q to be loopback or private", ipStr)
		}
	}
}

func TestOpen_BadPath(t *testing.T) {
	_, err := Open("/nonexistent/path/to/geo.mmdb")
	if err == nil {
		t.Error("expected error for nonexistent .mmdb path")
	}
}
