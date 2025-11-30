package command

import (
	"context"

	"github.com/0xsj/hexagonal-go/internal/permissions/domain"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// RevokeRoleRequest represents the input for revoking a role.
type RevokeRoleRequest struct {
	UserID     types.ID `json:"user_id"`
	RoleID     types.ID `json:"role_id"`
	TenantID   string   `json:"tenant_id"`
	ResourceID string   `json:"resource_id,omitempty"`
	RevokedBy  string   `json:"revoked_by"`
	Reason     string   `json:"reason,omitempty"`
}

// RevokeRoleResponse represents the output of revoking a role.
type RevokeRoleResponse struct {
	Success bool `json:"success"`
}

// RevokeRoleCommand handles role revocation.
type RevokeRoleCommand struct {
	roleRepo       domain.RoleRepository
	assignmentRepo domain.AssignmentRepository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewRevokeRoleCommand creates a new RevokeRoleCommand.
func NewRevokeRoleCommand(
	roleRepo domain.RoleRepository,
	assignmentRepo domain.AssignmentRepository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *RevokeRoleCommand {
	return &RevokeRoleCommand{
		roleRepo:       roleRepo,
		assignmentRepo: assignmentRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// Handle executes the revoke role command.
func (c *RevokeRoleCommand) Handle(ctx context.Context, req RevokeRoleRequest) (*RevokeRoleResponse, error) {
	const op = "RevokeRoleCommand.Handle"

	// Find assignment
	assignment, err := c.assignmentRepo.FindByUserAndRole(ctx, req.UserID, req.RoleID, req.TenantID, req.ResourceID)
	if err != nil {
		return nil, err
	}

	// Get role name for event
	role, err := c.roleRepo.FindByID(ctx, req.RoleID)
	if err != nil {
		c.logger.Warn("failed to get role for revocation event",
			logger.String("role_id", req.RoleID.String()),
			logger.Err(err),
		)
	}

	roleName := ""
	if role != nil {
		roleName = role.Name()
	}

	// Mark as revoked (emits event)
	assignment.Revoke(roleName, req.RevokedBy, req.Reason)

	// Delete assignment
	if err := c.assignmentRepo.Delete(ctx, assignment.ID()); err != nil {
		c.logger.Error("failed to delete assignment", logger.Err(err))
		return nil, err
	}

	// Publish domain events
	events := messaging.AsDomainEvents(assignment.Events())
	defer assignment.ClearEvents()

	if err := c.eventPublisher.PublishAll(ctx, "permissions", "assignment", events); err != nil {
		c.logger.Error("failed to publish events",
			logger.String("assignment_id", assignment.ID().String()),
			logger.Err(err),
		)
	}

	c.logger.Info("role revoked",
		logger.String("assignment_id", assignment.ID().String()),
		logger.String("user_id", req.UserID.String()),
		logger.String("role_id", req.RoleID.String()),
		logger.String("role_name", roleName),
		logger.String("revoked_by", req.RevokedBy),
	)

	return &RevokeRoleResponse{
		Success: true,
	}, nil
}
