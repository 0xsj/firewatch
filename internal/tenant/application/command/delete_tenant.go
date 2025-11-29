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

// DeleteTenantCommand handles tenant deletion.
type DeleteTenantCommand struct {
	repo           tenant.Repository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewDeleteTenantCommand creates a new DeleteTenantCommand.
func NewDeleteTenantCommand(
	repo tenant.Repository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *DeleteTenantCommand {
	return &DeleteTenantCommand{
		repo:           repo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// DeleteTenantRequest is the input for tenant deletion.
type DeleteTenantRequest struct {
	TenantID  string
	Reason    string
	DeletedBy string
}

// Handle executes the delete tenant command.
func (c *DeleteTenantCommand) Handle(ctx context.Context, req DeleteTenantRequest) (*dto.TenantDTO, error) {
	const op = "DeleteTenantCommand.Handle"

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

	// Delete tenant (soft delete)
	if err := t.Delete(req.Reason, req.DeletedBy); err != nil {
		return nil, err
	}

	// Increment version for optimistic locking
	t.IncrementVersion()

	// Save to repository
	if err := c.repo.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("%s: failed to update tenant: %w", op, err)
	}

	c.logger.Info("tenant deleted",
		logger.String("tenant_id", t.ID().String()),
		logger.String("reason", req.Reason),
		logger.String("deleted_by", req.DeletedBy),
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
