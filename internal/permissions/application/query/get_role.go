package query

import (
	"context"

	"github.com/0xsj/hexagonal-go/internal/permissions/application/dto"
	"github.com/0xsj/hexagonal-go/internal/permissions/domain"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// GetRoleByIDRequest represents the input for getting a role by ID.
type GetRoleByIDRequest struct {
	ID types.ID `json:"id"`
}

// GetRoleByNameRequest represents the input for getting a role by name.
type GetRoleByNameRequest struct {
	TenantID string `json:"tenant_id"`
	Name     string `json:"name"`
}

// GetRoleResponse represents the output of getting a role.
type GetRoleResponse struct {
	Role dto.RoleDTO `json:"role"`
}

// GetRoleQuery handles role retrieval.
type GetRoleQuery struct {
	roleRepo domain.RoleRepository
	logger   logger.Logger
}

// NewGetRoleQuery creates a new GetRoleQuery.
func NewGetRoleQuery(
	roleRepo domain.RoleRepository,
	logger logger.Logger,
) *GetRoleQuery {
	return &GetRoleQuery{
		roleRepo: roleRepo,
		logger:   logger,
	}
}

// HandleByID retrieves a role by ID.
func (q *GetRoleQuery) HandleByID(ctx context.Context, req GetRoleByIDRequest) (*GetRoleResponse, error) {
	role, err := q.roleRepo.FindByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	return &GetRoleResponse{
		Role: dto.RoleToDTO(role),
	}, nil
}

// HandleByName retrieves a role by name within a tenant.
func (q *GetRoleQuery) HandleByName(ctx context.Context, req GetRoleByNameRequest) (*GetRoleResponse, error) {
	role, err := q.roleRepo.FindByName(ctx, req.TenantID, req.Name)
	if err != nil {
		return nil, err
	}

	return &GetRoleResponse{
		Role: dto.RoleToDTO(role),
	}, nil
}
