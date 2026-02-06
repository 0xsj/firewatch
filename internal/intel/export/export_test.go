package export

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/0xsj/firewatch/internal/storage/models"
)

var testIOCs = []*models.IOC{
	{
		ID:        "ioc-1",
		Type:      models.IOCTypeIP,
		Value:     "203.0.113.50",
		FirstSeen: "2024-01-15T12:00:00Z",
		LastSeen:  "2024-01-15T14:00:00Z",
		Severity:  "high",
		Tags:      []string{"nextjs", "scanner"},
	},
	{
		ID:        "ioc-2",
		Type:      models.IOCTypeDomain,
		Value:     "evil.example.com",
		FirstSeen: "2024-01-15T13:00:00Z",
		LastSeen:  "2024-01-15T13:30:00Z",
		Severity:  "medium",
	},
}

var testCampaigns = []*models.Campaign{
	{
		ID:              "camp-1",
		Name:            "test-campaign",
		FirstSeen:       "2024-01-15T12:00:00Z",
		LastSeen:        "2024-01-15T14:00:00Z",
		AttackerIPs:     []string{"203.0.113.50", "198.51.100.10"},
		AttackerCount:   2,
		EventCount:      15,
		ModulesTargeted: []string{"nextjs", "wordpress"},
		Severity:        "high",
	},
}

func TestSTIX_ExportIOCs(t *testing.T) {
	s := NewSTIX()

	if s.Name() != "stix" {
		t.Errorf("Name() = %q, want stix", s.Name())
	}
	if s.ContentType() != "application/json" {
		t.Errorf("ContentType() = %q, want application/json", s.ContentType())
	}

	data, err := s.ExportIOCs(testIOCs)
	if err != nil {
		t.Fatalf("ExportIOCs: %v", err)
	}

	var bundle map[string]any
	if err := json.Unmarshal(data, &bundle); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if bundle["type"] != "bundle" {
		t.Errorf("type = %v, want bundle", bundle["type"])
	}
	objects, ok := bundle["objects"].([]any)
	if !ok || len(objects) != 2 {
		t.Errorf("objects count = %d, want 2", len(objects))
	}

	// Check first indicator has STIX pattern.
	indicator := objects[0].(map[string]any)
	pattern, _ := indicator["pattern"].(string)
	if !strings.Contains(pattern, "203.0.113.50") {
		t.Errorf("pattern = %q, should contain IP", pattern)
	}
}

func TestSTIX_ExportCampaigns(t *testing.T) {
	s := NewSTIX()
	data, err := s.ExportCampaigns(testCampaigns)
	if err != nil {
		t.Fatalf("ExportCampaigns: %v", err)
	}

	var bundle map[string]any
	if err := json.Unmarshal(data, &bundle); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	objects := bundle["objects"].([]any)
	if len(objects) != 1 {
		t.Errorf("objects = %d, want 1", len(objects))
	}
}

func TestMISP_ExportIOCs(t *testing.T) {
	m := NewMISP()

	if m.Name() != "misp" {
		t.Errorf("Name() = %q, want misp", m.Name())
	}

	data, err := m.ExportIOCs(testIOCs)
	if err != nil {
		t.Fatalf("ExportIOCs: %v", err)
	}

	var event map[string]any
	if err := json.Unmarshal(data, &event); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	attrs, ok := event["Attribute"].([]any)
	if !ok || len(attrs) != 2 {
		t.Errorf("attributes = %d, want 2", len(attrs))
	}
}

func TestCSV_ExportIOCs(t *testing.T) {
	c := NewCSV()

	if c.Name() != "csv" {
		t.Errorf("Name() = %q, want csv", c.Name())
	}
	if c.ContentType() != "text/csv" {
		t.Errorf("ContentType() = %q, want text/csv", c.ContentType())
	}

	data, err := c.ExportIOCs(testIOCs)
	if err != nil {
		t.Fatalf("ExportIOCs: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 { // header + 2 rows
		t.Errorf("lines = %d, want 3 (header + 2 rows)", len(lines))
	}
	if !strings.Contains(lines[0], "id,type,value") {
		t.Errorf("header = %q, missing expected columns", lines[0])
	}
	if !strings.Contains(lines[1], "203.0.113.50") {
		t.Errorf("row 1 = %q, missing IP", lines[1])
	}
}

func TestCSV_ExportCampaigns(t *testing.T) {
	c := NewCSV()
	data, err := c.ExportCampaigns(testCampaigns)
	if err != nil {
		t.Fatalf("ExportCampaigns: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 { // header + 1 row
		t.Errorf("lines = %d, want 2", len(lines))
	}
}
