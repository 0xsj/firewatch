package command

import (
	"context"

	"github.com/0xsj/hexagonal-go/internal/permissions/application/dto"
	"github.com/0xsj/hexagonal-go/internal/permissions/domain"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// CreateRoleRequest represents the input for creating a role.
type CreateRoleRequest struct {
	TenantID    string   `json:"tenant_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	CreatedBy   string   `json:"created_by"`
}

// CreateRoleResponse represents the output of creating a role.
type CreateRoleResponse struct {
	Role dto.RoleDTO `json:"role"`
}

// CreateRoleCommand handles role creation.
type CreateRoleCommand struct {
	roleRepo       domain.RoleRepository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewCreateRoleCommand creates a new CreateRoleCommand.
func NewCreateRoleCommand(
	roleRepo domain.RoleRepository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *CreateRoleCommand {
	return &CreateRoleCommand{
		roleRepo:       roleRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// Handle executes the create role command.
func (c *CreateRoleCommand) Handle(ctx context.Context, req CreateRoleRequest) (*CreateRoleResponse, error) {
	const op = "CreateRoleCommand.Handle"

	// Check if role name already exists
	exists, err := c.roleRepo.Exists(ctx, req.TenantID, req.Name)
	if err != nil {
		c.logger.Error("failed to check role existence", logger.Err(err))
		return nil, err
	}

	if exists {
		return nil, domain.ErrRoleAlreadyExists(op, req.Name)
	}

	// Parse permissions
	permissions, err := domain.ParsePermissionSet(req.Permissions)
	if err != nil {
		return nil, err
	}

	// Generate ID
	id := types.NewID()

	// Create role
	role, err := domain.CreateRole(
		id,
		req.TenantID,
		req.Name,
		req.Description,
		permissions,
		req.CreatedBy,
	)
	if err != nil {
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

	c.logger.Info("role created",
		logger.String("role_id", role.ID().String()),
		logger.String("role_name", role.Name()),
		logger.String("created_by", req.CreatedBy),
	)

	return &CreateRoleResponse{
		Role: dto.RoleToDTO(role),
	}, nil
}
