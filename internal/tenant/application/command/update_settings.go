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

// UpdateSettingsCommand handles tenant settings updates.
type UpdateSettingsCommand struct {
	repo      tenant.Repository
	publisher messaging.Publisher
	logger    logger.Logger
}

// NewUpdateSettingsCommand creates a new UpdateSettingsCommand.
func NewUpdateSettingsCommand(
	repo tenant.Repository,
	publisher messaging.Publisher,
	logger logger.Logger,
) *UpdateSettingsCommand {
	return &UpdateSettingsCommand{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
	}
}

// UpdateSettingsRequest is the input for settings update.
type UpdateSettingsRequest struct {
	TenantID  string
	Settings  map[string]any
	UpdatedBy string
}

// Handle executes the update settings command.
func (c *UpdateSettingsCommand) Handle(ctx context.Context, req UpdateSettingsRequest) (*dto.TenantSettingsDTO, error) {
	const op = "UpdateSettingsCommand.Handle"

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

	// Create new settings from request
	newSettings := tenant.NewSettingsFromMap(req.Settings)

	// Apply settings update
	if err := t.UpdateSettings(newSettings, req.UpdatedBy); err != nil {
		return nil, err
	}

	// Check if there are changes
	if len(t.Events()) == 0 {
		return dto.ToSettingsDTO(t), nil
	}

	// Increment version for optimistic locking
	t.IncrementVersion()

	// Save to repository
	if err := c.repo.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("%s: failed to update tenant: %w", op, err)
	}

	c.logger.Info("tenant settings updated",
		logger.String("tenant_id", t.ID().String()),
		logger.String("updated_by", req.UpdatedBy),
	)

	// Publish domain events
	if err := c.publishEvents(ctx, t); err != nil {
		c.logger.Error("failed to publish events",
			logger.String("tenant_id", t.ID().String()),
			logger.Err(err),
		)
	}

	return dto.ToSettingsDTO(t), nil
}

// publishEvents publishes all domain events from the aggregate.
func (c *UpdateSettingsCommand) publishEvents(ctx context.Context, t *tenant.Tenant) error {
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
func (c *UpdateSettingsCommand) convertDomainEvent(domainEvent tenant.Event) *messaging.BaseEvent {
	event := messaging.NewEvent(
		"tenant."+domainEvent.Type(),
		"tenant",
		domainEvent.Payload(),
	)

	event.WithMetadata("aggregate_id", domainEvent.AggregateID().String())
	event.WithMetadata("aggregate_type", "tenant")

	return event
}
