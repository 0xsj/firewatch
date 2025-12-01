// internal/audit/application/dto/audit_entry.go
package dto

import (
	"time"
)

// AuditEntryDTO represents a full audit entry for API responses.
type AuditEntryDTO struct {
	ID            string         `json:"id"`
	EventID       string         `json:"event_id"`
	EventType     string         `json:"event_type"`
	Source        string         `json:"source"`
	Timestamp     time.Time      `json:"timestamp"`
	TenantID      string         `json:"tenant_id,omitempty"`
	UserID        string         `json:"user_id,omitempty"`
	CorrelationID string         `json:"correlation_id,omitempty"`
	Payload       map[string]any `json:"payload,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
}

// AuditEntrySummaryDTO represents a condensed audit entry for list views.
type AuditEntrySummaryDTO struct {
	ID            string    `json:"id"`
	EventType     string    `json:"event_type"`
	Source        string    `json:"source"`
	Timestamp     time.Time `json:"timestamp"`
	TenantID      string    `json:"tenant_id,omitempty"`
	UserID        string    `json:"user_id,omitempty"`
	CorrelationID string    `json:"correlation_id,omitempty"`
}
