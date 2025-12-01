// internal/permissions/application/dto/role.go
package dto

import "time"

// RoleDTO represents a role for API responses.
type RoleDTO struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Permissions []string  `json:"permissions"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RoleSummaryDTO represents a role summary for list responses.
type RoleSummaryDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsSystem    bool   `json:"is_system"`
	Permissions int    `json:"permissions_count"`
}

// AssignmentDTO represents a role assignment for API responses.
type AssignmentDTO struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	RoleID     string     `json:"role_id"`
	RoleName   string     `json:"role_name,omitempty"`
	TenantID   string     `json:"tenant_id"`
	ResourceID string     `json:"resource_id,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	CreatedBy  string     `json:"created_by"`
}

// PermissionCheckDTO represents a permission check result.
type PermissionCheckDTO struct {
	Allowed    bool   `json:"allowed"`
	Permission string `json:"permission"`
	Reason     string `json:"reason,omitempty"`
}

// UserPermissionsDTO represents all permissions for a user.
type UserPermissionsDTO struct {
	UserID      string   `json:"user_id"`
	TenantID    string   `json:"tenant_id"`
	Permissions []string `json:"permissions"`
	Roles       []string `json:"roles"`
}
