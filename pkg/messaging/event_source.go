package messaging

import (
	"time"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// DomainEvent is the minimal interface that all domain events must satisfy.
// Domain event types (user.Event, tenant.Event, etc.) implicitly satisfy this
// through Go's structural typing — no import of this package is required in domains.
type DomainEvent interface {
	// Type returns the event type identifier (e.g., "user.registered", "tenant.created")
	Type() string

	// EventTime returns when the event occurred
	EventTime() time.Time

	// AggregateID returns the ID of the aggregate that produced this event
	AggregateID() types.ID

	// Payload returns the event data as a map for serialization
	Payload() map[string]any

	// Version returns the aggregate version at the time of the event
	Version() int
}

// TenantScopedEvent is implemented by events from aggregates that belong to a tenant.
// User, Session, and EmailTemplate events implement this; Tenant events do not
// (since the tenant aggregate IS the tenant, not owned by one).
type TenantScopedEvent interface {
	DomainEvent

	// AggregateTenantID returns the tenant ID that owns this aggregate
	AggregateTenantID() string
}

// AsDomainEvents converts a slice of domain-specific events to []DomainEvent.
// This is a helper for commands to use when calling DomainEventPublisher.
//
// Usage:
//
//	events := messaging.AsDomainEvents(user.Events())
//	defer user.ClearEvents()
//	publisher.PublishAll(ctx, "identity", "user", events)
func AsDomainEvents[E DomainEvent](events []E) []DomainEvent {
	result := make([]DomainEvent, len(events))
	for i, e := range events {
		result[i] = e
	}
	return result
}
