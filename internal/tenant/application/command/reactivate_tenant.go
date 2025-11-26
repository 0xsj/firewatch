package command

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/tenant/application/dto"
	tenant "github.com/0xsj/hexagonal-go/internal/tenant/domain"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// ReactivateTenantCommand handles tenant reactivation.
type ReactivateTenantCommand struct {
	repo      tenant.Repository
	publisher messaging.Publisher
	logger    logger.Logger
}

// NewReactivateTenantCommand creates a new ReactivateTenantCommand.
func NewReactivateTenantCommand(
	repo tenant.Repository,
	publisher messaging.Publisher,
	logger logger.Logger,
) *ReactivateTenantCommand {
	return &ReactivateTenantCommand{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
	}
}

// ReactivateTenantRequest is the input for tenant reactivation.
type ReactivateTenantRequest struct {
	TenantID      string
	ReactivatedBy string
}

// Handle executes the reactivate tenant command.
func (c *ReactivateTenantCommand) Handle(ctx context.Context, req ReactivateTenantRequest) (*dto.TenantDTO, error) {
	const op = "ReactivateTenantCommand.Handle"

	// Parse tenant ID
	tenantID, err := types.ParseID(req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid tenant_id: %w", op, err)
	}

	// Fetch tenant
	t, err := c.repo.FindByID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to find tenant: %w", op, err)
	}
	if t == nil {
		return nil, tenant.ErrTenantNotFound(op, req.TenantID)
	}

	// Reactivate tenant
	if err := t.Reactivate(req.ReactivatedBy); err != nil {
		return nil, err
	}

	// Increment version for optimistic locking
	t.IncrementVersion()

	// Save to repository
	if err := c.repo.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("%s: failed to update tenant: %w", op, err)
	}

	c.logger.Info("tenant reactivated",
		logger.String("tenant_id", t.ID().String()),
		logger.String("reactivated_by", req.ReactivatedBy),
	)

	// Publish domain events
	if err := c.publishEvents(ctx, t); err != nil {
		c.logger.Error("failed to publish events",
			logger.String("tenant_id", t.ID().String()),
			logger.Err(err),
		)
	}

	return dto.ToTenantDTO(t), nil
}

// publishEvents publishes all domain events from the aggregate.
func (c *ReactivateTenantCommand) publishEvents(ctx context.Context, t *tenant.Tenant) error {
	events := t.Events()
	defer t.ClearEvents()

	for _, domainEvent := range events {
		event := c.convertDomainEvent(domainEvent)

		if err := c.publisher.Publish(ctx, event); err != nil {
			return fmt.Errorf("failed to publish event %s: %w", domainEvent.Type(), err)
		}

		c.logger.Debug("event published",
			logger.String("event_type", event.Type()),
			logger.String("event_id", event.ID()),
		)
	}

	return nil
}

// convertDomainEvent converts a domain event to a messaging event.
func (c *ReactivateTenantCommand) convertDomainEvent(domainEvent tenant.Event) *messaging.BaseEvent {
	event := messaging.NewEvent(
		"tenant."+domainEvent.Type(),
		"tenant",
		domainEvent.Payload(),
	)

	event.WithMetadata("aggregate_id", domainEvent.AggregateID().String())
	event.WithMetadata("aggregate_type", "tenant")

	return event
}
