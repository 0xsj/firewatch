package query

import (
	"context"

	"github.com/0xsj/hexagonal-go/internal/permissions/application/dto"
	"github.com/0xsj/hexagonal-go/internal/permissions/domain"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// GetUserPermissionsRequest represents the input for getting user permissions.
type GetUserPermissionsRequest struct {
	UserID       types.ID `json:"user_id"`
	TenantID     string   `json:"tenant_id"`
	IdentityRole string   `json:"identity_role"` // User's role from identity domain
}

// GetUserPermissionsResponse represents the output of getting user permissions.
type GetUserPermissionsResponse struct {
	Permissions dto.UserPermissionsDTO `json:"permissions"`
}

// GetUserPermissionsQuery handles user permission retrieval.
type GetUserPermissionsQuery struct {
	roleRepo       domain.RoleRepository
	assignmentRepo domain.AssignmentRepository
	logger         logger.Logger
}

// NewGetUserPermissionsQuery creates a new GetUserPermissionsQuery.
func NewGetUserPermissionsQuery(
	roleRepo domain.RoleRepository,
	assignmentRepo domain.AssignmentRepository,
	logger logger.Logger,
) *GetUserPermissionsQuery {
	return &GetUserPermissionsQuery{
		roleRepo:       roleRepo,
		assignmentRepo: assignmentRepo,
		logger:         logger,
	}
}

// Handle retrieves all effective permissions for a user.
func (q *GetUserPermissionsQuery) Handle(ctx context.Context, req GetUserPermissionsRequest) (*GetUserPermissionsResponse, error) {
	// Collect all permissions
	allPermissions := domain.PermissionSet{}
	roleNames := []string{}

	// 1. Get default permissions from identity role
	defaultPerms := domain.GetDefaultPermissions(req.IdentityRole)
	for _, p := range defaultPerms {
		allPermissions.Add(p)
	}

	if req.IdentityRole != "" {
		roleNames = append(roleNames, req.IdentityRole)
	}

	// 2. Get explicit role assignments
	filters := &domain.AssignmentFilters{
		TenantID:       req.TenantID,
		IncludeExpired: false,
	}

	assignments, err := q.assignmentRepo.FindByUser(ctx, req.UserID, filters)
	if err != nil {
		q.logger.Error("failed to get user assignments",
			logger.String("user_id", req.UserID.String()),
			logger.Err(err),
		)
		return nil, err
	}

	// 3. Get permissions from assigned roles
	for _, assignment := range assignments {
		if !assignment.IsActive() {
			continue
		}

		role, err := q.roleRepo.FindByID(ctx, assignment.RoleID())
		if err != nil {
			q.logger.Warn("failed to get role for assignment",
				logger.String("role_id", assignment.RoleID().String()),
				logger.Err(err),
			)
			continue
		}

		roleNames = append(roleNames, role.Name())

		for _, p := range role.Permissions() {
			allPermissions.Add(p)
		}
	}

	return &GetUserPermissionsResponse{
		Permissions: dto.UserPermissionsDTO{
			UserID:      req.UserID.String(),
			TenantID:    req.TenantID,
			Permissions: allPermissions.Strings(),
			Roles:       roleNames,
		},
	}, nil
}
