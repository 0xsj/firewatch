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

func TestSQLiteStore_SaveAndGetHoneyToken(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	token := &models.HoneyToken{
		ID:        "ht-001",
		Kind:      "aws_access_key",
		Value:     "AKIAIOSFODNN7EXAMPLE",
		IssuedAt:  "2024-01-15T12:00:00Z",
		SourceIP:  "203.0.113.50",
		Module:    "cloud",
		Path:      "/latest/meta-data/iam/security-credentials/",
		RequestID: "req-abc",
	}

	if err := store.SaveHoneyToken(ctx, token); err != nil {
		t.Fatalf("SaveHoneyToken: %v", err)
	}

	got, err := store.GetHoneyTokenByValue(ctx, "AKIAIOSFODNN7EXAMPLE")
	if err != nil {
		t.Fatalf("GetHoneyTokenByValue: %v", err)
	}
	if got == nil {
		t.Fatal("GetHoneyTokenByValue returned nil")
	}
	if got.ID != "ht-001" {
		t.Errorf("ID = %q, want ht-001", got.ID)
	}
	if got.Kind != "aws_access_key" {
		t.Errorf("Kind = %q, want aws_access_key", got.Kind)
	}
	if got.SourceIP != "203.0.113.50" {
		t.Errorf("SourceIP = %q, want 203.0.113.50", got.SourceIP)
	}
	if got.Module != "cloud" {
		t.Errorf("Module = %q, want cloud", got.Module)
	}
	if got.RequestID != "req-abc" {
		t.Errorf("RequestID = %q, want req-abc", got.RequestID)
	}
}

func TestSQLiteStore_GetHoneyTokenByValue_NotFound(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	got, err := store.GetHoneyTokenByValue(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent token, got nil")
	}
	if got != nil {
		t.Errorf("expected nil token, got %+v", got)
	}
}

func TestSQLiteStore_ListHoneyTokens(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	tokens := []*models.HoneyToken{
		{ID: "ht-1", Kind: "aws_access_key", Value: "AKIA1", IssuedAt: "2024-01-15T12:00:00Z", SourceIP: "10.0.0.1", Module: "cloud", Path: "/iam/", RequestID: "r1"},
		{ID: "ht-2", Kind: "aws_secret_key", Value: "secret1", IssuedAt: "2024-01-15T12:00:00Z", SourceIP: "10.0.0.1", Module: "cloud", Path: "/iam/", RequestID: "r1"},
		{ID: "ht-3", Kind: "db_password", Value: "dbpass1", IssuedAt: "2024-01-15T12:01:00Z", SourceIP: "10.0.0.2", Module: "exposure", Path: "/.env", RequestID: "r2"},
		{ID: "ht-4", Kind: "aws_access_key", Value: "AKIA2", IssuedAt: "2024-01-15T12:02:00Z", SourceIP: "10.0.0.3", Module: "cloud", Path: "/iam/", RequestID: "r3"},
	}
	for _, tk := range tokens {
		if err := store.SaveHoneyToken(ctx, tk); err != nil {
			t.Fatalf("SaveHoneyToken(%s): %v", tk.ID, err)
		}
	}

	// List all.
	all, err := store.ListHoneyTokens(ctx, HoneyTokenFilter{})
	if err != nil {
		t.Fatalf("ListHoneyTokens: %v", err)
	}
	if len(all) != 4 {
		t.Errorf("ListHoneyTokens() count = %d, want 4", len(all))
	}

	// Filter by kind.
	accessKeys, err := store.ListHoneyTokens(ctx, HoneyTokenFilter{Kind: "aws_access_key"})
	if err != nil {
		t.Fatalf("ListHoneyTokens(kind): %v", err)
	}
	if len(accessKeys) != 2 {
		t.Errorf("ListHoneyTokens(kind=aws_access_key) count = %d, want 2", len(accessKeys))
	}

	// Filter by module.
	exposureTokens, err := store.ListHoneyTokens(ctx, HoneyTokenFilter{Module: "exposure"})
	if err != nil {
		t.Fatalf("ListHoneyTokens(module): %v", err)
	}
	if len(exposureTokens) != 1 {
		t.Errorf("ListHoneyTokens(module=exposure) count = %d, want 1", len(exposureTokens))
	}

	// Filter by source IP.
	ipTokens, err := store.ListHoneyTokens(ctx, HoneyTokenFilter{SourceIP: "10.0.0.1"})
	if err != nil {
		t.Fatalf("ListHoneyTokens(ip): %v", err)
	}
	if len(ipTokens) != 2 {
		t.Errorf("ListHoneyTokens(ip=10.0.0.1) count = %d, want 2", len(ipTokens))
	}

	// Limit.
	limited, err := store.ListHoneyTokens(ctx, HoneyTokenFilter{Limit: 2})
	if err != nil {
		t.Fatalf("ListHoneyTokens(limit): %v", err)
	}
	if len(limited) != 2 {
		t.Errorf("ListHoneyTokens(limit=2) count = %d, want 2", len(limited))
	}

	// Offset.
	offset, err := store.ListHoneyTokens(ctx, HoneyTokenFilter{Limit: 2, Offset: 2})
	if err != nil {
		t.Fatalf("ListHoneyTokens(offset): %v", err)
	}
	if len(offset) != 2 {
		t.Errorf("ListHoneyTokens(limit=2,offset=2) count = %d, want 2", len(offset))
	}
}
