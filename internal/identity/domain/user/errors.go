package user

import (
	"github.com/0xsj/hexagonal-go/pkg/errors"
)

// Domain-specific error codes for the User aggregate.
// These codes are stable across versions and used by clients.
//
// Naming convention: USER_RESOURCE_CONDITION
const (
	// User not found errors
	ErrCodeUserNotFound = errors.Code("USER_NOT_FOUND")

	// Email-related errors
	ErrCodeEmailAlreadyTaken        = errors.Code("USER_EMAIL_TAKEN")
	ErrCodeEmailNotVerified         = errors.Code("USER_EMAIL_NOT_VERIFIED")
	ErrCodeEmailInvalid             = errors.Code("USER_EMAIL_INVALID")
	ErrCodeEmailVerificationExpired = errors.Code("USER_EMAIL_VERIFICATION_EXPIRED")

	// Password-related errors
	ErrCodePasswordInvalid = errors.Code("USER_PASSWORD_INVALID")
	ErrCodePasswordTooWeak = errors.Code("USER_PASSWORD_TOO_WEAK")
	ErrCodePasswordExpired = errors.Code("USER_PASSWORD_EXPIRED")

	// Account status errors
	ErrCodeAccountLocked    = errors.Code("USER_ACCOUNT_LOCKED")
	ErrCodeAccountSuspended = errors.Code("USER_ACCOUNT_SUSPENDED")
	ErrCodeAccountDeleted   = errors.Code("USER_ACCOUNT_DELETED")

	// Authentication errors
	ErrCodeInvalidCredentials = errors.Code("USER_INVALID_CREDENTIALS")
	ErrCodeTooManyAttempts    = errors.Code("USER_TOO_MANY_LOGIN_ATTEMPTS")

	// Authorization errors
	ErrCodeInsufficientRole = errors.Code("USER_INSUFFICIENT_ROLE")
)

// Register domain error codes with the global registry.
// This happens once when the package is imported.
func init() {
	// User not found
	errors.Register(
		ErrCodeUserNotFound,
		errors.KindNotFound,
		"user not found",
	)

	// Email errors
	errors.Register(
		ErrCodeEmailAlreadyTaken,
		errors.KindConflict,
		"email address is already registered",
	)

	errors.Register(
		ErrCodeEmailNotVerified,
		errors.KindForbidden,
		"email address not verified",
	)

	errors.Register(
		ErrCodeEmailInvalid,
		errors.KindValidation,
		"invalid email address",
	)

	errors.Register(
		ErrCodeEmailVerificationExpired,
		errors.KindForbidden,
		"email verification link has expired",
	)

	// Password errors
	errors.Register(
		ErrCodePasswordInvalid,
		errors.KindValidation,
		"invalid password",
	)

	errors.Register(
		ErrCodePasswordTooWeak,
		errors.KindValidation,
		"password does not meet security requirements",
	)

	errors.Register(
		ErrCodePasswordExpired,
		errors.KindForbidden,
		"password has expired and must be changed",
	)

	// Account status errors
	errors.Register(
		ErrCodeAccountLocked,
		errors.KindForbidden,
		"account is locked due to multiple failed login attempts",
	)

	errors.Register(
		ErrCodeAccountSuspended,
		errors.KindForbidden,
		"account has been suspended",
	)

	errors.Register(
		ErrCodeAccountDeleted,
		errors.KindNotFound,
		"account has been deleted",
	)

	// Authentication errors
	errors.Register(
		ErrCodeInvalidCredentials,
		errors.KindUnauthorized,
		"invalid email or password",
	)

	errors.Register(
		ErrCodeTooManyAttempts,
		errors.KindRateLimit,
		"too many login attempts, account temporarily locked",
	)

	// Authorization errors
	errors.Register(
		ErrCodeInsufficientRole,
		errors.KindForbidden,
		"user role does not have sufficient permissions",
	)
}

// ============================================================================
// Error Constructor Helpers
// ============================================================================

// ErrUserNotFound creates a user not found error.
func ErrUserNotFound(operation string, userID string) error {
	return errors.NotFound(operation, "user").
		WithCode(ErrCodeUserNotFound).
		WithMeta("user_id", userID)
}

// ErrEmailAlreadyTaken creates an email already taken error.
func ErrEmailAlreadyTaken(operation string, email string) error {
	return errors.Conflict(operation, "email").
		WithCode(ErrCodeEmailAlreadyTaken).
		WithMeta("email", email)
}

// ErrEmailNotVerified creates an email not verified error.
func ErrEmailNotVerified(operation string, email string) error {
	return errors.Forbidden(operation, "email not verified").
		WithCode(ErrCodeEmailNotVerified).
		WithMeta("email", email)
}

// ErrPasswordInvalid creates a password invalid error.
func ErrPasswordInvalid(operation string) error {
	return errors.Validation(operation, "invalid password").
		WithCode(ErrCodePasswordInvalid)
}

// ErrPasswordTooWeak creates a password too weak error.
func ErrPasswordTooWeak(operation string, requirements string) error {
	return errors.Validation(operation, "password too weak").
		WithCode(ErrCodePasswordTooWeak).
		WithMeta("requirements", requirements)
}

// ErrAccountLocked creates an account locked error.
func ErrAccountLocked(operation string, lockedUntil string) error {
	return errors.Forbidden(operation, "account locked").
		WithCode(ErrCodeAccountLocked).
		WithMeta("locked_until", lockedUntil)
}

// ErrAccountSuspended creates an account suspended error.
func ErrAccountSuspended(operation string, reason string) error {
	return errors.Forbidden(operation, "account suspended").
		WithCode(ErrCodeAccountSuspended).
		WithMeta("reason", reason)
}

// ErrInvalidCredentials creates an invalid credentials error.
func ErrInvalidCredentials(operation string) error {
	return errors.Unauthorized(operation, "invalid email or password").
		WithCode(ErrCodeInvalidCredentials)
}

// ErrTooManyAttempts creates a too many attempts error.
func ErrTooManyAttempts(operation string, retryAfter string) error {
	return errors.RateLimit(operation, "too many login attempts").
		WithCode(ErrCodeTooManyAttempts).
		WithMeta("retry_after", retryAfter)
}

// ErrInsufficientRole creates an insufficient role error.
func ErrInsufficientRole(operation string, requiredRole, currentRole string) error {
	return errors.Forbidden(operation, "insufficient role").
		WithCode(ErrCodeInsufficientRole).
		WithMeta("required_role", requiredRole).
		WithMeta("current_role", currentRole)
}

// ErrEmailInvalid creates an email invalid error.
func ErrEmailInvalid(operation string, email string) error {
	return errors.Validation(operation, "invalid email address").
		WithCode(ErrCodeEmailInvalid).
		WithMeta("email", email)
}
