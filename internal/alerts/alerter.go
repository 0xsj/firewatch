package alerts

import "context"

// Alert represents a notification triggered by a detection event.
type Alert struct {
	ID         string   `json:"id"`
	Timestamp  string   `json:"timestamp"`
	Severity   string   `json:"severity"`
	Module     string   `json:"module"`
	Title      string   `json:"title"`
	Message    string   `json:"message"`
	SourceIP   string   `json:"source_ip"`
	Path       string   `json:"path"`
	Method     string   `json:"method"`
	UserAgent  string   `json:"user_agent,omitempty"`
	Signatures []string `json:"signatures,omitempty"`
	RequestID  string   `json:"request_id,omitempty"`
}

// Alerter is the interface for sending alert notifications.
// Implementations handle the transport (Slack, Discord, webhook, etc.).
type Alerter interface {
	// Name returns the alerter identifier (e.g., "slack", "discord").
	Name() string

	// Send dispatches an alert. Implementations should be safe for
	// concurrent use and should not block indefinitely.
	Send(ctx context.Context, alert Alert) error
}

// Severity ordering for filtering.
var severityRank = map[string]int{
	"info":     0,
	"low":      1,
	"medium":   2,
	"high":     3,
	"critical": 4,
}

// MeetsSeverity reports whether the alert's severity meets or
// exceeds the minimum threshold.
func MeetsSeverity(alertSeverity, minSeverity string) bool {
	return severityRank[alertSeverity] >= severityRank[minSeverity]
}
