package storage

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/0xsj/firewatch/internal/storage/models"
)

var testEventCounter atomic.Int64

func profilingTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func setupProfilingTest(t *testing.T) (*ProfilingStore, *SQLiteStore) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "profiling_test.db")
	inner, err := NewSQLite(dbPath)
	if err != nil {
		t.Fatalf("NewSQLite: %v", err)
	}
	t.Cleanup(func() { inner.Close() })

	ps := NewProfilingStore(inner, profilingTestLogger())
	return ps, inner
}

func TestProfiling_FirstEventCreatesAttacker(t *testing.T) {
	ps, inner := setupProfilingTest(t)
	ctx := context.Background()

	event := testEvent("192.168.1.1", "wordpress", "/wp-login.php", "medium")
	if err := ps.SaveEvent(ctx, event); err != nil {
		t.Fatalf("SaveEvent: %v", err)
	}

	// Wait for async profiling
	time.Sleep(100 * time.Millisecond)

	attacker, err := inner.GetAttackerByIP(ctx, "192.168.1.1")
	if err != nil {
		t.Fatalf("GetAttackerByIP: %v", err)
	}
	if attacker.IP != "192.168.1.1" {
		t.Errorf("IP = %q, want 192.168.1.1", attacker.IP)
	}
	if attacker.TotalEvents != 1 {
		t.Errorf("TotalEvents = %d, want 1", attacker.TotalEvents)
	}
	if attacker.Severity != "medium" {
		t.Errorf("Severity = %q, want medium", attacker.Severity)
	}
}

func TestProfiling_SecondEventUpdatesAttacker(t *testing.T) {
	ps, inner := setupProfilingTest(t)
	ctx := context.Background()

	event1 := testEvent("10.0.0.1", "wordpress", "/wp-login.php", "medium")
	if err := ps.SaveEvent(ctx, event1); err != nil {
		t.Fatalf("SaveEvent 1: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	event2 := testEvent("10.0.0.1", "exposure", "/.env", "high")
	event2.UserAgent = "curl/7.68"
	if err := ps.SaveEvent(ctx, event2); err != nil {
		t.Fatalf("SaveEvent 2: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	attacker, err := inner.GetAttackerByIP(ctx, "10.0.0.1")
	if err != nil {
		t.Fatalf("GetAttackerByIP: %v", err)
	}
	if attacker.TotalEvents != 2 {
		t.Errorf("TotalEvents = %d, want 2", attacker.TotalEvents)
	}
	if len(attacker.ModulesTargeted) != 2 {
		t.Errorf("ModulesTargeted = %v, want 2 modules", attacker.ModulesTargeted)
	}
}

func TestProfiling_SeverityEscalation(t *testing.T) {
	ps, inner := setupProfilingTest(t)
	ctx := context.Background()

	event1 := testEvent("172.16.0.1", "api", "/api/v1", "low")
	if err := ps.SaveEvent(ctx, event1); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)

	event2 := testEvent("172.16.0.1", "cve", "/solr/admin", "critical")
	if err := ps.SaveEvent(ctx, event2); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)

	attacker, err := inner.GetAttackerByIP(ctx, "172.16.0.1")
	if err != nil {
		t.Fatal(err)
	}
	if attacker.Severity != "critical" {
		t.Errorf("Severity = %q, want critical (escalated)", attacker.Severity)
	}
}

func TestProfiling_ModuleAggregation(t *testing.T) {
	ps, inner := setupProfilingTest(t)
	ctx := context.Background()

	modules := []string{"wordpress", "exposure", "api", "nextjs"}
	for _, mod := range modules {
		event := testEvent("10.10.10.10", mod, "/"+mod, "medium")
		if err := ps.SaveEvent(ctx, event); err != nil {
			t.Fatal(err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	time.Sleep(100 * time.Millisecond)

	attacker, err := inner.GetAttackerByIP(ctx, "10.10.10.10")
	if err != nil {
		t.Fatal(err)
	}
	if len(attacker.ModulesTargeted) != 4 {
		t.Errorf("ModulesTargeted = %v (len %d), want 4", attacker.ModulesTargeted, len(attacker.ModulesTargeted))
	}
}

func TestProfiling_AutoTagging(t *testing.T) {
	ps, inner := setupProfilingTest(t)
	ctx := context.Background()

	event := testEvent("192.168.0.1", "detection", "/", "critical")
	event.Signatures = []string{"generic-scanner-001"}
	if err := ps.SaveEvent(ctx, event); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)

	attacker, err := inner.GetAttackerByIP(ctx, "192.168.0.1")
	if err != nil {
		t.Fatal(err)
	}

	hasTag := func(tag string) bool {
		for _, t := range attacker.Tags {
			if t == tag {
				return true
			}
		}
		return false
	}

	if !hasTag("scanner") {
		t.Error("missing tag 'scanner'")
	}
	if !hasTag("high-threat") {
		t.Error("missing tag 'high-threat'")
	}
}

func TestProfiling_PathBounding(t *testing.T) {
	ps, inner := setupProfilingTest(t)
	ctx := context.Background()

	for i := 0; i < 110; i++ {
		event := testEvent("10.20.30.40", "api", "/path/"+string(rune('a'+i%26)), "low")
		if err := ps.SaveEvent(ctx, event); err != nil {
			t.Fatal(err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(200 * time.Millisecond)

	attacker, err := inner.GetAttackerByIP(ctx, "10.20.30.40")
	if err != nil {
		t.Fatal(err)
	}
	if len(attacker.PathsProbed) > 100 {
		t.Errorf("PathsProbed = %d, want <= 100", len(attacker.PathsProbed))
	}
}

func TestHigherSeverity(t *testing.T) {
	tests := []struct {
		a, b, want string
	}{
		{"info", "low", "low"},
		{"low", "info", "low"},
		{"medium", "high", "high"},
		{"critical", "low", "critical"},
		{"", "medium", "medium"},
	}
	for _, tt := range tests {
		got := higherSeverity(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("higherSeverity(%q, %q) = %q, want %q", tt.a, tt.b, got, tt.want)
		}
	}
}

func testEvent(ip, module, path, severity string) *models.Event {
	n := testEventCounter.Add(1)
	return &models.Event{
		ID:        fmt.Sprintf("evt-%s-%s-%d", ip, module, n),
		Timestamp: "2024-02-10T12:00:00Z",
		RequestID: fmt.Sprintf("req-%d", n),
		SourceIP:  ip,
		Module:    module,
		Method:    "GET",
		Path:      path,
		UserAgent: "test-agent",
		Severity:  severity,
	}
}
