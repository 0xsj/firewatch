package command

import (
	"context"

	"github.com/0xsj/hexagonal-go/internal/permissions/application/dto"
	"github.com/0xsj/hexagonal-go/internal/permissions/domain"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// UpdateRoleRequest represents the input for updating a role.
type UpdateRoleRequest struct {
	ID          types.ID `json:"id"`
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	UpdatedBy   string   `json:"updated_by"`
}

// UpdateRoleResponse represents the output of updating a role.
type UpdateRoleResponse struct {
	Role dto.RoleDTO `json:"role"`
}

// UpdateRoleCommand handles role updates.
type UpdateRoleCommand struct {
	roleRepo       domain.RoleRepository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewUpdateRoleCommand creates a new UpdateRoleCommand.
func NewUpdateRoleCommand(
	roleRepo domain.RoleRepository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *UpdateRoleCommand {
	return &UpdateRoleCommand{
		roleRepo:       roleRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// Handle executes the update role command.
func (c *UpdateRoleCommand) Handle(ctx context.Context, req UpdateRoleRequest) (*UpdateRoleResponse, error) {
	const op = "UpdateRoleCommand.Handle"

	// Find role
	role, err := c.roleRepo.FindByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// Prepare update values
	name := ""
	if req.Name != nil {
		name = *req.Name
	}

	description := role.Description()
	if req.Description != nil {
		description = *req.Description
	}

	// Update role
	if err := role.Update(name, description, req.UpdatedBy); err != nil {
		return nil, err
	}

	// Save role
	if err := c.roleRepo.Save(ctx, role); err != nil {
		c.logger.Error("failed to save role", logger.Err(err))
		return nil, err
	}

	// Publish domain events
	events := messaging.AsDomainEvents(role.Events())
	defer role.ClearEvents()

	if err := c.eventPublisher.PublishAll(ctx, "permissions", "role", events); err != nil {
		c.logger.Error("failed to publish events",
			logger.String("role_id", role.ID().String()),
			logger.Err(err),
		)
	}

	c.logger.Info("role updated",
		logger.String("role_id", role.ID().String()),
		logger.String("role_name", role.Name()),
		logger.String("updated_by", req.UpdatedBy),
	)

	return &UpdateRoleResponse{
		Role: dto.RoleToDTO(role),
	}, nil
}
