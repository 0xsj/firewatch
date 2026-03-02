package storage

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/0xsj/firewatch/internal/alerts"
	"github.com/0xsj/firewatch/internal/storage/models"
)

func TestAlertingStore_SaveEvent_DispatchesAlert(t *testing.T) {
	inner := newTestStore(t)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	alertMgr := alerts.NewManager(logger, 0)
	// No alerters registered — just verify it doesn't panic and the
	// event is still saved.
	as := NewAlertingStore(inner, alertMgr)

	ctx := context.Background()
	event := &models.Event{
		ID:         "evt-alert-001",
		Timestamp:  "2024-01-15T12:00:00Z",
		RequestID:  "req-001",
		SourceIP:   "203.0.113.50",
		Module:     "nextjs",
		Method:     "POST",
		Path:       "/",
		Severity:   "high",
		Signatures: []string{"nextjs-action-001"},
	}

	if err := as.SaveEvent(ctx, event); err != nil {
		t.Fatalf("SaveEvent: %v", err)
	}

	// Verify event was actually stored.
	got, err := inner.GetEvent(ctx, "evt-alert-001")
	if err != nil {
		t.Fatalf("GetEvent: %v", err)
	}
	if got.ID != "evt-alert-001" {
		t.Errorf("ID = %q, want evt-alert-001", got.ID)
	}
}

func TestAlertingStore_BuildTitle(t *testing.T) {
	tests := []struct {
		event *models.Event
		want  string
	}{
		{
			event: &models.Event{Module: "nextjs", Signatures: []string{"nextjs-action-001"}},
			want:  "[nextjs] nextjs-action-001",
		},
		{
			event: &models.Event{Module: "wordpress", Method: "GET", Path: "/wp-login.php"},
			want:  "[wordpress] GET /wp-login.php",
		},
	}

	for _, tt := range tests {
		got := buildTitle(tt.event)
		if got != tt.want {
			t.Errorf("buildTitle() = %q, want %q", got, tt.want)
		}
	}
}

func TestAlertingStore_SkipsWhenNoAlerters(t *testing.T) {
	inner := newTestStore(t)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	alertMgr := alerts.NewManager(logger, 0)

	// Count() == 0, so the fast path should be taken.
	as := NewAlertingStore(inner, alertMgr)

	ctx := context.Background()
	start := time.Now()
	err := as.SaveEvent(ctx, &models.Event{
		ID:        "evt-fast",
		Timestamp: "2024-01-15T12:00:00Z",
		RequestID: "r1",
		SourceIP:  "10.0.0.1",
		Module:    "test",
		Method:    "GET",
		Path:      "/",
		Severity:  "info",
	})
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("SaveEvent: %v", err)
	}
	// Should be fast (no HTTP calls).
	if elapsed > 100*time.Millisecond {
		t.Errorf("SaveEvent took %v, expected fast path", elapsed)
	}
}
