package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/0xsj/firewatch/internal/storage/models"
)

func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	store, err := NewSQLite(path)
	if err != nil {
		t.Fatalf("NewSQLite: %v", err)
	}
	t.Cleanup(func() {
		store.Close()
		os.Remove(path)
	})
	return store
}

func TestSQLiteStore_SaveAndGetEvent(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	event := &models.Event{
		ID:         "evt-001",
		Timestamp:  "2024-01-15T12:00:00Z",
		RequestID:  "req-001",
		SourceIP:   "203.0.113.50",
		Module:     "nextjs",
		Method:     "POST",
		Path:       "/",
		Severity:   "high",
		Signatures: []string{"nextjs-action-001"},
		Headers:    map[string]string{"User-Agent": "curl/8.0"},
	}

	if err := store.SaveEvent(ctx, event); err != nil {
		t.Fatalf("SaveEvent: %v", err)
	}

	got, err := store.GetEvent(ctx, "evt-001")
	if err != nil {
		t.Fatalf("GetEvent: %v", err)
	}

	if got.ID != event.ID {
		t.Errorf("ID = %q, want %q", got.ID, event.ID)
	}
	if got.SourceIP != "203.0.113.50" {
		t.Errorf("SourceIP = %q, want 203.0.113.50", got.SourceIP)
	}
	if got.Module != "nextjs" {
		t.Errorf("Module = %q, want nextjs", got.Module)
	}
	if got.Severity != "high" {
		t.Errorf("Severity = %q, want high", got.Severity)
	}
	if len(got.Signatures) != 1 || got.Signatures[0] != "nextjs-action-001" {
		t.Errorf("Signatures = %v, want [nextjs-action-001]", got.Signatures)
	}
}

func TestSQLiteStore_ListEvents(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	events := []*models.Event{
		{ID: "evt-1", Timestamp: "2024-01-15T12:00:00Z", RequestID: "r1", SourceIP: "10.0.0.1", Module: "nextjs", Method: "GET", Path: "/", Severity: "info"},
		{ID: "evt-2", Timestamp: "2024-01-15T12:01:00Z", RequestID: "r2", SourceIP: "10.0.0.2", Module: "wordpress", Method: "GET", Path: "/wp-login.php", Severity: "medium"},
		{ID: "evt-3", Timestamp: "2024-01-15T12:02:00Z", RequestID: "r3", SourceIP: "10.0.0.1", Module: "nextjs", Method: "POST", Path: "/", Severity: "high"},
	}
	for _, e := range events {
		if err := store.SaveEvent(ctx, e); err != nil {
			t.Fatalf("SaveEvent(%s): %v", e.ID, err)
		}
	}

	// List all.
	all, err := store.ListEvents(ctx, EventFilter{})
	if err != nil {
		t.Fatalf("ListEvents: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("ListEvents() count = %d, want 3", len(all))
	}

	// Filter by module.
	nextjsEvents, err := store.ListEvents(ctx, EventFilter{Module: "nextjs"})
	if err != nil {
		t.Fatalf("ListEvents(nextjs): %v", err)
	}
	if len(nextjsEvents) != 2 {
		t.Errorf("ListEvents(nextjs) count = %d, want 2", len(nextjsEvents))
	}

	// Filter by IP.
	ipEvents, err := store.ListEvents(ctx, EventFilter{SourceIP: "10.0.0.2"})
	if err != nil {
		t.Fatalf("ListEvents(ip): %v", err)
	}
	if len(ipEvents) != 1 {
		t.Errorf("ListEvents(ip) count = %d, want 1", len(ipEvents))
	}

	// Count.
	count, err := store.CountEvents(ctx, EventFilter{})
	if err != nil {
		t.Fatalf("CountEvents: %v", err)
	}
	if count != 3 {
		t.Errorf("CountEvents() = %d, want 3", count)
	}

	// Limit.
	limited, err := store.ListEvents(ctx, EventFilter{Limit: 1})
	if err != nil {
		t.Fatalf("ListEvents(limit): %v", err)
	}
	if len(limited) != 1 {
		t.Errorf("ListEvents(limit=1) count = %d, want 1", len(limited))
	}
}

func TestSQLiteStore_SaveAndGetAttacker(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	attacker := &models.Attacker{
		ID:              "atk-001",
		FirstSeen:       "2024-01-15T12:00:00Z",
		LastSeen:        "2024-01-15T14:00:00Z",
		IP:              "203.0.113.50",
		TotalEvents:     15,
		ModulesTargeted: []string{"nextjs", "wordpress"},
		Severity:        "high",
		Tags:            []string{"scanner"},
	}

	if err := store.SaveAttacker(ctx, attacker); err != nil {
		t.Fatalf("SaveAttacker: %v", err)
	}

	got, err := store.GetAttackerByIP(ctx, "203.0.113.50")
	if err != nil {
		t.Fatalf("GetAttackerByIP: %v", err)
	}
	if got.IP != "203.0.113.50" {
		t.Errorf("IP = %q, want 203.0.113.50", got.IP)
	}
	if got.TotalEvents != 15 {
		t.Errorf("TotalEvents = %d, want 15", got.TotalEvents)
	}
}

func TestSQLiteStore_SaveIOC(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	ioc := &models.IOC{
		ID:        "ioc-001",
		Type:      models.IOCTypeIP,
		Value:     "203.0.113.50",
		FirstSeen: "2024-01-15T12:00:00Z",
		LastSeen:  "2024-01-15T14:00:00Z",
		Severity:  "high",
		Tags:      []string{"scanner"},
	}

	if err := store.SaveIOC(ctx, ioc); err != nil {
		t.Fatalf("SaveIOC: %v", err)
	}

	iocs, err := store.ListIOCs(ctx, IOCFilter{Type: models.IOCTypeIP})
	if err != nil {
		t.Fatalf("ListIOCs: %v", err)
	}
	if len(iocs) != 1 {
		t.Fatalf("ListIOCs count = %d, want 1", len(iocs))
	}
	if iocs[0].Value != "203.0.113.50" {
		t.Errorf("IOC value = %q, want 203.0.113.50", iocs[0].Value)
	}
}
