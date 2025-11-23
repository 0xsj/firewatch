package errors

import "errors"

// Sentinel errors for fast error checking using errors.Is().
//
// These are base errors that can be wrapped with additional context.
// They allow zero-allocation error checks in hot paths.
var (
	ErrNotFound       = errors.New("not found")
	ErrValidation     = errors.New("validation failed")
	ErrConflict       = errors.New("conflict")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrDomain         = errors.New("domain error")
	ErrInfrastructure = errors.New("infrastructure error")
	ErrTimeout        = errors.New("timeout")
	ErrRateLimit      = errors.New("rate limit exceeded")
	ErrInternal       = errors.New("internal error")
)

// sentinelForKind returns the sentinel error for a given Kind.
// This is used internally when creating errors to preserve sentinel identity.
func sentinelForKind(kind Kind) error {
	switch kind {
	case KindNotFound:
		return ErrNotFound
	case KindValidation:
		return ErrValidation
	case KindConflict:
		return ErrConflict
	case KindUnauthorized:
		return ErrUnauthorized
	case KindForbidden:
		return ErrForbidden
	case KindDomain:
		return ErrDomain
	case KindInfrastructure:
		return ErrInfrastructure
	case KindTimeout:
		return ErrTimeout
	case KindRateLimit:
		return ErrRateLimit
	case KindInternal:
		return ErrInternal
	default:
		return nil
	}
}
