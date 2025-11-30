package domain

import (
	"regexp"
	"strings"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Role is the aggregate root for permission roles.
// It encapsulates a named collection of permissions.
type Role struct {
	id          types.ID
	tenantID    string
	name        string
	description string
	permissions PermissionSet
	isSystem    bool
	createdAt   types.Timestamp
	updatedAt   types.Timestamp
	version     int
	events      []Event
}

// Role name constraints.
const (
	RoleNameMinLength = 2
	RoleNameMaxLength = 64
)

// roleNamePattern defines valid role name format.
// Allows lowercase letters, numbers, underscores, and hyphens.
var roleNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

// ============================================================================
// Getters
// ============================================================================

// ID returns the role ID.
func (r *Role) ID() types.ID {
	return r.id
}

// TenantID returns the tenant ID.
func (r *Role) TenantID() string {
	return r.tenantID
}

// Name returns the role name.
func (r *Role) Name() string {
	return r.name
}

// Description returns the role description.
func (r *Role) Description() string {
	return r.description
}

// Permissions returns the role permissions.
func (r *Role) Permissions() PermissionSet {
	return r.permissions
}

// IsSystem returns whether this is a system role.
func (r *Role) IsSystem() bool {
	return r.isSystem
}

// CreatedAt returns the creation timestamp.
func (r *Role) CreatedAt() types.Timestamp {
	return r.createdAt
}

// UpdatedAt returns the last update timestamp.
func (r *Role) UpdatedAt() types.Timestamp {
	return r.updatedAt
}

// Version returns the aggregate version.
func (r *Role) Version() int {
	return r.version
}

// Events returns uncommitted domain events.
func (r *Role) Events() []Event {
	return r.events
}

// ClearEvents clears uncommitted domain events.
func (r *Role) ClearEvents() {
	r.events = nil
}

// ============================================================================
// Factory Methods
// ============================================================================

// CreateRole creates a new role.
// Emits RoleCreated event.
func CreateRole(
	id types.ID,
	tenantID string,
	name string,
	description string,
	permissions PermissionSet,
	createdBy string,
) (*Role, error) {
	const op = "Role.Create"

	if err := validateRoleName(op, name); err != nil {
		return nil, err
	}

	now := types.Now()

	role := &Role{
		id:          id,
		tenantID:    tenantID,
		name:        name,
		description: description,
		permissions: permissions,
		isSystem:    false,
		createdAt:   now,
		updatedAt:   now,
		version:     1,
		events:      make([]Event, 0),
	}

	role.addEvent(NewRoleCreated(id, tenantID, name, description, permissions, false, createdBy))

	return role, nil
}

// CreateSystemRole creates a system role (cannot be modified/deleted).
// Does NOT emit events - system roles are predefined.
func CreateSystemRole(
	id types.ID,
	name string,
	description string,
	permissions PermissionSet,
) *Role {
	now := types.Now()

	return &Role{
		id:          id,
		tenantID:    "*", // System roles apply to all tenants
		name:        name,
		description: description,
		permissions: permissions,
		isSystem:    true,
		createdAt:   now,
		updatedAt:   now,
		version:     1,
		events:      make([]Event, 0),
	}
}

// Reconstitute recreates a role from stored state (used by repository).
// Does NOT emit events - only for loading from database.
func Reconstitute(
	id types.ID,
	tenantID string,
	name string,
	description string,
	permissions PermissionSet,
	isSystem bool,
	createdAt types.Timestamp,
	updatedAt types.Timestamp,
	version int,
) *Role {
	return &Role{
		id:          id,
		tenantID:    tenantID,
		name:        name,
		description: description,
		permissions: permissions,
		isSystem:    isSystem,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		version:     version,
		events:      make([]Event, 0),
	}
}

// ============================================================================
// Commands
// ============================================================================

// Update updates the role's metadata.
// Emits RoleUpdated event.
func (r *Role) Update(name, description string, updatedBy string) error {
	const op = "Role.Update"

	if r.isSystem {
		return ErrRoleIsSystem(op, r.name)
	}

	var updatedFields []string

	name = strings.TrimSpace(name)
	if name != "" && name != r.name {
		if err := validateRoleName(op, name); err != nil {
			return err
		}
		r.name = name
		updatedFields = append(updatedFields, "name")
	}

	description = strings.TrimSpace(description)
	if description != r.description {
		r.description = description
		updatedFields = append(updatedFields, "description")
	}

	if len(updatedFields) > 0 {
		r.updatedAt = types.Now()
		r.version++
		r.addEvent(NewRoleUpdated(r.id, r.tenantID, r.name, updatedFields, updatedBy, r.version))
	}

	return nil
}

// AddPermission adds a permission to the role.
// Emits RolePermissionAdded event.
func (r *Role) AddPermission(permission Permission, addedBy string) error {
	const op = "Role.AddPermission"

	if r.isSystem {
		return ErrRoleIsSystem(op, r.name)
	}

	// Check if already exists
	for _, p := range r.permissions {
		if p.Equals(permission) {
			return nil // Already has this permission
		}
	}

	r.permissions.Add(permission)
	r.updatedAt = types.Now()
	r.version++
	r.addEvent(NewRolePermissionAdded(r.id, r.tenantID, r.name, permission, addedBy, r.version))

	return nil
}

// RemovePermission removes a permission from the role.
// Emits RolePermissionRemoved event.
func (r *Role) RemovePermission(permission Permission, removedBy string) error {
	const op = "Role.RemovePermission"

	if r.isSystem {
		return ErrRoleIsSystem(op, r.name)
	}

	// Check if exists
	found := false
	for _, p := range r.permissions {
		if p.Equals(permission) {
			found = true
			break
		}
	}

	if !found {
		return nil // Permission not found, nothing to remove
	}

	r.permissions.Remove(permission)
	r.updatedAt = types.Now()
	r.version++
	r.addEvent(NewRolePermissionRemoved(r.id, r.tenantID, r.name, permission, removedBy, r.version))

	return nil
}

// SetPermissions replaces all permissions on the role.
func (r *Role) SetPermissions(permissions PermissionSet, updatedBy string) error {
	const op = "Role.SetPermissions"

	if r.isSystem {
		return ErrRoleIsSystem(op, r.name)
	}

	r.permissions = permissions
	r.updatedAt = types.Now()
	r.version++
	r.addEvent(NewRoleUpdated(r.id, r.tenantID, r.name, []string{"permissions"}, updatedBy, r.version))

	return nil
}

// MarkDeleted prepares the role for deletion.
// Emits RoleDeleted event.
func (r *Role) MarkDeleted(deletedBy string) error {
	const op = "Role.MarkDeleted"

	if r.isSystem {
		return ErrRoleIsSystem(op, r.name)
	}

	r.version++
	r.addEvent(NewRoleDeleted(r.id, r.tenantID, r.name, deletedBy, r.version))

	return nil
}

// ============================================================================
// Query Methods
// ============================================================================

// HasPermission checks if the role has a permission (including implied).
func (r *Role) HasPermission(permission Permission) bool {
	return r.permissions.Contains(permission)
}

// ============================================================================
// Internal
// ============================================================================

// addEvent adds a domain event.
func (r *Role) addEvent(event Event) {
	r.events = append(r.events, event)
}

// validateRoleName validates a role name.
func validateRoleName(op, name string) error {
	name = strings.TrimSpace(name)

	if len(name) < RoleNameMinLength {
		return ErrRoleNameInvalid(op, name, "role name must be at least 2 characters")
	}

	if len(name) > RoleNameMaxLength {
		return ErrRoleNameInvalid(op, name, "role name must be at most 64 characters")
	}

	if !roleNamePattern.MatchString(name) {
		return ErrRoleNameInvalid(op, name, "role name must start with a letter and contain only lowercase letters, numbers, underscores, and hyphens")
	}

	return nil
}

// ============================================================================
// Predefined System Roles
// ============================================================================

// System role names that map to identity.user.Role
const (
	SystemRoleNameSuperAdmin = "super_admin"
	SystemRoleNameAdmin      = "admin"
	SystemRoleNameModerator  = "moderator"
	SystemRoleNameUser       = "user"
	SystemRoleNameGuest      = "guest"
)

// DefaultRolePermissions maps identity roles to their default permissions.
var DefaultRolePermissions = map[string]PermissionSet{
	SystemRoleNameGuest: {},

	SystemRoleNameUser: {
		PermissionReadSelf,
		PermissionUpdateSelf,
	},

	SystemRoleNameModerator: {
		PermissionReadSelf,
		PermissionUpdateSelf,
		PermissionReadUsers,
		PermissionReadFlags,
		PermissionReadAuditLogs,
	},

	SystemRoleNameAdmin: {
		PermissionManageUsers,
		PermissionManageFlags,
		PermissionReadAuditLogs,
		PermissionManageRoles,
		PermissionManageAPIKeys,
	},

	SystemRoleNameSuperAdmin: {
		PermissionSuperAdmin, // *:* - full access
	},
}

// GetDefaultPermissions returns the default permissions for an identity role.
func GetDefaultPermissions(identityRole string) PermissionSet {
	if perms, ok := DefaultRolePermissions[identityRole]; ok {
		return perms
	}
	return PermissionSet{}
}
