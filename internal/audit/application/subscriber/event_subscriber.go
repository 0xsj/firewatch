package subscriber

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/audit/domain"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// EventSubscriber subscribes to all events and creates audit entries.
// Implements messaging.EventHandler.
type EventSubscriber struct {
	repo   domain.Repository
	logger logger.Logger
}

// NewEventSubscriber creates a new event subscriber.
func NewEventSubscriber(repo domain.Repository, logger logger.Logger) *EventSubscriber {
	return &EventSubscriber{
		repo:   repo,
		logger: logger,
	}
}

// Register registers the subscriber to receive all events.
func (s *EventSubscriber) Register(sub messaging.Subscriber) error {
	return sub.SubscribeAll(s)
}

// Handle processes any event and creates an audit entry.
// Implements messaging.EventHandler.
func (s *EventSubscriber) Handle(ctx context.Context, event messaging.Event) error {
	const op = "audit.EventSubscriber.Handle"

	s.logger.Debug("received event for audit",
		logger.String("event_id", event.ID()),
		logger.String("event_type", event.Type()),
		logger.String("source", event.Source()),
	)

	// Extract context from event metadata
	tenantID := messaging.GetTenantID(event)
	userID := messaging.GetUserID(event)
	correlationID := messaging.GetCorrelationID(event)

	// Create audit entry
	entry := domain.NewAuditEntry(
		event.ID(),
		event.Type(),
		event.Source(),
		event.Timestamp(),
		tenantID,
		userID,
		correlationID,
		event.Data(),
		event.Metadata(),
	)

	// Use background context for persistence to avoid HTTP request context cancellation.
	// The event has already been received, so we want to persist it regardless of
	// the original request's lifecycle.
	saveCtx := context.Background()

	// Save to repository
	if err := s.repo.Save(saveCtx, entry); err != nil {
		s.logger.Error("failed to save audit entry",
			logger.String("event_id", event.ID()),
			logger.String("event_type", event.Type()),
			logger.Err(err),
		)
		return fmt.Errorf("%s: failed to save audit entry: %w", op, err)
	}

	s.logger.Debug("audit entry created",
		logger.String("audit_id", entry.ID().String()),
		logger.String("event_type", event.Type()),
	)

	return nil
}
