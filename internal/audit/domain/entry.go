package domain

import (
	"time"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// AuditEntry represents a record of an event that occurred in the system.
// Immutable once created - audit entries are never modified.
type AuditEntry struct {
	id        types.ID
	eventID   string
	eventType string
	source    string
	timestamp time.Time

	// Context
	tenantID      string
	userID        string
	correlationID string

	// Event data
	payload  map[string]any
	metadata map[string]any

	// Record keeping
	createdAt time.Time
}

// NewAuditEntry creates a new audit entry from an event.
func NewAuditEntry(
	eventID string,
	eventType string,
	source string,
	timestamp time.Time,
	tenantID string,
	userID string,
	correlationID string,
	payload map[string]any,
	metadata map[string]any,
) *AuditEntry {
	return &AuditEntry{
		id:            types.NewID(),
		eventID:       eventID,
		eventType:     eventType,
		source:        source,
		timestamp:     timestamp,
		tenantID:      tenantID,
		userID:        userID,
		correlationID: correlationID,
		payload:       payload,
		metadata:      metadata,
		createdAt:     time.Now().UTC(),
	}
}

// Getters - AuditEntry is immutable

func (e *AuditEntry) ID() types.ID             { return e.id }
func (e *AuditEntry) EventID() string          { return e.eventID }
func (e *AuditEntry) EventType() string        { return e.eventType }
func (e *AuditEntry) Source() string           { return e.source }
func (e *AuditEntry) Timestamp() time.Time     { return e.timestamp }
func (e *AuditEntry) TenantID() string         { return e.tenantID }
func (e *AuditEntry) UserID() string           { return e.userID }
func (e *AuditEntry) CorrelationID() string    { return e.correlationID }
func (e *AuditEntry) Payload() map[string]any  { return e.payload }
func (e *AuditEntry) Metadata() map[string]any { return e.metadata }
func (e *AuditEntry) CreatedAt() time.Time     { return e.createdAt }

// Snapshot is used for persistence/reconstitution.
type Snapshot struct {
	ID            string
	EventID       string
	EventType     string
	Source        string
	Timestamp     time.Time
	TenantID      string
	UserID        string
	CorrelationID string
	Payload       map[string]any
	Metadata      map[string]any
	CreatedAt     time.Time
}

// ToSnapshot converts the entry to a snapshot for persistence.
func (e *AuditEntry) ToSnapshot() Snapshot {
	return Snapshot{
		ID:            e.id.String(),
		EventID:       e.eventID,
		EventType:     e.eventType,
		Source:        e.source,
		Timestamp:     e.timestamp,
		TenantID:      e.tenantID,
		UserID:        e.userID,
		CorrelationID: e.correlationID,
		Payload:       e.payload,
		Metadata:      e.metadata,
		CreatedAt:     e.createdAt,
	}
}

// FromSnapshot reconstitutes an entry from a snapshot.
func FromSnapshot(s Snapshot) (*AuditEntry, error) {
	id, err := types.ParseID(s.ID)
	if err != nil {
		return nil, err
	}

	return &AuditEntry{
		id:            id,
		eventID:       s.EventID,
		eventType:     s.EventType,
		source:        s.Source,
		timestamp:     s.Timestamp,
		tenantID:      s.TenantID,
		userID:        s.UserID,
		correlationID: s.CorrelationID,
		payload:       s.Payload,
		metadata:      s.Metadata,
		createdAt:     s.CreatedAt,
	}, nil
}
