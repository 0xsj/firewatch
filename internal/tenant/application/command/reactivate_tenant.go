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
	repo           tenant.Repository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewReactivateTenantCommand creates a new ReactivateTenantCommand.
func NewReactivateTenantCommand(
	repo tenant.Repository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *ReactivateTenantCommand {
	return &ReactivateTenantCommand{
		repo:           repo,
		eventPublisher: eventPublisher,
		logger:         logger,
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
	events := messaging.AsDomainEvents(t.Events())
	defer t.ClearEvents()

	if err := c.eventPublisher.PublishAll(ctx, "tenant", "tenant", events); err != nil {
		c.logger.Error("failed to publish events",
			logger.String("tenant_id", t.ID().String()),
			logger.Err(err),
		)
	}

	return dto.ToTenantDTO(t), nil
}
