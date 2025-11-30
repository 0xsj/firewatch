package command

import (
	"context"

	"github.com/0xsj/hexagonal-go/internal/permissions/domain"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// DeleteRoleRequest represents the input for deleting a role.
type DeleteRoleRequest struct {
	ID        types.ID `json:"id"`
	DeletedBy string   `json:"deleted_by"`
}

// DeleteRoleResponse represents the output of deleting a role.
type DeleteRoleResponse struct {
	Success bool `json:"success"`
}

// DeleteRoleCommand handles role deletion.
type DeleteRoleCommand struct {
	roleRepo       domain.RoleRepository
	assignmentRepo domain.AssignmentRepository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewDeleteRoleCommand creates a new DeleteRoleCommand.
func NewDeleteRoleCommand(
	roleRepo domain.RoleRepository,
	assignmentRepo domain.AssignmentRepository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *DeleteRoleCommand {
	return &DeleteRoleCommand{
		roleRepo:       roleRepo,
		assignmentRepo: assignmentRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// Handle executes the delete role command.
func (c *DeleteRoleCommand) Handle(ctx context.Context, req DeleteRoleRequest) (*DeleteRoleResponse, error) {
	const op = "DeleteRoleCommand.Handle"

	// Find role
	role, err := c.roleRepo.FindByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// Mark for deletion (validates not a system role)
	if err := role.MarkDeleted(req.DeletedBy); err != nil {
		return nil, err
	}

	// Delete all assignments for this role
	if err := c.assignmentRepo.DeleteByRole(ctx, req.ID); err != nil {
		c.logger.Error("failed to delete role assignments", logger.Err(err))
		return nil, err
	}

	// Delete role
	if err := c.roleRepo.Delete(ctx, req.ID); err != nil {
		c.logger.Error("failed to delete role", logger.Err(err))
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

	c.logger.Info("role deleted",
		logger.String("role_id", role.ID().String()),
		logger.String("role_name", role.Name()),
		logger.String("deleted_by", req.DeletedBy),
	)

	return &DeleteRoleResponse{
		Success: true,
	}, nil
}
