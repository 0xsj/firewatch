package errors

import (
	"errors"
	"fmt"
)

// Error is a structured error with rich context for production applications.
//
// It provides:
//   - Kind: Category for protocol mapping (HTTP status, gRPC code, etc.)
//   - Code: Machine-readable identifier for client error handling
//   - Operation: Context about where the error occurred
//   - Message: Human-readable description
//   - Metadata: Additional structured context (user IDs, resource IDs, etc.)
//   - Severity: Error severity level for logging and alerting
//   - Retryable: Whether the operation can be safely retried
//   - Err: Wrapped underlying error for error chain support
//
// Error implements the error interface and supports errors.Is/As/Unwrap.
type Error struct {
	// Kind categorizes the error for protocol mapping.
	// Used to determine HTTP status codes, gRPC codes, etc.
	Kind Kind

	// Code is a machine-readable error identifier.
	// Examples: USER_EMAIL_TAKEN, ORDER_ALREADY_SHIPPED
	// Used by clients for programmatic error handling.
	Code Code

	// Operation identifies where the error occurred.
	// Format: "Package.Type.Method" or "Service.Method"
	// Examples: "users.Repository.Create", "Handler.ProcessPayment"
	Operation string

	// Message is a human-readable error description.
	// Should be clear and actionable for developers.
	// Avoid including sensitive information (passwords, tokens, etc.)
	Message string

	// Metadata contains additional structured context.
	// Use for: user IDs, resource IDs, field names, validation errors
	// Avoid: sensitive data, large objects
	Metadata map[string]any

	// Severity indicates the error severity level.
	// Used for: log level mapping, alerting decisions
	Severity Severity

	// Retryable indicates if the operation can be safely retried.
	// True for: timeouts, infrastructure errors, rate limits
	// False for: validation errors, not found, conflicts
	Retryable bool

	// Err is the wrapped underlying error.
	// Used for error chain support with errors.Is/As/Unwrap.
	// Typically wraps a sentinel error for fast checking.
	Err error
}

// Error implements the error interface.
// Returns a formatted error message with operation context.
func (e *Error) Error() string {
	if e.Operation != "" {
		return fmt.Sprintf("%s: %s", e.Operation, e.Message)
	}
	return e.Message
}

// Unwrap returns the wrapped error for errors.Is/As support.
// This allows error chain walking with the stdlib errors package.
func (e *Error) Unwrap() error {
	return e.Err
}

// ============================================================================
// Constructors - Create errors by Kind
// ============================================================================

// New creates a new Error with the specified kind, code, and message.
// This is the base constructor - prefer using specific constructors like
// NotFound(), Validation(), etc. for better clarity.
func New(kind Kind, code Code, message string) *Error {
	return &Error{
		Kind:      kind,
		Code:      code,
		Message:   message,
		Severity:  SeverityError,
		Retryable: kind.IsRetryableByDefault(),
		Err:       sentinelForKind(kind),
	}
}

// NotFound creates a not found error (404, gRPC NotFound).
// Use when a requested resource doesn't exist.
//
// Example:
//
//	return errors.NotFound("users.Repository.FindByID", "user")
func NotFound(operation, resource string) *Error {
	return &Error{
		Kind:      KindNotFound,
		Code:      CodeNotFound,
		Operation: operation,
		Message:   fmt.Sprintf("%s not found", resource),
		Severity:  SeverityError,
		Retryable: false,
		Err:       ErrNotFound,
	}
}

// NotFoundf creates a not found error with formatted message.
func NotFoundf(operation, format string, args ...any) *Error {
	return &Error{
		Kind:      KindNotFound,
		Code:      CodeNotFound,
		Operation: operation,
		Message:   fmt.Sprintf(format, args...),
		Severity:  SeverityError,
		Retryable: false,
		Err:       ErrNotFound,
	}
}

// Validation creates a validation error (400, gRPC InvalidArgument).
// Use when client input fails validation rules.
//
// Example:
//
//	return errors.Validation("users.Handler.Create", "invalid email format")
func Validation(operation, message string) *Error {
	return &Error{
		Kind:      KindValidation,
		Code:      CodeInvalidInput,
		Operation: operation,
		Message:   message,
		Severity:  SeverityError,
		Retryable: false,
		Err:       ErrValidation,
	}
}

// Validationf creates a validation error with formatted message.
func Validationf(operation, format string, args ...any) *Error {
	return &Error{
		Kind:      KindValidation,
		Code:      CodeInvalidInput,
		Operation: operation,
		Message:   fmt.Sprintf(format, args...),
		Severity:  SeverityError,
		Retryable: false,
		Err:       ErrValidation,
	}
}

// Conflict creates a conflict error (409, gRPC AlreadyExists).
// Use when a resource already exists or version conflict occurs.
//
// Example:
//
//	return errors.Conflict("users.Repository.Create", "email")
func Conflict(operation, resource string) *Error {
	return &Error{
		Kind:      KindConflict,
		Code:      CodeAlreadyExists,
		Operation: operation,
		Message:   fmt.Sprintf("%s already exists", resource),
		Severity:  SeverityError,
		Retryable: false,
		Err:       ErrConflict,
	}
}

// Conflictf creates a conflict error with formatted message.
func Conflictf(operation, format string, args ...any) *Error {
	return &Error{
		Kind:      KindConflict,
		Code:      CodeConflict,
		Operation: operation,
		Message:   fmt.Sprintf(format, args...),
		Severity:  SeverityError,
		Retryable: false,
		Err:       ErrConflict,
	}
}

// Unauthorized creates an unauthorized error (401, gRPC Unauthenticated).
// Use when authentication is required or credentials are invalid.
//
// Example:
//
//	return errors.Unauthorized("auth.Middleware.Authenticate", "missing token")
func Unauthorized(operation, message string) *Error {
	return &Error{
		Kind:      KindUnauthorized,
		Code:      CodeUnauthenticated,
		Operation: operation,
		Message:   message,
		Severity:  SeverityError,
		Retryable: false,
		Err:       ErrUnauthorized,
	}
}

// Unauthorizedf creates an unauthorized error with formatted message.
func Unauthorizedf(operation, format string, args ...any) *Error {
	return &Error{
		Kind:      KindUnauthorized,
		Code:      CodeUnauthenticated,
		Operation: operation,
		Message:   fmt.Sprintf(format, args...),
		Severity:  SeverityError,
		Retryable: false,
		Err:       ErrUnauthorized,
	}
}

// Forbidden creates a forbidden error (403, gRPC PermissionDenied).
// Use when the authenticated user lacks permission for the operation.
//
// Example:
//
//	return errors.Forbidden("orders.Service.Cancel", "insufficient permissions")
func Forbidden(operation, message string) *Error {
	return &Error{
		Kind:      KindForbidden,
		Code:      CodeUnauthorized,
		Operation: operation,
		Message:   message,
		Severity:  SeverityError,
		Retryable: false,
		Err:       ErrForbidden,
	}
}

// Forbiddenf creates a forbidden error with formatted message.
func Forbiddenf(operation, format string, args ...any) *Error {
	return &Error{
		Kind:      KindForbidden,
		Code:      CodeUnauthorized,
		Operation: operation,
		Message:   fmt.Sprintf(format, args...),
		Severity:  SeverityError,
		Retryable: false,
		Err:       ErrForbidden,
	}
}

// Domain creates a domain error (422, gRPC FailedPrecondition).
// Use when a business rule is violated.
//
// Example:
//
//	return errors.Domain("orders.Service.Cancel", "cannot cancel shipped order")
func Domain(operation, message string) *Error {
	return &Error{
		Kind:      KindDomain,
		Code:      CodeConflict,
		Operation: operation,
		Message:   message,
		Severity:  SeverityError,
		Retryable: false,
		Err:       ErrDomain,
	}
}

// Domainf creates a domain error with formatted message.
func Domainf(operation, format string, args ...any) *Error {
	return &Error{
		Kind:      KindDomain,
		Code:      CodeConflict,
		Operation: operation,
		Message:   fmt.Sprintf(format, args...),
		Severity:  SeverityError,
		Retryable: false,
		Err:       ErrDomain,
	}
}

// Infrastructure creates an infrastructure error (503, gRPC Unavailable).
// Use when an external system (database, API, cache) fails.
//
// Example:
//
//	return errors.Infrastructure("users.Repository.Create", dbErr)
func Infrastructure(operation string, err error) *Error {
	message := "infrastructure error"
	if err != nil {
		message = err.Error()
	}

	return &Error{
		Kind:      KindInfrastructure,
		Code:      CodeUnavailable,
		Operation: operation,
		Message:   message,
		Severity:  SeverityError,
		Retryable: true,
		Err:       fmt.Errorf("%w: %v", ErrInfrastructure, err),
	}
}

// Timeout creates a timeout error (504, gRPC DeadlineExceeded).
// Use when an operation exceeds its deadline.
//
// Example:
//
//	return errors.Timeout("api.Client.Request", "request took too long")
func Timeout(operation, message string) *Error {
	return &Error{
		Kind:      KindTimeout,
		Code:      CodeTimeout,
		Operation: operation,
		Message:   message,
		Severity:  SeverityError,
		Retryable: true,
		Err:       ErrTimeout,
	}
}

// Timeoutf creates a timeout error with formatted message.
func Timeoutf(operation, format string, args ...any) *Error {
	return &Error{
		Kind:      KindTimeout,
		Code:      CodeTimeout,
		Operation: operation,
		Message:   fmt.Sprintf(format, args...),
		Severity:  SeverityError,
		Retryable: true,
		Err:       ErrTimeout,
	}
}

// RateLimit creates a rate limit error (429, gRPC ResourceExhausted).
// Use when rate limits or quotas are exceeded.
//
// Example:
//
//	return errors.RateLimit("api.Middleware.CheckRate", "too many requests")
func RateLimit(operation, message string) *Error {
	return &Error{
		Kind:      KindRateLimit,
		Code:      CodeRateLimitExceeded,
		Operation: operation,
		Message:   message,
		Severity:  SeverityError,
		Retryable: true,
		Err:       ErrRateLimit,
	}
}

// RateLimitf creates a rate limit error with formatted message.
func RateLimitf(operation, format string, args ...any) *Error {
	return &Error{
		Kind:      KindRateLimit,
		Code:      CodeRateLimitExceeded,
		Operation: operation,
		Message:   fmt.Sprintf(format, args...),
		Severity:  SeverityError,
		Retryable: true,
		Err:       ErrRateLimit,
	}
}

// Internal creates an internal error (500, gRPC Internal).
// Use for unexpected errors, bugs, or panics.
// This should trigger alerts in production.
//
// Example:
//
//	return errors.Internal("service.Process", unexpectedErr)
func Internal(operation string, err error) *Error {
	message := "internal error"
	if err != nil {
		message = err.Error()
	}

	return &Error{
		Kind:      KindInternal,
		Code:      CodeInternal,
		Operation: operation,
		Message:   message,
		Severity:  SeverityError,
		Retryable: false,
		Err:       fmt.Errorf("%w: %v", ErrInternal, err),
	}
}

// Internalf creates an internal error with formatted message.
func Internalf(operation, format string, args ...any) *Error {
	return &Error{
		Kind:      KindInternal,
		Code:      CodeInternal,
		Operation: operation,
		Message:   fmt.Sprintf(format, args...),
		Severity:  SeverityError,
		Retryable: false,
		Err:       ErrInternal,
	}
}

// ============================================================================
// Wrapping - Preserve error context while adding operation
// ============================================================================

// Wrap wraps an error with additional operation context.
// If err is already an *Error, it preserves Kind, Code, and Metadata,
// but updates the Operation to reflect the current context.
//
// If err is a standard error, it wraps it as KindInternal.
//
// Example:
//
//	user, err := repo.FindByID(ctx, id)
//	if err != nil {
//	    return errors.Wrap(err, "users.Service.GetUser")
//	}
func Wrap(err error, operation string) error {
	if err == nil {
		return nil
	}

	// If already our Error type, preserve kind and update operation
	var appErr *Error
	if As(err, &appErr) {
		return &Error{
			Kind:      appErr.Kind,
			Code:      appErr.Code,
			Operation: operation, // Update operation to current context
			Message:   appErr.Message,
			Metadata:  appErr.Metadata,
			Severity:  appErr.Severity,
			Retryable: appErr.Retryable,
			Err:       appErr.Err, // Preserve wrapped error
		}
	}

	// Standard error - wrap as Internal
	return &Error{
		Kind:      KindInternal,
		Code:      CodeInternal,
		Operation: operation,
		Message:   err.Error(),
		Severity:  SeverityError,
		Retryable: false,
		Err:       fmt.Errorf("%w: %v", ErrInternal, err),
	}
}

// Wrapf wraps an error with formatted operation and message.
func Wrapf(err error, operation, format string, args ...any) error {
	if err == nil {
		return nil
	}

	wrapped := Wrap(err, operation).(*Error)
	wrapped.Message = fmt.Sprintf(format, args...)
	return wrapped
}

// ============================================================================
// Builder Methods - Fluent API for enriching errors
// ============================================================================

// WithCode sets a custom error code.
// Use this to override the default code with a domain-specific one.
//
// Example:
//
//	return errors.Conflict(op, "email").
//	    WithCode("USER_EMAIL_TAKEN")
func (e *Error) WithCode(code Code) *Error {
	e.Code = code
	return e
}

// WithOperation sets the operation context.
func (e *Error) WithOperation(operation string) *Error {
	e.Operation = operation
	return e
}

// WithMessage sets a custom message.
func (e *Error) WithMessage(message string) *Error {
	e.Message = message
	return e
}

// WithMessagef sets a formatted message.
func (e *Error) WithMessagef(format string, args ...any) *Error {
	e.Message = fmt.Sprintf(format, args...)
	return e
}

// WithMeta adds a metadata key-value pair.
// Metadata is merged with existing metadata.
//
// Example:
//
//	return errors.NotFound(op, "user").
//	    WithMeta("user_id", userID).
//	    WithMeta("email", email)
func (e *Error) WithMeta(key string, value any) *Error {
	if e.Metadata == nil {
		e.Metadata = make(map[string]any)
	}
	e.Metadata[key] = value
	return e
}

// WithMetadata replaces all metadata with the provided map.
func (e *Error) WithMetadata(metadata map[string]any) *Error {
	e.Metadata = metadata
	return e
}

// WithSeverity sets the error severity level.
//
// Example:
//
//	return errors.Internal(op, err).
//	    WithSeverity(errors.SeverityWarning) // Not critical
func (e *Error) WithSeverity(severity Severity) *Error {
	e.Severity = severity
	return e
}

// WithRetryable sets whether the error is retryable.
//
// Example:
//
//	return errors.Domain(op, "insufficient balance").
//	    WithRetryable(true) // Can retry after user adds funds
func (e *Error) WithRetryable(retryable bool) *Error {
	e.Retryable = retryable
	return e
}

// WithKind sets a custom error kind.
// Use this carefully - prefer using the correct constructor instead.
func (e *Error) WithKind(kind Kind) *Error {
	e.Kind = kind
	e.Err = sentinelForKind(kind) // Update wrapped sentinel
	return e
}

// WithEntity adds entity type and ID to metadata.
// Useful for domain errors related to specific entities.
//
// Example:
//
//	return errors.NotFound(op, "user").
//	    WithEntity("user", userID)
func (e *Error) WithEntity(entityType string, entityID any) *Error {
	return e.
		WithMeta("entity_type", entityType).
		WithMeta("entity_id", entityID)
}

// WithTenant adds tenant ID to metadata.
// For multi-tenancy, this helps track which tenant encountered the error.
//
// Example:
//
//	return errors.NotFound(op, "user").
//	    WithTenant(tenantID)
func (e *Error) WithTenant(tenantID string) *Error {
	return e.WithMeta("tenant_id", tenantID)
}

// ============================================================================
// Type Assertions
// ============================================================================

// As is a convenience wrapper around errors.As.
// It checks if err is or wraps an *Error and assigns it to target.
//
// Example:
//
//	var appErr *errors.Error
//	if errors.As(err, &appErr) {
//	    code := appErr.Code
//	    meta := appErr.Metadata
//	}
func As(err error, target **Error) bool {
	if err == nil {
		return false
	}

	// Try direct type assertion first (faster)
	if e, ok := err.(*Error); ok {
		*target = e
		return true
	}

	// Fall back to errors.As for wrapped errors
	return errors.As(err, target)
}

// AsError extracts an *Error from err if it exists.
// Returns nil if err is not an *Error.
//
// Example:
//
//	appErr := errors.AsError(err)
//	if appErr != nil {
//	    code := appErr.Code
//	}
func AsError(err error) *Error {
	var e *Error
	if As(err, &e) {
		return e
	}
	return nil
}

// Is checks if an error matches a target error using errors.Is.
// This is a convenience wrapper providing API consistency with As.
func Is(err, target error) bool {
	return errors.Is(err, target)
}
