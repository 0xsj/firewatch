package command

import (
	"context"
	"time"

	"github.com/0xsj/hexagonal-go/internal/permissions/application/dto"
	"github.com/0xsj/hexagonal-go/internal/permissions/domain"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// AssignRoleRequest represents the input for assigning a role.
type AssignRoleRequest struct {
	UserID     types.ID   `json:"user_id"`
	RoleID     types.ID   `json:"role_id"`
	TenantID   string     `json:"tenant_id"`
	ResourceID string     `json:"resource_id,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	AssignedBy string     `json:"assigned_by"`
}

// AssignRoleResponse represents the output of assigning a role.
type AssignRoleResponse struct {
	Assignment dto.AssignmentDTO `json:"assignment"`
}

// AssignRoleCommand handles role assignment.
type AssignRoleCommand struct {
	roleRepo       domain.RoleRepository
	assignmentRepo domain.AssignmentRepository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewAssignRoleCommand creates a new AssignRoleCommand.
func NewAssignRoleCommand(
	roleRepo domain.RoleRepository,
	assignmentRepo domain.AssignmentRepository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *AssignRoleCommand {
	return &AssignRoleCommand{
		roleRepo:       roleRepo,
		assignmentRepo: assignmentRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// Handle executes the assign role command.
func (c *AssignRoleCommand) Handle(ctx context.Context, req AssignRoleRequest) (*AssignRoleResponse, error) {
	const op = "AssignRoleCommand.Handle"

	// Verify role exists
	role, err := c.roleRepo.FindByID(ctx, req.RoleID)
	if err != nil {
		return nil, err
	}

	// Check if assignment already exists
	exists, err := c.assignmentRepo.Exists(ctx, req.UserID, req.RoleID, req.TenantID, req.ResourceID)
	if err != nil {
		c.logger.Error("failed to check assignment existence", logger.Err(err))
		return nil, err
	}

	if exists {
		return nil, domain.ErrAssignmentExists(op, req.UserID.String(), req.RoleID.String())
	}

	// Generate ID
	id := types.NewID()

	// Create assignment
	assignment := domain.AssignRole(
		id,
		req.UserID,
		req.RoleID,
		req.TenantID,
		role.Name(),
		req.ResourceID,
		req.ExpiresAt,
		req.AssignedBy,
	)

	// Save assignment
	if err := c.assignmentRepo.Save(ctx, assignment); err != nil {
		c.logger.Error("failed to save assignment", logger.Err(err))
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

	c.logger.Info("role assigned",
		logger.String("assignment_id", assignment.ID().String()),
		logger.String("user_id", req.UserID.String()),
		logger.String("role_id", req.RoleID.String()),
		logger.String("role_name", role.Name()),
		logger.String("assigned_by", req.AssignedBy),
	)

	return &AssignRoleResponse{
		Assignment: dto.AssignmentToDTO(assignment, role.Name()),
	}, nil
}
