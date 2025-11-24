package dto

import (
	"time"

	"github.com/0xsj/hexagonal-go/internal/notifications/domain"
)

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

// NewNotificationDTO creates a DTO from a notification.
func NewNotificationDTO(n *domain.Notification) *NotificationDTO {
	return &NotificationDTO{
		ID:            n.ID().String(),
		TenantID:      n.TenantID(),
		Channel:       n.Channel().String(),
		Recipient:     n.Recipient(),
		Subject:       n.Subject(),
		Status:        n.Status().String(),
		Attempts:      n.Attempts(),
		LastError:     n.LastError(),
		SentAt:        n.SentAt(),
		UserID:        n.UserID(),
		CorrelationID: n.CorrelationID(),
		EventType:     n.EventType(),
		CreatedAt:     n.CreatedAt(),
		UpdatedAt:     n.UpdatedAt(),
	}
}
