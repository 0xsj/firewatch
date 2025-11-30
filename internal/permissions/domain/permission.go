package domain

import (
	"strings"
)

// ============================================================================
// Action
// ============================================================================

// Action represents an operation that can be performed on a resource.
type Action string

const (
	ActionCreate Action = "create"
	ActionRead   Action = "read"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
	ActionManage Action = "manage" // Implies all CRUD operations
	ActionAll    Action = "*"      // Wildcard - all actions
)

// validActions is the set of valid actions.
var validActions = map[Action]bool{
	ActionCreate: true,
	ActionRead:   true,
	ActionUpdate: true,
	ActionDelete: true,
	ActionManage: true,
	ActionAll:    true,
}

// IsValid checks if the action is valid.
func (a Action) IsValid() bool {
	return validActions[a]
}

// String returns the string representation.
func (a Action) String() string {
	return string(a)
}

// Implies checks if this action implies another action.
// For example, ActionManage implies ActionCreate, ActionRead, etc.
func (a Action) Implies(other Action) bool {
	if a == ActionAll {
		return true
	}
	if a == ActionManage {
		return other == ActionCreate || other == ActionRead ||
			other == ActionUpdate || other == ActionDelete
	}
	return a == other
}

// ============================================================================
// Resource
// ============================================================================

// Resource represents something that can be acted upon.
type Resource string

const (
	ResourceUsers     Resource = "users"
	ResourceTenants   Resource = "tenants"
	ResourceFlags     Resource = "flags"
	ResourceRoles     Resource = "roles"
	ResourceAPIKeys   Resource = "api_keys"
	ResourceAuditLogs Resource = "audit_logs"
	ResourceSelf      Resource = "self" // Special: user's own resources
	ResourceAll       Resource = "*"    // Wildcard - all resources
)

// validResources is the set of valid resources.
var validResources = map[Resource]bool{
	ResourceUsers:     true,
	ResourceTenants:   true,
	ResourceFlags:     true,
	ResourceRoles:     true,
	ResourceAPIKeys:   true,
	ResourceAuditLogs: true,
	ResourceSelf:      true,
	ResourceAll:       true,
}

// IsValid checks if the resource is valid.
func (r Resource) IsValid() bool {
	return validResources[r]
}

// String returns the string representation.
func (r Resource) String() string {
	return string(r)
}

// Implies checks if this resource implies another resource.
func (r Resource) Implies(other Resource) bool {
	if r == ResourceAll {
		return true
	}
	return r == other
}

// ============================================================================
// Permission
// ============================================================================

// Permission represents the ability to perform an action on a resource.
// Format: "action:resource" (e.g., "read:users", "manage:flags", "*:*")
type Permission struct {
	action   Action
	resource Resource
}

// NewPermission creates a new Permission with validation.
func NewPermission(action Action, resource Resource) (Permission, error) {
	const op = "Permission.New"

	if !action.IsValid() {
		return Permission{}, ErrActionInvalid(op, string(action))
	}

	if !resource.IsValid() {
		return Permission{}, ErrResourceInvalid(op, string(resource))
	}

	return Permission{
		action:   action,
		resource: resource,
	}, nil
}

// MustNewPermission creates a new Permission, panicking on error.
// Use only for known-valid permissions (e.g., constants).
func MustNewPermission(action Action, resource Resource) Permission {
	p, err := NewPermission(action, resource)
	if err != nil {
		panic(err)
	}
	return p
}

// ParsePermission parses a permission string (e.g., "read:users").
func ParsePermission(s string) (Permission, error) {
	const op = "Permission.Parse"

	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return Permission{}, ErrPermissionInvalid(op, s, "must be in format action:resource")
	}

	action := Action(parts[0])
	resource := Resource(parts[1])

	return NewPermission(action, resource)
}

// Action returns the action component.
func (p Permission) Action() Action {
	return p.action
}

// Resource returns the resource component.
func (p Permission) Resource() Resource {
	return p.resource
}

// String returns the string representation (e.g., "read:users").
func (p Permission) String() string {
	return p.action.String() + ":" + p.resource.String()
}

// IsEmpty returns true if the permission is empty.
func (p Permission) IsEmpty() bool {
	return p.action == "" && p.resource == ""
}

// Implies checks if this permission implies another permission.
// For example, "manage:users" implies "read:users".
// And "*:*" implies everything.
func (p Permission) Implies(other Permission) bool {
	return p.action.Implies(other.action) && p.resource.Implies(other.resource)
}

// Equals checks if two permissions are equal.
func (p Permission) Equals(other Permission) bool {
	return p.action == other.action && p.resource == other.resource
}

// ============================================================================
// Predefined Permissions
// ============================================================================

var (
	// User permissions
	PermissionCreateUsers = MustNewPermission(ActionCreate, ResourceUsers)
	PermissionReadUsers   = MustNewPermission(ActionRead, ResourceUsers)
	PermissionUpdateUsers = MustNewPermission(ActionUpdate, ResourceUsers)
	PermissionDeleteUsers = MustNewPermission(ActionDelete, ResourceUsers)
	PermissionManageUsers = MustNewPermission(ActionManage, ResourceUsers)

	// Tenant permissions
	PermissionCreateTenants = MustNewPermission(ActionCreate, ResourceTenants)
	PermissionReadTenants   = MustNewPermission(ActionRead, ResourceTenants)
	PermissionUpdateTenants = MustNewPermission(ActionUpdate, ResourceTenants)
	PermissionDeleteTenants = MustNewPermission(ActionDelete, ResourceTenants)
	PermissionManageTenants = MustNewPermission(ActionManage, ResourceTenants)

	// Flag permissions
	PermissionCreateFlags = MustNewPermission(ActionCreate, ResourceFlags)
	PermissionReadFlags   = MustNewPermission(ActionRead, ResourceFlags)
	PermissionUpdateFlags = MustNewPermission(ActionUpdate, ResourceFlags)
	PermissionDeleteFlags = MustNewPermission(ActionDelete, ResourceFlags)
	PermissionManageFlags = MustNewPermission(ActionManage, ResourceFlags)

	// Role permissions
	PermissionCreateRoles = MustNewPermission(ActionCreate, ResourceRoles)
	PermissionReadRoles   = MustNewPermission(ActionRead, ResourceRoles)
	PermissionUpdateRoles = MustNewPermission(ActionUpdate, ResourceRoles)
	PermissionDeleteRoles = MustNewPermission(ActionDelete, ResourceRoles)
	PermissionManageRoles = MustNewPermission(ActionManage, ResourceRoles)

	// API Key permissions
	PermissionCreateAPIKeys = MustNewPermission(ActionCreate, ResourceAPIKeys)
	PermissionReadAPIKeys   = MustNewPermission(ActionRead, ResourceAPIKeys)
	PermissionDeleteAPIKeys = MustNewPermission(ActionDelete, ResourceAPIKeys)
	PermissionManageAPIKeys = MustNewPermission(ActionManage, ResourceAPIKeys)

	// Audit log permissions
	PermissionReadAuditLogs = MustNewPermission(ActionRead, ResourceAuditLogs)

	// Self permissions (user's own resources)
	PermissionReadSelf   = MustNewPermission(ActionRead, ResourceSelf)
	PermissionUpdateSelf = MustNewPermission(ActionUpdate, ResourceSelf)
	PermissionDeleteSelf = MustNewPermission(ActionDelete, ResourceSelf)

	// Super admin - full access
	PermissionSuperAdmin = MustNewPermission(ActionAll, ResourceAll)
)

// ============================================================================
// Permission Set
// ============================================================================

// PermissionSet is a collection of permissions with helper methods.
type PermissionSet []Permission

// Contains checks if the set contains a permission that implies the given permission.
func (ps PermissionSet) Contains(p Permission) bool {
	for _, perm := range ps {
		if perm.Implies(p) {
			return true
		}
	}
	return false
}

// Add adds a permission to the set if it doesn't already exist.
func (ps *PermissionSet) Add(p Permission) {
	for _, existing := range *ps {
		if existing.Equals(p) {
			return
		}
	}
	*ps = append(*ps, p)
}

// Remove removes a permission from the set.
func (ps *PermissionSet) Remove(p Permission) {
	for i, existing := range *ps {
		if existing.Equals(p) {
			*ps = append((*ps)[:i], (*ps)[i+1:]...)
			return
		}
	}
}

// Strings returns the string representations of all permissions.
func (ps PermissionSet) Strings() []string {
	result := make([]string, len(ps))
	for i, p := range ps {
		result[i] = p.String()
	}
	return result
}

// ParsePermissionSet parses a slice of permission strings.
func ParsePermissionSet(strings []string) (PermissionSet, error) {
	perms := make(PermissionSet, 0, len(strings))
	for _, s := range strings {
		p, err := ParsePermission(s)
		if err != nil {
			return nil, err
		}
		perms.Add(p)
	}
	return perms, nil
}
