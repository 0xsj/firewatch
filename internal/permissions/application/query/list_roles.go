package query

import (
	"context"

	"github.com/0xsj/hexagonal-go/internal/permissions/application/dto"
	"github.com/0xsj/hexagonal-go/internal/permissions/domain"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// ListRolesRequest represents the input for listing roles.
type ListRolesRequest struct {
	TenantID      string `json:"tenant_id"`
	IncludeSystem bool   `json:"include_system"`
	Search        string `json:"search"`
	Limit         int    `json:"limit"`
	Offset        int    `json:"offset"`
}

// ListRolesResponse represents the output of listing roles.
type ListRolesResponse struct {
	Roles []dto.RoleSummaryDTO `json:"roles"`
	Total int                  `json:"total"`
}

// ListRolesQuery handles role listing.
type ListRolesQuery struct {
	roleRepo domain.RoleRepository
	logger   logger.Logger
}

// NewListRolesQuery creates a new ListRolesQuery.
func NewListRolesQuery(
	roleRepo domain.RoleRepository,
	logger logger.Logger,
) *ListRolesQuery {
	return &ListRolesQuery{
		roleRepo: roleRepo,
		logger:   logger,
	}
}

// Handle retrieves all roles for a tenant.
func (q *ListRolesQuery) Handle(ctx context.Context, req ListRolesRequest) (*ListRolesResponse, error) {
	// Set defaults
	if req.Limit <= 0 {
		req.Limit = 50
	}

	filters := &domain.RoleFilters{
		IncludeSystem: req.IncludeSystem,
		Search:        req.Search,
		Limit:         req.Limit,
		Offset:        req.Offset,
	}

	roles, err := q.roleRepo.FindAll(ctx, req.TenantID, filters)
	if err != nil {
		q.logger.Error("failed to list roles", logger.Err(err))
		return nil, err
	}

	return &ListRolesResponse{
		Roles: dto.RolesToSummaryDTOs(roles),
		Total: len(roles),
	}, nil
}
