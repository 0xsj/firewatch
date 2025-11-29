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

// UpdateTenantCommand handles tenant updates.
type UpdateTenantCommand struct {
	repo           tenant.Repository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewUpdateTenantCommand creates a new UpdateTenantCommand.
func NewUpdateTenantCommand(
	repo tenant.Repository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *UpdateTenantCommand {
	return &UpdateTenantCommand{
		repo:           repo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// UpdateTenantRequest is the input for tenant update.
type UpdateTenantRequest struct {
	TenantID  string
	Name      *string
	UpdatedBy string
}

// Handle executes the update tenant command.
func (c *UpdateTenantCommand) Handle(ctx context.Context, req UpdateTenantRequest) (*dto.TenantDTO, error) {
	const op = "UpdateTenantCommand.Handle"

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

	// Apply updates
	name := ""
	if req.Name != nil {
		name = *req.Name
	}

	if err := t.Update(name, req.UpdatedBy); err != nil {
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

	c.logger.Info("tenant updated",
		logger.String("tenant_id", t.ID().String()),
		logger.String("updated_by", req.UpdatedBy),
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
