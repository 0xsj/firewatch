package domain

import (
	"context"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// ============================================================================
// Role Repository
// ============================================================================

// RoleRepository defines the interface for role persistence.
type RoleRepository interface {
	// Save persists a role (create or update).
	Save(ctx context.Context, role *Role) error

	// FindByID retrieves a role by its ID.
	FindByID(ctx context.Context, id types.ID) (*Role, error)

	// FindByName retrieves a role by name within a tenant.
	FindByName(ctx context.Context, tenantID, name string) (*Role, error)

	// FindAll retrieves all roles for a tenant.
	FindAll(ctx context.Context, tenantID string, filters *RoleFilters) ([]*Role, error)

	// FindSystemRoles retrieves all system roles.
	FindSystemRoles(ctx context.Context) ([]*Role, error)

	// Delete removes a role by ID.
	Delete(ctx context.Context, id types.ID) error

	// Exists checks if a role name exists within a tenant.
	Exists(ctx context.Context, tenantID, name string) (bool, error)
}

// RoleFilters defines optional filters for listing roles.
type RoleFilters struct {
	IncludeSystem bool   // Include system roles in results
	Search        string // Search in name and description
	Limit         int
	Offset        int
}

// ============================================================================
// Assignment Repository
// ============================================================================

// AssignmentRepository defines the interface for assignment persistence.
type AssignmentRepository interface {
	// Save persists an assignment.
	Save(ctx context.Context, assignment *Assignment) error

	// FindByID retrieves an assignment by its ID.
	FindByID(ctx context.Context, id types.ID) (*Assignment, error)

	// FindByUserAndRole retrieves an assignment by user, role, and scope.
	FindByUserAndRole(ctx context.Context, userID, roleID types.ID, tenantID, resourceID string) (*Assignment, error)

	// FindByUser retrieves all assignments for a user.
	FindByUser(ctx context.Context, userID types.ID, filters *AssignmentFilters) ([]*Assignment, error)

	// FindByRole retrieves all assignments for a role.
	FindByRole(ctx context.Context, roleID types.ID, filters *AssignmentFilters) ([]*Assignment, error)

	// FindByTenant retrieves all assignments within a tenant.
	FindByTenant(ctx context.Context, tenantID string, filters *AssignmentFilters) ([]*Assignment, error)

	// Delete removes an assignment by ID.
	Delete(ctx context.Context, id types.ID) error

	// DeleteByUserAndRole removes an assignment by user, role, and scope.
	DeleteByUserAndRole(ctx context.Context, userID, roleID types.ID, tenantID, resourceID string) error

	// DeleteByUser removes all assignments for a user.
	DeleteByUser(ctx context.Context, userID types.ID) error

	// DeleteByRole removes all assignments for a role.
	DeleteByRole(ctx context.Context, roleID types.ID) error

	// DeleteExpired removes all expired assignments.
	DeleteExpired(ctx context.Context) (int64, error)

	// Exists checks if an assignment exists.
	Exists(ctx context.Context, userID, roleID types.ID, tenantID, resourceID string) (bool, error)
}

// AssignmentFilters defines optional filters for listing assignments.
type AssignmentFilters struct {
	TenantID       string // Filter by tenant
	ResourceID     string // Filter by resource
	IncludeExpired bool   // Include expired assignments
	Limit          int
	Offset         int
}

// ============================================================================
// Permission Query Service
// ============================================================================

// PermissionQueryService provides efficient permission lookups.
// This is a read-optimized interface for authorization checks.
type PermissionQueryService interface {
	// GetUserPermissions returns all effective permissions for a user.
	// Combines permissions from:
	// 1. User's identity role (default permissions)
	// 2. Explicit role assignments
	GetUserPermissions(ctx context.Context, userID types.ID, tenantID string) (PermissionSet, error)

	// GetUserPermissionsForResource returns permissions for a specific resource.
	GetUserPermissionsForResource(ctx context.Context, userID types.ID, tenantID, resourceID string) (PermissionSet, error)

	// HasPermission checks if a user has a specific permission.
	HasPermission(ctx context.Context, userID types.ID, tenantID string, permission Permission) (bool, error)

	// HasPermissionForResource checks if a user has a permission for a specific resource.
	HasPermissionForResource(ctx context.Context, userID types.ID, tenantID, resourceID string, permission Permission) (bool, error)

	// GetUserRoles returns all roles assigned to a user.
	GetUserRoles(ctx context.Context, userID types.ID, tenantID string) ([]*Role, error)
}
