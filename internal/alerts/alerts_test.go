package alerts

import (
	"log/slog"
	"os"
	"testing"
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
	m := NewManager(testLogger())
	if m.Count() != 0 {
		t.Errorf("Count() = %d, want 0", m.Count())
	}
}

func TestManager_Register(t *testing.T) {
	m := NewManager(testLogger())
	m.Register(NewSlack("https://hooks.slack.com/test"), "medium")
	m.Register(NewDiscord("https://discord.com/api/webhooks/test"), "high")

	if m.Count() != 2 {
		t.Errorf("Count() = %d, want 2", m.Count())
	}
}
