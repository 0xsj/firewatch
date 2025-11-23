package messaging

import (
	"encoding/json"
	"time"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Event represents a domain event that has occurred in the system.
// Events are immutable records of something that happened.
//
// Design principles:
//   - Events are facts - they describe what happened (past tense)
//   - Events are immutable - once created, they cannot be changed
//   - Events contain all data needed by subscribers
//   - Events are serializable for storage and transport
//
// Event naming convention: {domain}.{entity}.{action}
// Examples:
//   - "identity.user.registered"
//   - "identity.user.email_verified"
//   - "orders.order.placed"
//   - "payments.payment.succeeded"
type Event interface {
	// ID returns the unique event identifier (ULID).
	ID() string

	// Type returns the event type (e.g., "identity.user.registered").
	Type() string

	// Source returns the domain/service that emitted the event.
	Source() string

	// Data returns the event payload.
	Data() map[string]any

	// Timestamp returns when the event occurred.
	Timestamp() time.Time

	// Metadata returns additional context (correlation ID, tenant ID, etc.).
	Metadata() map[string]any

	// JSON serialization
	MarshalJSON() ([]byte, error)
}

// BaseEvent is the concrete implementation of Event.
// Use NewEvent() to create instances.
type BaseEvent struct {
	id        string
	eventType string
	source    string
	data      map[string]any
	timestamp time.Time
	metadata  map[string]any
}

// NewEvent creates a new event with the given type, source, and data.
//
// Example:
//
//	event := messaging.NewEvent(
//	    "identity.user.registered",
//	    "identity",
//	    map[string]any{
//	        "user_id": "019ab0e9-a81d-7789-a571-bb68bee36595",
//	        "email":   "user@example.com",
//	    },
//	)
func NewEvent(eventType, source string, data map[string]any) *BaseEvent {
	return &BaseEvent{
		id:        types.NewID().String(),
		eventType: eventType,
		source:    source,
		data:      data,
		timestamp: time.Now().UTC(),
		metadata:  make(map[string]any),
	}
}

// NewEventWithID creates an event with a specific ID (for reconstitution).
func NewEventWithID(id, eventType, source string, data map[string]any, timestamp time.Time) *BaseEvent {
	return &BaseEvent{
		id:        id,
		eventType: eventType,
		source:    source,
		data:      data,
		timestamp: timestamp,
		metadata:  make(map[string]any),
	}
}

// ============================================================================
// Event Interface Implementation
// ============================================================================

func (e *BaseEvent) ID() string               { return e.id }
func (e *BaseEvent) Type() string             { return e.eventType }
func (e *BaseEvent) Source() string           { return e.source }
func (e *BaseEvent) Data() map[string]any     { return e.data }
func (e *BaseEvent) Timestamp() time.Time     { return e.timestamp }
func (e *BaseEvent) Metadata() map[string]any { return e.metadata }

// ============================================================================
// Metadata Helpers
// ============================================================================

// WithMetadata adds a metadata key-value pair to the event.
// Returns the event for method chaining.
//
// Example:
//
//	event.WithMetadata("correlation_id", "abc-123")
//	     .WithMetadata("tenant_id", "acme-corp")
func (e *BaseEvent) WithMetadata(key string, value any) *BaseEvent {
	e.metadata[key] = value
	return e
}

// WithCorrelationID sets the correlation ID for request tracing.
func (e *BaseEvent) WithCorrelationID(correlationID string) *BaseEvent {
	return e.WithMetadata("correlation_id", correlationID)
}

// WithCausationID sets the causation ID (the event that caused this event).
func (e *BaseEvent) WithCausationID(causationID string) *BaseEvent {
	return e.WithMetadata("causation_id", causationID)
}

// WithTenantID sets the tenant ID for multi-tenancy.
func (e *BaseEvent) WithTenantID(tenantID string) *BaseEvent {
	return e.WithMetadata("tenant_id", tenantID)
}

// WithUserID sets the user ID who triggered the event.
func (e *BaseEvent) WithUserID(userID string) *BaseEvent {
	return e.WithMetadata("user_id", userID)
}

// WithIPAddress sets the IP address of the originator.
func (e *BaseEvent) WithIPAddress(ip string) *BaseEvent {
	return e.WithMetadata("ip_address", ip)
}

// ============================================================================
// JSON Serialization
// ============================================================================

// EventJSON represents the JSON structure of an event.
type EventJSON struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Source    string         `json:"source"`
	Data      map[string]any `json:"data"`
	Timestamp string         `json:"timestamp"` // RFC3339
	Metadata  map[string]any `json:"metadata"`
}

// MarshalJSON implements json.Marshaler.
func (e *BaseEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(EventJSON{
		ID:        e.id,
		Type:      e.eventType,
		Source:    e.source,
		Data:      e.data,
		Timestamp: e.timestamp.Format(time.RFC3339Nano),
		Metadata:  e.metadata,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (e *BaseEvent) UnmarshalJSON(data []byte) error {
	var eventJSON EventJSON
	if err := json.Unmarshal(data, &eventJSON); err != nil {
		return err
	}

	timestamp, err := time.Parse(time.RFC3339Nano, eventJSON.Timestamp)
	if err != nil {
		return err
	}

	e.id = eventJSON.ID
	e.eventType = eventJSON.Type
	e.source = eventJSON.Source
	e.data = eventJSON.Data
	e.timestamp = timestamp
	e.metadata = eventJSON.Metadata

	return nil
}

// String returns a human-readable representation of the event.
func (e *BaseEvent) String() string {
	return e.eventType + " [" + e.id + "]"
}

// ============================================================================
// Helper Functions
// ============================================================================

// ParseEvent parses a JSON string into an Event.
func ParseEvent(jsonData []byte) (Event, error) {
	var event BaseEvent
	if err := json.Unmarshal(jsonData, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

// MustParseEvent parses a JSON string into an Event or panics.
// Only use for testing or when you're certain the JSON is valid.
func MustParseEvent(jsonData []byte) Event {
	event, err := ParseEvent(jsonData)
	if err != nil {
		panic(err)
	}
	return event
}
