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

// SuspendTenantCommand handles tenant suspension.
type SuspendTenantCommand struct {
	repo           tenant.Repository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewSuspendTenantCommand creates a new SuspendTenantCommand.
func NewSuspendTenantCommand(
	repo tenant.Repository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *SuspendTenantCommand {
	return &SuspendTenantCommand{
		repo:           repo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// SuspendTenantRequest is the input for tenant suspension.
type SuspendTenantRequest struct {
	TenantID    string
	Reason      string
	SuspendedBy string
}

// Handle executes the suspend tenant command.
func (c *SuspendTenantCommand) Handle(ctx context.Context, req SuspendTenantRequest) (*dto.TenantDTO, error) {
	const op = "SuspendTenantCommand.Handle"

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

	// Suspend tenant (domain logic handles version increment)
	if err := t.Suspend(req.Reason, req.SuspendedBy); err != nil {
		return nil, err
	}

	// Save to repository
	if err := c.repo.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("%s: failed to update tenant: %w", op, err)
	}

	c.logger.Info("tenant suspended",
		logger.String("tenant_id", t.ID().String()),
		logger.String("reason", req.Reason),
		logger.String("suspended_by", req.SuspendedBy),
	)

	// Publish domain events using shared publisher
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
