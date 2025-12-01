// internal/audit/application/dto/requests.go
package dto

import (
	"time"
)

// GetEntryRequest represents the input for getting a single audit entry.
type GetEntryRequest struct {
	ID string `json:"id" validate:"required"`
}

// ListEntriesRequest represents a query to list audit entries with filters.
type ListEntriesRequest struct {
	// Filters
	TenantID      string     `json:"tenant_id,omitempty"`
	UserID        string     `json:"user_id,omitempty"`
	EventType     string     `json:"event_type,omitempty"`
	Source        string     `json:"source,omitempty"`
	CorrelationID string     `json:"correlation_id,omitempty"`
	FromTimestamp *time.Time `json:"from_timestamp,omitempty"`
	ToTimestamp   *time.Time `json:"to_timestamp,omitempty"`

	// Pagination
	Limit  int `json:"limit" validate:"min=1,max=100"`
	Offset int `json:"offset" validate:"min=0"`
}

// GetResourceTrailRequest represents the input for getting a resource's audit trail.
type GetResourceTrailRequest struct {
	// Resource identification (derived from event payload)
	ResourceType string `json:"resource_type" validate:"required"` // e.g., "user", "tenant", "flag"
	ResourceID   string `json:"resource_id" validate:"required"`

	// Optional filters
	TenantID      string     `json:"tenant_id,omitempty"`
	FromTimestamp *time.Time `json:"from_timestamp,omitempty"`
	ToTimestamp   *time.Time `json:"to_timestamp,omitempty"`

	// Pagination
	Limit  int `json:"limit" validate:"min=1,max=100"`
	Offset int `json:"offset" validate:"min=0"`
}

// GetActorActivityRequest represents the input for getting an actor's activity.
type GetActorActivityRequest struct {
	// Actor identification
	UserID string `json:"user_id" validate:"required"`

	// Optional filters
	TenantID      string     `json:"tenant_id,omitempty"`
	EventType     string     `json:"event_type,omitempty"`
	Source        string     `json:"source,omitempty"`
	FromTimestamp *time.Time `json:"from_timestamp,omitempty"`
	ToTimestamp   *time.Time `json:"to_timestamp,omitempty"`

	// Pagination
	Limit  int `json:"limit" validate:"min=1,max=100"`
	Offset int `json:"offset" validate:"min=0"`
}
