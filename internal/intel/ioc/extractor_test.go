package ioc

import (
	"testing"

	"github.com/0xsj/firewatch/internal/storage/models"
)

func TestExtractor_FromEvent_ExtractsSourceIP(t *testing.T) {
	x := NewExtractor()
	event := &models.Event{
		ID:         "evt-1",
		Timestamp:  "2024-01-15T12:00:00Z",
		SourceIP:   "203.0.113.50",
		Module:     "nextjs",
		Severity:   "high",
		Signatures: []string{"nextjs-action-001"},
	}

	iocs := x.FromEvent(event)

	var found bool
	for _, ioc := range iocs {
		if ioc.Type == models.IOCTypeIP && ioc.Value == "203.0.113.50" {
			found = true
			if ioc.Severity != "high" {
				t.Errorf("IP IOC severity = %q, want high", ioc.Severity)
			}
			if len(ioc.Tags) == 0 {
				t.Error("IP IOC has no tags")
			}
		}
	}
	if !found {
		t.Error("source IP not extracted as IOC")
	}
}

func TestExtractor_FromEvent_ExtractsRefererDomain(t *testing.T) {
	x := NewExtractor()
	event := &models.Event{
		ID:        "evt-2",
		Timestamp: "2024-01-15T12:00:00Z",
		SourceIP:  "10.0.0.1",
		Module:    "wordpress",
		Severity:  "medium",
		Headers: map[string]string{
			"referer": "https://evil.example.com/scanner?target=victim",
		},
	}

	iocs := x.FromEvent(event)

	var foundDomain, foundURL bool
	for _, ioc := range iocs {
		if ioc.Type == models.IOCTypeDomain && ioc.Value == "evil.example.com" {
			foundDomain = true
		}
		if ioc.Type == models.IOCTypeURL {
			foundURL = true
		}
	}
	if !foundDomain {
		t.Error("referer domain not extracted as IOC")
	}
	if !foundURL {
		t.Error("referer URL not extracted as IOC")
	}
}

func TestExtractor_FromEvent_ExtractsXFF(t *testing.T) {
	x := NewExtractor()
	event := &models.Event{
		ID:        "evt-3",
		Timestamp: "2024-01-15T12:00:00Z",
		SourceIP:  "10.0.0.1",
		Module:    "api",
		Severity:  "low",
		Headers: map[string]string{
			"x-forwarded-for": "198.51.100.10, 203.0.113.5",
		},
	}

	iocs := x.FromEvent(event)

	xffIPs := 0
	for _, ioc := range iocs {
		if ioc.Type == models.IOCTypeIP && ioc.Value != "10.0.0.1" {
			xffIPs++
		}
	}
	if xffIPs < 2 {
		t.Errorf("extracted %d XFF IPs, want at least 2", xffIPs)
	}
}

func TestExtractor_FromEvents_Deduplicates(t *testing.T) {
	x := NewExtractor()
	events := []*models.Event{
		{
			ID:        "evt-1",
			Timestamp: "2024-01-15T12:00:00Z",
			SourceIP:  "203.0.113.50",
			Module:    "nextjs",
			Severity:  "low",
		},
		{
			ID:        "evt-2",
			Timestamp: "2024-01-15T12:05:00Z",
			SourceIP:  "203.0.113.50",
			Module:    "wordpress",
			Severity:  "high",
		},
	}

	iocs := x.FromEvents(events)

	// Same IP should appear once, with the highest severity.
	ipCount := 0
	for _, ioc := range iocs {
		if ioc.Type == models.IOCTypeIP && ioc.Value == "203.0.113.50" {
			ipCount++
			if ioc.Severity != "high" {
				t.Errorf("merged IOC severity = %q, want high (highest)", ioc.Severity)
			}
		}
	}
	if ipCount != 1 {
		t.Errorf("same IP appeared %d times, want 1 (deduplicated)", ipCount)
	}
}

func TestExtractor_FromEvent_InvalidIP(t *testing.T) {
	x := NewExtractor()
	event := &models.Event{
		ID:        "evt-4",
		Timestamp: "2024-01-15T12:00:00Z",
		SourceIP:  "not-an-ip",
		Module:    "nextjs",
		Severity:  "info",
	}

	iocs := x.FromEvent(event)
	for _, ioc := range iocs {
		if ioc.Type == models.IOCTypeIP && ioc.Value == "not-an-ip" {
			t.Error("invalid IP should not be extracted as IOC")
		}
	}
}
