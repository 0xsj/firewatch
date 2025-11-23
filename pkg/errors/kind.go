// Package errors provides a comprehensive error handling system with kinds,
// codes, metadata, and protocol-specific mappings for Go applications.
package errors

// Kind represents the category of an error.
// Kinds map to standard protocol status codes (HTTP, gRPC, GraphQL).
type Kind uint8

const (
	// KindOther represents an unknown or uncategorized error.
	// This is the zero value and should be avoided in production code.
	// Maps to: HTTP 500, gRPC Unknown, GraphQL INTERNAL_ERROR
	KindOther Kind = iota

	// KindNotFound indicates a requested resource does not exist.
	// Use when: User requests entity by ID that doesn't exist.
	// Maps to: HTTP 404, gRPC NotFound, GraphQL NOT_FOUND
	KindNotFound

	// KindValidation indicates invalid input from the client.
	// Use when: Request fails validation rules (format, length, type).
	// Maps to: HTTP 400, gRPC InvalidArgument, GraphQL BAD_USER_INPUT
	KindValidation

	// KindConflict indicates a resource conflict or version mismatch.
	// Use when: Creating resource that already exists, optimistic locking fails.
	// Maps to: HTTP 409, gRPC AlreadyExists/FailedPrecondition, GraphQL CONFLICT
	KindConflict

	// KindUnauthorized indicates missing or invalid authentication.
	// Use when: No auth token, invalid token, expired token.
	// Maps to: HTTP 401, gRPC Unauthenticated, GraphQL UNAUTHENTICATED
	KindUnauthorized

	// KindForbidden indicates insufficient permissions.
	// Use when: Authenticated but not authorized for this resource/action.
	// Maps to: HTTP 403, gRPC PermissionDenied, GraphQL FORBIDDEN
	KindForbidden

	// KindDomain indicates a business rule violation.
	// Use when: Request is valid but violates domain logic.
	// Example: Cannot cancel shipped order, insufficient balance.
	// Maps to: HTTP 422, gRPC FailedPrecondition, GraphQL BUSINESS_RULE_VIOLATION
	KindDomain

	// KindInfrastructure indicates an external system failure.
	// Use when: Database error, external API failure, cache unavailable.
	// Maps to: HTTP 503, gRPC Unavailable, GraphQL SERVICE_UNAVAILABLE
	KindInfrastructure

	// KindTimeout indicates an operation exceeded its deadline.
	// Use when: Request times out, deadline exceeded, context cancelled.
	// Maps to: HTTP 504, gRPC DeadlineExceeded, GraphQL TIMEOUT
	KindTimeout

	// KindRateLimit indicates too many requests from client.
	// Use when: Rate limit exceeded, quota exhausted, throttling applied.
	// Maps to: HTTP 429, gRPC ResourceExhausted, GraphQL RATE_LIMITED
	KindRateLimit

	// KindInternal indicates an unexpected internal error.
	// Use when: Panic recovered, assertion failed, unexpected state.
	// This is a catch-all for bugs and should trigger alerts.
	// Maps to: HTTP 500, gRPC Internal, GraphQL INTERNAL_ERROR
	KindInternal
)

// String returns the human-readable name of the error kind.
// This is used for logging, metrics, and error responses.
func (k Kind) String() string {
	switch k {
	case KindNotFound:
		return "not_found"
	case KindValidation:
		return "validation"
	case KindConflict:
		return "conflict"
	case KindUnauthorized:
		return "unauthorized"
	case KindForbidden:
		return "forbidden"
	case KindDomain:
		return "domain"
	case KindInfrastructure:
		return "infrastructure"
	case KindTimeout:
		return "timeout"
	case KindRateLimit:
		return "rate_limit"
	case KindInternal:
		return "internal"
	default:
		return "other"
	}
}

// IsClientError returns true if the error kind represents a client-side error
// that should not be retried without modification.
func (k Kind) IsClientError() bool {
	switch k {
	case KindValidation, KindUnauthorized, KindForbidden, KindNotFound, KindConflict, KindDomain:
		return true
	default:
		return false
	}
}

// IsServerError returns true if the error kind represents a server-side error
// that might be retryable or indicates an infrastructure problem.
func (k Kind) IsServerError() bool {
	switch k {
	case KindInfrastructure, KindTimeout, KindInternal:
		return true
	default:
		return false
	}
}

// IsRetryableByDefault returns true if errors of this kind are typically safe
// to retry. Note: Individual errors can override this with the Retryable field.
func (k Kind) IsRetryableByDefault() bool {
	switch k {
	case KindInfrastructure, KindTimeout, KindRateLimit:
		return true
	default:
		return false
	}
}
