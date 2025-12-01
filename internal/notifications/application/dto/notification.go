// internal/notifications/application/dto/notification.go
package dto

import "time"

// SendEmailRequest is the input for sending an email notification.
type SendEmailRequest struct {
	TenantID      string
	Recipient     string
	Subject       string
	TextBody      string
	HTMLBody      string
	UserID        string
	CorrelationID string
	EventType     string
}

// NotificationDTO is the response for a notification.
type NotificationDTO struct {
	ID            string     `json:"id"`
	TenantID      string     `json:"tenant_id,omitempty"`
	Channel       string     `json:"channel"`
	Recipient     string     `json:"recipient"`
	Subject       string     `json:"subject"`
	Status        string     `json:"status"`
	Attempts      int        `json:"attempts"`
	LastError     string     `json:"last_error,omitempty"`
	SentAt        *time.Time `json:"sent_at,omitempty"`
	UserID        string     `json:"user_id,omitempty"`
	CorrelationID string     `json:"correlation_id,omitempty"`
	EventType     string     `json:"event_type,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
