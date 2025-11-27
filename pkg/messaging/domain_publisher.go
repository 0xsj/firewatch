package messaging

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// DomainEventPublisher handles the conversion and publishing of domain events
// to the messaging infrastructure. It encapsulates the boilerplate that was
// previously duplicated across all command handlers.
type DomainEventPublisher struct {
	publisher Publisher
	logger    logger.Logger
}

// NewDomainEventPublisher creates a new DomainEventPublisher.
func NewDomainEventPublisher(publisher Publisher, logger logger.Logger) *DomainEventPublisher {
	return &DomainEventPublisher{
		publisher: publisher,
		logger:    logger,
	}
}

// PublishAll publishes all domain events with proper metadata.
//
// Parameters:
//   - ctx: Context for correlation ID, tracing, etc.
//   - domainName: The domain prefix for event types (e.g., "identity", "tenant", "email")
//   - aggregateType: The type of aggregate (e.g., "user", "tenant", "template")
//   - events: The domain events to publish
//
// Usage:
//
//	events := messaging.AsDomainEvents(user.Events())
//	defer user.ClearEvents()
//	if err := publisher.PublishAll(ctx, "identity", "user", events); err != nil {
//	    log.Error("failed to publish events", logger.Err(err))
//	}
func (p *DomainEventPublisher) PublishAll(
	ctx context.Context,
	domainName string,
	aggregateType string,
	events []DomainEvent,
) error {
	for _, domainEvent := range events {
		if err := p.publishOne(ctx, domainName, aggregateType, domainEvent); err != nil {
			return err
		}
	}
	return nil
}

// publishOne converts and publishes a single domain event.
func (p *DomainEventPublisher) publishOne(
	ctx context.Context,
	domainName string,
	aggregateType string,
	domainEvent DomainEvent,
) error {
	// Build event type with domain prefix: "identity.user.registered"
	eventType := domainName + "." + domainEvent.Type()

	// Create messaging event with context metadata (correlation ID, tracing, etc.)
	event := NewEventFromContext(
		ctx,
		eventType,
		domainName,
		domainEvent.Payload(),
	)

	// Add tenant context if the event is tenant-scoped
	if tenantScoped, ok := domainEvent.(TenantScopedEvent); ok {
		event.WithTenantID(tenantScoped.AggregateTenantID())

		// For user aggregates, also set user ID
		if aggregateType == "user" {
			event.WithUserID(domainEvent.AggregateID().String())
		}
	} else if aggregateType == "tenant" {
		// For tenant aggregates, the aggregate ID IS the tenant ID
		event.WithTenantID(domainEvent.AggregateID().String())
	}

	// Add standard aggregate metadata
	event.WithMetadata("aggregate_id", domainEvent.AggregateID().String())
	event.WithMetadata("aggregate_type", aggregateType)
	event.WithMetadata("event_version", domainEvent.Version())

	// Publish
	if err := p.publisher.Publish(ctx, event); err != nil {
		return fmt.Errorf("failed to publish event %s: %w", eventType, err)
	}

	p.logger.Debug("domain event published",
		logger.String("event_type", event.Type()),
		logger.String("event_id", event.ID()),
		logger.String("aggregate_type", aggregateType),
		logger.String("aggregate_id", domainEvent.AggregateID().String()),
	)

	return nil
}
