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

// ChangePlanCommand handles tenant plan changes.
type ChangePlanCommand struct {
	repo      tenant.Repository
	publisher messaging.Publisher
	logger    logger.Logger
}

// NewChangePlanCommand creates a new ChangePlanCommand.
func NewChangePlanCommand(
	repo tenant.Repository,
	publisher messaging.Publisher,
	logger logger.Logger,
) *ChangePlanCommand {
	return &ChangePlanCommand{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
	}
}

// ChangePlanRequest is the input for plan change.
type ChangePlanRequest struct {
	TenantID  string
	Plan      string
	Reason    string
	ChangedBy string
}

// Handle executes the change plan command.
func (c *ChangePlanCommand) Handle(ctx context.Context, req ChangePlanRequest) (*dto.TenantDTO, error) {
	const op = "ChangePlanCommand.Handle"

	// Parse tenant ID
	tenantID, err := types.ParseID(req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid tenant_id: %w", op, err)
	}

	// Parse plan
	newPlan, err := tenant.ParsePlan(req.Plan)
	if err != nil {
		return nil, tenant.ErrPlanInvalid(op, req.Plan)
	}

	// Fetch tenant
	t, err := c.repo.FindByID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to find tenant: %w", op, err)
	}
	if t == nil {
		return nil, tenant.ErrTenantNotFound(op, req.TenantID)
	}

	// Change plan
	if err := t.ChangePlan(newPlan, req.ChangedBy, req.Reason); err != nil {
		return nil, err
	}

	// Check if there are changes
	if len(t.Events()) == 0 {
		return dto.ToTenantDTO(t), nil
	}

	// Increment version for optimistic locking
	t.IncrementVersion()

	// Save to repository
	if err := c.repo.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("%s: failed to update tenant: %w", op, err)
	}

	c.logger.Info("tenant plan changed",
		logger.String("tenant_id", t.ID().String()),
		logger.String("new_plan", newPlan.String()),
		logger.String("changed_by", req.ChangedBy),
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
func (c *ChangePlanCommand) publishEvents(ctx context.Context, t *tenant.Tenant) error {
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
func (c *ChangePlanCommand) convertDomainEvent(domainEvent tenant.Event) *messaging.BaseEvent {
	event := messaging.NewEvent(
		"tenant."+domainEvent.Type(),
		"tenant",
		domainEvent.Payload(),
	)

	event.WithMetadata("aggregate_id", domainEvent.AggregateID().String())
	event.WithMetadata("aggregate_type", "tenant")

	return event
}
