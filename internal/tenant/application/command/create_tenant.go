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

// CreateTenantCommand handles tenant creation.
type CreateTenantCommand struct {
	repo      tenant.Repository
	publisher messaging.Publisher
	logger    logger.Logger
}

// NewCreateTenantCommand creates a new CreateTenantCommand.
func NewCreateTenantCommand(
	repo tenant.Repository,
	publisher messaging.Publisher,
	logger logger.Logger,
) *CreateTenantCommand {
	return &CreateTenantCommand{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
	}
}

// CreateTenantRequest is the input for tenant creation.
type CreateTenantRequest struct {
	Slug      string
	Name      string
	Plan      string
	OwnerID   string
	CreatedBy string
}

// Handle executes the create tenant command.
func (c *CreateTenantCommand) Handle(ctx context.Context, req CreateTenantRequest) (*dto.TenantDTO, error) {
	const op = "CreateTenantCommand.Handle"

	// Parse and validate slug
	slug, err := tenant.NewSlug(req.Slug)
	if err != nil {
		return nil, err
	}

	// Check if slug already exists
	exists, err := c.repo.SlugExists(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to check slug: %w", op, err)
	}
	if exists {
		return nil, tenant.ErrSlugAlreadyTaken(op, req.Slug)
	}

	// Parse plan (default to free)
	plan := tenant.PlanFree
	if req.Plan != "" {
		plan, err = tenant.ParsePlan(req.Plan)
		if err != nil {
			return nil, tenant.ErrPlanInvalid(op, req.Plan)
		}
	}

	// Parse owner ID
	ownerID, err := types.ParseID(req.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid owner_id: %w", op, err)
	}

	// Create tenant
	tenantID := types.NewID()
	t, err := tenant.Create(
		tenantID,
		slug,
		req.Name,
		plan,
		ownerID,
		req.CreatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Save to repository
	if err := c.repo.Save(ctx, t); err != nil {
		return nil, fmt.Errorf("%s: failed to save tenant: %w", op, err)
	}

	c.logger.Info("tenant created",
		logger.String("tenant_id", t.ID().String()),
		logger.String("slug", t.Slug().String()),
		logger.String("owner_id", ownerID.String()),
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
func (c *CreateTenantCommand) publishEvents(ctx context.Context, t *tenant.Tenant) error {
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
func (c *CreateTenantCommand) convertDomainEvent(domainEvent tenant.Event) *messaging.BaseEvent {
	event := messaging.NewEvent(
		"tenant."+domainEvent.Type(),
		"tenant",
		domainEvent.Payload(),
	)

	event.WithMetadata("aggregate_id", domainEvent.AggregateID().String())
	event.WithMetadata("aggregate_type", "tenant")

	return event
}
