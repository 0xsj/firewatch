package tenant

import (
	"github.com/0xsj/hexagonal-go/pkg/errors"
)

// Domain-specific error codes for the Tenant aggregate.
// These codes are stable across versions and used by clients.
//
// Naming convention: TENANT_RESOURCE_CONDITION
const (
	// Tenant not found errors
	ErrCodeTenantNotFound = errors.Code("TENANT_NOT_FOUND")

	// Slug-related errors
	ErrCodeSlugAlreadyTaken = errors.Code("TENANT_SLUG_TAKEN")
	ErrCodeSlugInvalid      = errors.Code("TENANT_SLUG_INVALID")
	ErrCodeSlugReserved     = errors.Code("TENANT_SLUG_RESERVED")

	// Status-related errors
	ErrCodeTenantSuspended     = errors.Code("TENANT_SUSPENDED")
	ErrCodeTenantDeleted       = errors.Code("TENANT_DELETED")
	ErrCodeInvalidStatusChange = errors.Code("TENANT_INVALID_STATUS_CHANGE")

	// Plan-related errors
	ErrCodePlanInvalid   = errors.Code("TENANT_PLAN_INVALID")
	ErrCodePlanDowngrade = errors.Code("TENANT_PLAN_DOWNGRADE_RESTRICTED")

	// Settings-related errors
	ErrCodeSettingsInvalid = errors.Code("TENANT_SETTINGS_INVALID")

	// General errors
	ErrCodeTenantNameInvalid = errors.Code("TENANT_NAME_INVALID")
)

// Register domain error codes with the global registry.
// This happens once when the package is imported.
func init() {
	// Tenant not found
	errors.Register(
		ErrCodeTenantNotFound,
		errors.KindNotFound,
		"tenant not found",
	)

	// Slug errors
	errors.Register(
		ErrCodeSlugAlreadyTaken,
		errors.KindConflict,
		"tenant slug is already taken",
	)

	errors.Register(
		ErrCodeSlugInvalid,
		errors.KindValidation,
		"invalid tenant slug format",
	)

	errors.Register(
		ErrCodeSlugReserved,
		errors.KindValidation,
		"tenant slug is reserved",
	)

	// Status errors
	errors.Register(
		ErrCodeTenantSuspended,
		errors.KindForbidden,
		"tenant has been suspended",
	)

	errors.Register(
		ErrCodeTenantDeleted,
		errors.KindNotFound,
		"tenant has been deleted",
	)

	errors.Register(
		ErrCodeInvalidStatusChange,
		errors.KindDomain,
		"invalid tenant status transition",
	)

	// Plan errors
	errors.Register(
		ErrCodePlanInvalid,
		errors.KindValidation,
		"invalid tenant plan",
	)

	errors.Register(
		ErrCodePlanDowngrade,
		errors.KindDomain,
		"plan downgrade not permitted",
	)

	// Settings errors
	errors.Register(
		ErrCodeSettingsInvalid,
		errors.KindValidation,
		"invalid tenant settings",
	)

	// General errors
	errors.Register(
		ErrCodeTenantNameInvalid,
		errors.KindValidation,
		"invalid tenant name",
	)
}

// ============================================================================
// Error Constructor Helpers
// ============================================================================

// ErrTenantNotFound creates a tenant not found error.
func ErrTenantNotFound(operation string, tenantID string) error {
	return errors.NotFound(operation, "tenant").
		WithCode(ErrCodeTenantNotFound).
		WithMeta("tenant_id", tenantID)
}

// ErrTenantNotFoundBySlug creates a tenant not found error by slug.
func ErrTenantNotFoundBySlug(operation string, slug string) error {
	return errors.NotFound(operation, "tenant").
		WithCode(ErrCodeTenantNotFound).
		WithMeta("slug", slug)
}

// ErrSlugAlreadyTaken creates a slug already taken error.
func ErrSlugAlreadyTaken(operation string, slug string) error {
	return errors.Conflict(operation, "slug").
		WithCode(ErrCodeSlugAlreadyTaken).
		WithMeta("slug", slug)
}

// ErrSlugInvalid creates a slug invalid error.
func ErrSlugInvalid(operation string, slug string, reason string) error {
	return errors.Validation(operation, "invalid slug format").
		WithCode(ErrCodeSlugInvalid).
		WithMeta("slug", slug).
		WithMeta("reason", reason)
}

// ErrSlugReserved creates a slug reserved error.
func ErrSlugReserved(operation string, slug string) error {
	return errors.Validation(operation, "slug is reserved").
		WithCode(ErrCodeSlugReserved).
		WithMeta("slug", slug)
}

// ErrTenantSuspended creates a tenant suspended error.
func ErrTenantSuspended(operation string, tenantID string, reason string) error {
	return errors.Forbidden(operation, "tenant suspended").
		WithCode(ErrCodeTenantSuspended).
		WithMeta("tenant_id", tenantID).
		WithMeta("reason", reason)
}

// ErrTenantDeleted creates a tenant deleted error.
func ErrTenantDeleted(operation string, tenantID string) error {
	return errors.NotFound(operation, "tenant").
		WithCode(ErrCodeTenantDeleted).
		WithMeta("tenant_id", tenantID)
}

// ErrInvalidStatusChange creates an invalid status transition error.
func ErrInvalidStatusChange(operation string, from string, to string) error {
	return errors.Domain(operation, "invalid status transition").
		WithCode(ErrCodeInvalidStatusChange).
		WithMeta("from_status", from).
		WithMeta("to_status", to)
}

// ErrPlanInvalid creates a plan invalid error.
func ErrPlanInvalid(operation string, plan string) error {
	return errors.Validation(operation, "invalid plan").
		WithCode(ErrCodePlanInvalid).
		WithMeta("plan", plan)
}

// ErrPlanDowngrade creates a plan downgrade restricted error.
func ErrPlanDowngrade(operation string, from string, to string, reason string) error {
	return errors.Domain(operation, "plan downgrade not permitted").
		WithCode(ErrCodePlanDowngrade).
		WithMeta("from_plan", from).
		WithMeta("to_plan", to).
		WithMeta("reason", reason)
}

// ErrSettingsInvalid creates a settings invalid error.
func ErrSettingsInvalid(operation string, reason string) error {
	return errors.Validation(operation, "invalid settings").
		WithCode(ErrCodeSettingsInvalid).
		WithMeta("reason", reason)
}

// ErrTenantNameInvalid creates a tenant name invalid error.
func ErrTenantNameInvalid(operation string, name string, reason string) error {
	return errors.Validation(operation, "invalid tenant name").
		WithCode(ErrCodeTenantNameInvalid).
		WithMeta("name", name).
		WithMeta("reason", reason)
}
