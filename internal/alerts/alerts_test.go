package alerts

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestMeetsSeverity(t *testing.T) {
	tests := []struct {
		alert string
		min   string
		want  bool
	}{
		{"critical", "info", true},
		{"critical", "critical", true},
		{"high", "critical", false},
		{"medium", "medium", true},
		{"low", "medium", false},
		{"info", "info", true},
		{"info", "low", false},
	}

	for _, tt := range tests {
		got := MeetsSeverity(tt.alert, tt.min)
		if got != tt.want {
			t.Errorf("MeetsSeverity(%q, %q) = %v, want %v", tt.alert, tt.min, got, tt.want)
		}
	}
}

func TestManager_EmptyAlerts(t *testing.T) {
	m := NewManager(testLogger(), 0)
	if m.Count() != 0 {
		t.Errorf("Count() = %d, want 0", m.Count())
	}
}

func TestManager_Register(t *testing.T) {
	m := NewManager(testLogger(), 0)
	m.Register(NewSlack("https://hooks.slack.com/test"), "medium")
	m.Register(NewDiscord("https://discord.com/api/webhooks/test"), "high")

	if m.Count() != 2 {
		t.Errorf("Count() = %d, want 2", m.Count())
	}
}

// recordingAlerter captures sent alerts for test assertions.
type recordingAlerter struct {
	mu    sync.Mutex
	calls []Alert
}

func (r *recordingAlerter) Name() string { return "test" }

func (r *recordingAlerter) Send(_ context.Context, a Alert) error {
	r.mu.Lock()
	r.calls = append(r.calls, a)
	r.mu.Unlock()
	return nil
}

func (r *recordingAlerter) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.calls)
}

func TestDedupKey(t *testing.T) {
	a1 := Alert{SourceIP: "1.2.3.4", Module: "nextjs", Signatures: []string{"sig-001"}}
	a2 := Alert{SourceIP: "5.6.7.8", Module: "nextjs", Signatures: []string{"sig-001"}}
	a3 := Alert{SourceIP: "1.2.3.4", Module: "wordpress", Signatures: []string{"sig-001"}}
	a4 := Alert{SourceIP: "1.2.3.4", Module: "nextjs", Path: "/login"}

	k1 := dedupKey(a1)
	k2 := dedupKey(a2)
	k3 := dedupKey(a3)
	k4 := dedupKey(a4)

	// Same IP+module+sig → same key.
	if k1 != dedupKey(a1) {
		t.Error("identical alerts should produce the same key")
	}
	// Different IP → different key.
	if k1 == k2 {
		t.Error("different IPs should produce different keys")
	}
	// Different module → different key.
	if k1 == k3 {
		t.Error("different modules should produce different keys")
	}
	// No signatures → falls back to path.
	if k4 != "1.2.3.4|nextjs|/login" {
		t.Errorf("dedupKey without signatures = %q, want 1.2.3.4|nextjs|/login", k4)
	}
}

func TestManager_DedupSuppresses(t *testing.T) {
	rec := &recordingAlerter{}
	m := NewManager(testLogger(), 5*time.Minute)
	defer m.Stop()
	m.Register(rec, "info")

	alert := Alert{
		ID:         "a1",
		Severity:   "high",
		Module:     "nextjs",
		SourceIP:   "1.2.3.4",
		Signatures: []string{"sig-001"},
	}

	ctx := context.Background()
	m.Send(ctx, alert)
	m.Send(ctx, alert) // duplicate within window

	if got := rec.count(); got != 1 {
		t.Errorf("expected 1 alert sent, got %d", got)
	}
}

func TestManager_DedupAllowsDistinct(t *testing.T) {
	rec := &recordingAlerter{}
	m := NewManager(testLogger(), 5*time.Minute)
	defer m.Stop()
	m.Register(rec, "info")

	a1 := Alert{ID: "a1", Severity: "high", Module: "nextjs", SourceIP: "1.2.3.4", Signatures: []string{"sig-001"}}
	a2 := Alert{ID: "a2", Severity: "high", Module: "nextjs", SourceIP: "5.6.7.8", Signatures: []string{"sig-001"}}

	ctx := context.Background()
	m.Send(ctx, a1)
	m.Send(ctx, a2)

	if got := rec.count(); got != 2 {
		t.Errorf("expected 2 alerts sent (distinct IPs), got %d", got)
	}
}

func TestManager_DedupDisabled(t *testing.T) {
	rec := &recordingAlerter{}
	m := NewManager(testLogger(), 0) // dedup disabled
	m.Register(rec, "info")

	alert := Alert{
		ID:         "a1",
		Severity:   "high",
		Module:     "nextjs",
		SourceIP:   "1.2.3.4",
		Signatures: []string{"sig-001"},
	}

	ctx := context.Background()
	m.Send(ctx, alert)
	m.Send(ctx, alert)

	if got := rec.count(); got != 2 {
		t.Errorf("expected 2 alerts (dedup disabled), got %d", got)
	}
}

func TestManager_DedupExpires(t *testing.T) {
	rec := &recordingAlerter{}
	m := NewManager(testLogger(), 10*time.Millisecond)
	defer m.Stop()
	m.Register(rec, "info")

	alert := Alert{
		ID:         "a1",
		Severity:   "high",
		Module:     "nextjs",
		SourceIP:   "1.2.3.4",
		Signatures: []string{"sig-001"},
	}

	ctx := context.Background()
	m.Send(ctx, alert)
	time.Sleep(20 * time.Millisecond)
	m.Send(ctx, alert) // should fire again after expiry

	if got := rec.count(); got != 2 {
		t.Errorf("expected 2 alerts (after expiry), got %d", got)
	}
}
