package user

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Role represents a user's role in the system.
// Roles determine permissions and access levels.
type Role string

const (
	// RoleGuest has minimal permissions (read-only, public data).
	RoleGuest Role = "guest"

	// RoleUser has standard user permissions.
	// Can manage their own data but not others.
	RoleUser Role = "user"

	// RoleModerator has elevated permissions.
	// Can moderate content and manage users within their tenant.
	RoleModerator Role = "moderator"

	// RoleAdmin has full permissions within their tenant.
	// Can manage all users, settings, and data for their tenant.
	RoleAdmin Role = "admin"

	// RoleSuperAdmin has global permissions across all tenants.
	// Platform-level administration.
	RoleSuperAdmin Role = "super_admin"
)

// AllRoles returns all valid user roles.
func AllRoles() []Role {
	return []Role{
		RoleGuest,
		RoleUser,
		RoleModerator,
		RoleAdmin,
		RoleSuperAdmin,
	}
}

// IsValid returns true if the role is a valid user role.
func (r Role) IsValid() bool {
	switch r {
	case RoleGuest, RoleUser, RoleModerator, RoleAdmin, RoleSuperAdmin:
		return true
	default:
		return false
	}
}

// Validate returns an error if the role is invalid.
func (r Role) Validate() error {
	if !r.IsValid() {
		return fmt.Errorf("invalid user role: %s", r)
	}
	return nil
}

// String returns the string representation of the role.
func (r Role) String() string {
	return string(r)
}

// ============================================================================
// Role Hierarchy
// ============================================================================

// Level returns the hierarchical level of the role.
// Higher numbers indicate more permissions.
// Used for comparing roles and checking hierarchy.
func (r Role) Level() int {
	switch r {
	case RoleGuest:
		return 0
	case RoleUser:
		return 10
	case RoleModerator:
		return 20
	case RoleAdmin:
		return 30
	case RoleSuperAdmin:
		return 40
	default:
		return -1
	}
}

// IsHigherThan returns true if this role has higher permissions than the other.
func (r Role) IsHigherThan(other Role) bool {
	return r.Level() > other.Level()
}

// IsHigherOrEqualTo returns true if this role has higher or equal permissions.
func (r Role) IsHigherOrEqualTo(other Role) bool {
	return r.Level() >= other.Level()
}

// IsLowerThan returns true if this role has lower permissions than the other.
func (r Role) IsLowerThan(other Role) bool {
	return r.Level() < other.Level()
}

// IsLowerOrEqualTo returns true if this role has lower or equal permissions.
func (r Role) IsLowerOrEqualTo(other Role) bool {
	return r.Level() <= other.Level()
}

// ============================================================================
// Role Checks
// ============================================================================

// IsGuest returns true if the role is guest.
func (r Role) IsGuest() bool {
	return r == RoleGuest
}

// IsUser returns true if the role is user.
func (r Role) IsUser() bool {
	return r == RoleUser
}

// IsModerator returns true if the role is moderator.
func (r Role) IsModerator() bool {
	return r == RoleModerator
}

// IsAdmin returns true if the role is admin.
func (r Role) IsAdmin() bool {
	return r == RoleAdmin
}

// IsSuperAdmin returns true if the role is super admin.
func (r Role) IsSuperAdmin() bool {
	return r == RoleSuperAdmin
}

// IsAdminOrHigher returns true if the role is admin or super admin.
func (r Role) IsAdminOrHigher() bool {
	return r.IsHigherOrEqualTo(RoleAdmin)
}

// IsModeratorOrHigher returns true if the role is moderator or higher.
func (r Role) IsModeratorOrHigher() bool {
	return r.IsHigherOrEqualTo(RoleModerator)
}

// ============================================================================
// Permissions
// ============================================================================

// CanManageUsers returns true if the role can manage other users.
// Moderators and above can manage users.
func (r Role) CanManageUsers() bool {
	return r.IsModeratorOrHigher()
}

// CanManageTenant returns true if the role can manage tenant settings.
// Only admins can manage tenant settings.
func (r Role) CanManageTenant() bool {
	return r.IsAdminOrHigher()
}

// CanAccessAllTenants returns true if the role can access all tenants.
// Only super admins can access all tenants.
func (r Role) CanAccessAllTenants() bool {
	return r.IsSuperAdmin()
}

// CanPromoteUser returns true if this role can promote a user to the target role.
// Users can only promote to roles lower than their own.
func (r Role) CanPromoteUser(targetRole Role) bool {
	// Can't promote to a role equal or higher than your own
	return r.IsHigherThan(targetRole)
}

// CanDemoteUser returns true if this role can demote a user from the source role.
// Users can only demote users with roles lower than their own.
func (r Role) CanDemoteUser(sourceRole Role) bool {
	// Can't demote a user with a role equal or higher than your own
	return r.IsHigherThan(sourceRole)
}

// CanModifyUser returns true if this role can modify a user with the given role.
// Users can only modify users with roles lower than their own.
func (r Role) CanModifyUser(targetRole Role) bool {
	return r.IsHigherThan(targetRole)
}

// ============================================================================
// Role Transitions
// ============================================================================

// CanBePromotedTo checks if this role can be promoted to the target role.
func (r Role) CanBePromotedTo(target Role) error {
	if err := target.Validate(); err != nil {
		return err
	}

	if target.IsLowerOrEqualTo(r) {
		return fmt.Errorf("cannot promote from %s to %s: target role is not higher", r, target)
	}

	// Super admin promotion requires special handling (usually manual/platform decision)
	if target == RoleSuperAdmin {
		return fmt.Errorf("cannot promote to super_admin: requires platform approval")
	}

	return nil
}

// CanBeDemotedTo checks if this role can be demoted to the target role.
func (r Role) CanBeDemotedTo(target Role) error {
	if err := target.Validate(); err != nil {
		return err
	}

	if target.IsHigherOrEqualTo(r) {
		return fmt.Errorf("cannot demote from %s to %s: target role is not lower", r, target)
	}

	return nil
}

// ============================================================================
// Parsing
// ============================================================================

// ParseRole parses a string into a Role.
// Returns an error if the string is not a valid role.
func ParseRole(s string) (Role, error) {
	role := Role(s)
	if err := role.Validate(); err != nil {
		return "", err
	}
	return role, nil
}

// MustParseRole parses a string into a Role and panics if invalid.
// Only use this for constants where you're certain the value is valid.
func MustParseRole(s string) Role {
	role, err := ParseRole(s)
	if err != nil {
		panic(fmt.Sprintf("invalid role: %v", err))
	}
	return role
}

// ============================================================================
// Database Marshaling
// ============================================================================

// Scan implements sql.Scanner for reading from database.
func (r *Role) Scan(value interface{}) error {
	if value == nil {
		*r = ""
		return nil
	}

	var str string
	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("cannot scan %T into Role", value)
	}

	role, err := ParseRole(str)
	if err != nil {
		return err
	}

	*r = role
	return nil
}

// Value implements driver.Valuer for writing to database.
func (r Role) Value() (driver.Value, error) {
	if r == "" {
		return nil, nil
	}

	if err := r.Validate(); err != nil {
		return nil, err
	}

	return string(r), nil
}

// ============================================================================
// JSON Marshaling
// ============================================================================

// MarshalJSON implements json.Marshaler.
func (r Role) MarshalJSON() ([]byte, error) {
	if r == "" {
		return []byte("null"), nil
	}
	return fmt.Appendf(nil, "%q", r), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *Role) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	if str == "" {
		*r = ""
		return nil
	}

	role, err := ParseRole(str)
	if err != nil {
		return err
	}

	*r = role
	return nil
}

// ============================================================================
// Default Role
// ============================================================================

// DefaultRole returns the default role for new users.
func DefaultRole() Role {
	return RoleUser
}
