package jobs

// Job type constants for notification jobs.
const (
	// JobTypeSendEmail sends an email notification.
	JobTypeSendEmail = "notifications.send_email"
)

// SendEmailPayload is the payload for send_email jobs.
type SendEmailPayload struct {
	TenantID      string         `json:"tenant_id,omitempty"`
	Recipient     string         `json:"recipient"`
	Subject       string         `json:"subject"`
	HTMLBody      string         `json:"html_body"`
	TextBody      string         `json:"text_body,omitempty"`
	UserID        string         `json:"user_id,omitempty"`
	CorrelationID string         `json:"correlation_id,omitempty"`
	EventType     string         `json:"event_type,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
}
