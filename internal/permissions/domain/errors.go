package domain

import (
	"github.com/0xsj/hexagonal-go/pkg/errors"
)

// Domain-specific error codes for the Permissions domain.
// These codes are stable across versions and used by clients.
//
// Naming convention: PERMISSION_RESOURCE_CONDITION
const (
	// Role errors
	ErrCodeRoleNotFound      = errors.Code("ROLE_NOT_FOUND")
	ErrCodeRoleAlreadyExists = errors.Code("ROLE_ALREADY_EXISTS")
	ErrCodeRoleIsSystem      = errors.Code("ROLE_IS_SYSTEM")
	ErrCodeRoleNameInvalid   = errors.Code("ROLE_NAME_INVALID")

	// Assignment errors
	ErrCodeAssignmentNotFound = errors.Code("ASSIGNMENT_NOT_FOUND")
	ErrCodeAssignmentExists   = errors.Code("ASSIGNMENT_ALREADY_EXISTS")
	ErrCodeAssignmentExpired  = errors.Code("ASSIGNMENT_EXPIRED")

	// Permission errors
	ErrCodePermissionInvalid = errors.Code("PERMISSION_INVALID")
	ErrCodePermissionDenied  = errors.Code("PERMISSION_DENIED")

	// Action/Resource errors
	ErrCodeActionInvalid   = errors.Code("ACTION_INVALID")
	ErrCodeResourceInvalid = errors.Code("RESOURCE_INVALID")
)

// Register domain error codes with the global registry.
// This happens once when the package is imported.
func init() {
	// Role errors
	errors.Register(
		ErrCodeRoleNotFound,
		errors.KindNotFound,
		"role not found",
	)

	errors.Register(
		ErrCodeRoleAlreadyExists,
		errors.KindConflict,
		"role already exists",
	)

	errors.Register(
		ErrCodeRoleIsSystem,
		errors.KindForbidden,
		"cannot modify system role",
	)

	errors.Register(
		ErrCodeRoleNameInvalid,
		errors.KindValidation,
		"invalid role name",
	)

	// Assignment errors
	errors.Register(
		ErrCodeAssignmentNotFound,
		errors.KindNotFound,
		"role assignment not found",
	)

	errors.Register(
		ErrCodeAssignmentExists,
		errors.KindConflict,
		"role assignment already exists",
	)

	errors.Register(
		ErrCodeAssignmentExpired,
		errors.KindForbidden,
		"role assignment has expired",
	)

	// Permission errors
	errors.Register(
		ErrCodePermissionInvalid,
		errors.KindValidation,
		"invalid permission",
	)

	errors.Register(
		ErrCodePermissionDenied,
		errors.KindForbidden,
		"permission denied",
	)

	// Action/Resource errors
	errors.Register(
		ErrCodeActionInvalid,
		errors.KindValidation,
		"invalid action",
	)

	errors.Register(
		ErrCodeResourceInvalid,
		errors.KindValidation,
		"invalid resource",
	)
}

// ============================================================================
// Error Constructor Helpers
// ============================================================================

// ErrRoleNotFound creates a role not found error.
func ErrRoleNotFound(operation string, roleID string) error {
	return errors.NotFound(operation, "role").
		WithCode(ErrCodeRoleNotFound).
		WithMeta("role_id", roleID)
}

// ErrRoleNotFoundByName creates a role not found by name error.
func ErrRoleNotFoundByName(operation string, name string) error {
	return errors.NotFound(operation, "role").
		WithCode(ErrCodeRoleNotFound).
		WithMeta("role_name", name)
}

// ErrRoleAlreadyExists creates a role already exists error.
func ErrRoleAlreadyExists(operation string, name string) error {
	return errors.Conflict(operation, "role").
		WithCode(ErrCodeRoleAlreadyExists).
		WithMeta("role_name", name)
}

// ErrRoleIsSystem creates a system role modification error.
func ErrRoleIsSystem(operation string, name string) error {
	return errors.Forbidden(operation, "cannot modify system role").
		WithCode(ErrCodeRoleIsSystem).
		WithMeta("role_name", name)
}

// ErrRoleNameInvalid creates a role name invalid error.
func ErrRoleNameInvalid(operation string, name string, reason string) error {
	return errors.Validation(operation, "invalid role name").
		WithCode(ErrCodeRoleNameInvalid).
		WithMeta("role_name", name).
		WithMeta("reason", reason)
}

// ErrAssignmentNotFound creates an assignment not found error.
func ErrAssignmentNotFound(operation string, assignmentID string) error {
	return errors.NotFound(operation, "assignment").
		WithCode(ErrCodeAssignmentNotFound).
		WithMeta("assignment_id", assignmentID)
}

// ErrAssignmentExists creates an assignment already exists error.
func ErrAssignmentExists(operation string, userID string, roleID string) error {
	return errors.Conflict(operation, "assignment").
		WithCode(ErrCodeAssignmentExists).
		WithMeta("user_id", userID).
		WithMeta("role_id", roleID)
}

// ErrAssignmentExpired creates an assignment expired error.
func ErrAssignmentExpired(operation string, assignmentID string) error {
	return errors.Forbidden(operation, "assignment expired").
		WithCode(ErrCodeAssignmentExpired).
		WithMeta("assignment_id", assignmentID)
}

// ErrPermissionInvalid creates a permission invalid error.
func ErrPermissionInvalid(operation string, permission string, reason string) error {
	return errors.Validation(operation, "invalid permission").
		WithCode(ErrCodePermissionInvalid).
		WithMeta("permission", permission).
		WithMeta("reason", reason)
}

// ErrPermissionDenied creates a permission denied error.
func ErrPermissionDenied(operation string, action string, resource string) error {
	return errors.Forbidden(operation, "permission denied").
		WithCode(ErrCodePermissionDenied).
		WithMeta("action", action).
		WithMeta("resource", resource)
}

// ErrActionInvalid creates an invalid action error.
func ErrActionInvalid(operation string, action string) error {
	return errors.Validation(operation, "invalid action").
		WithCode(ErrCodeActionInvalid).
		WithMeta("action", action)
}

// ErrResourceInvalid creates an invalid resource error.
func ErrResourceInvalid(operation string, resource string) error {
	return errors.Validation(operation, "invalid resource").
		WithCode(ErrCodeResourceInvalid).
		WithMeta("resource", resource)
}
