package query

import (
	"context"

	"github.com/0xsj/hexagonal-go/internal/permissions/application/dto"
	"github.com/0xsj/hexagonal-go/internal/permissions/domain"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// CheckPermissionRequest represents the input for checking a permission.
type CheckPermissionRequest struct {
	UserID       types.ID `json:"user_id"`
	TenantID     string   `json:"tenant_id"`
	IdentityRole string   `json:"identity_role"`
	Action       string   `json:"action"`
	Resource     string   `json:"resource"`
	ResourceID   string   `json:"resource_id,omitempty"` // For resource-specific checks
}

// CheckPermissionResponse represents the output of checking a permission.
type CheckPermissionResponse struct {
	Result dto.PermissionCheckDTO `json:"result"`
}

// CheckPermissionQuery handles permission checks.
type CheckPermissionQuery struct {
	roleRepo       domain.RoleRepository
	assignmentRepo domain.AssignmentRepository
	logger         logger.Logger
}

// NewCheckPermissionQuery creates a new CheckPermissionQuery.
func NewCheckPermissionQuery(
	roleRepo domain.RoleRepository,
	assignmentRepo domain.AssignmentRepository,
	logger logger.Logger,
) *CheckPermissionQuery {
	return &CheckPermissionQuery{
		roleRepo:       roleRepo,
		assignmentRepo: assignmentRepo,
		logger:         logger,
	}
}

// Handle checks if a user has a specific permission.
func (q *CheckPermissionQuery) Handle(ctx context.Context, req CheckPermissionRequest) (*CheckPermissionResponse, error) {
	// Parse the requested permission
	permission, err := domain.NewPermission(domain.Action(req.Action), domain.Resource(req.Resource))
	if err != nil {
		return &CheckPermissionResponse{
			Result: dto.PermissionCheckDTO{
				Allowed:    false,
				Permission: req.Action + ":" + req.Resource,
				Reason:     "invalid permission format",
			},
		}, nil
	}

	// 1. Check default permissions from identity role
	defaultPerms := domain.GetDefaultPermissions(req.IdentityRole)
	if defaultPerms.Contains(permission) {
		return &CheckPermissionResponse{
			Result: dto.PermissionCheckDTO{
				Allowed:    true,
				Permission: permission.String(),
				Reason:     "granted by identity role: " + req.IdentityRole,
			},
		}, nil
	}

	// 2. Check explicit role assignments
	filters := &domain.AssignmentFilters{
		TenantID:       req.TenantID,
		ResourceID:     req.ResourceID,
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

	for _, assignment := range assignments {
		if !assignment.IsActive() {
			continue
		}

		// Check scope match
		if !assignment.MatchesScope(req.TenantID, req.ResourceID) {
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

		if role.HasPermission(permission) {
			return &CheckPermissionResponse{
				Result: dto.PermissionCheckDTO{
					Allowed:    true,
					Permission: permission.String(),
					Reason:     "granted by role: " + role.Name(),
				},
			}, nil
		}
	}

	// Permission denied
	return &CheckPermissionResponse{
		Result: dto.PermissionCheckDTO{
			Allowed:    false,
			Permission: permission.String(),
			Reason:     "no matching role or permission found",
		},
	}, nil
}

// HasPermission checks if a user has a specific permission.
// Implements middleware.PermissionChecker interface.
func (q *CheckPermissionQuery) HasPermission(ctx context.Context, userID, tenantID, identityRole, action, resource string) (bool, error) {
	// Parse the permission
	permission, err := domain.NewPermission(domain.Action(action), domain.Resource(resource))
	if err != nil {
		return false, nil
	}

	// 1. Check default permissions from identity role
	defaultPerms := domain.GetDefaultPermissions(identityRole)
	if defaultPerms.Contains(permission) {
		return true, nil
	}

	// 2. Check explicit role assignments
	uid, err := types.ParseID(userID)
	if err != nil {
		return false, nil
	}

	filters := &domain.AssignmentFilters{
		TenantID:       tenantID,
		IncludeExpired: false,
	}

	assignments, err := q.assignmentRepo.FindByUser(ctx, uid, filters)
	if err != nil {
		return false, err
	}

	for _, assignment := range assignments {
		if !assignment.IsActive() {
			continue
		}

		role, err := q.roleRepo.FindByID(ctx, assignment.RoleID())
		if err != nil {
			continue
		}

		if role.HasPermission(permission) {
			return true, nil
		}
	}

	return false, nil
}
