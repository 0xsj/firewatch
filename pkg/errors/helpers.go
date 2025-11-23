package errors

import (
	"errors"
	"fmt"
)

// ============================================================================
// Extraction Helpers
// ============================================================================

// KindOf extracts the Kind from an error.
// Returns KindInternal if the error is not an *Error.
//
// Example:
//
//	if errors.KindOf(err) == errors.KindNotFound {
//	    return http.StatusNotFound
//	}
func KindOf(err error) Kind {
	if err == nil {
		return KindOther
	}

	var e *Error
	if As(err, &e) {
		return e.Kind
	}

	return KindInternal
}

// CodeOf extracts the Code from an error.
// Returns empty string if the error is not an *Error.
//
// Example:
//
//	code := errors.CodeOf(err)
//	if code == "USER_EMAIL_TAKEN" {
//	    // Handle specific error
//	}
func CodeOf(err error) Code {
	if err == nil {
		return ""
	}

	var e *Error
	if As(err, &e) {
		return e.Code
	}

	return ""
}

// OpOf extracts the Operation from an error.
// Returns empty string if the error is not an *Error.
//
// Example:
//
//	op := errors.OpOf(err)
//	log.Error("operation failed", "operation", op, "error", err)
func OpOf(err error) string {
	if err == nil {
		return ""
	}

	var e *Error
	if As(err, &e) {
		return e.Operation
	}

	return ""
}

// MessageOf extracts the Message from an error.
// Returns err.Error() if the error is not an *Error.
//
// Example:
//
//	msg := errors.MessageOf(err)
//	fmt.Println(msg)
func MessageOf(err error) string {
	if err == nil {
		return ""
	}

	var e *Error
	if As(err, &e) {
		return e.Message
	}

	return err.Error()
}

// MetaOf extracts the Metadata from an error.
// Returns nil if the error is not an *Error or has no metadata.
//
// Example:
//
//	meta := errors.MetaOf(err)
//	if userID, ok := meta["user_id"].(string); ok {
//	    log.Info("error context", "user_id", userID)
//	}
func MetaOf(err error) map[string]any {
	if err == nil {
		return nil
	}

	var e *Error
	if As(err, &e) {
		return e.Metadata
	}

	return nil
}

// SeverityOf extracts the Severity from an error.
// Returns SeverityError if the error is not an *Error.
//
// Example:
//
//	severity := errors.SeverityOf(err)
//	if severity == errors.SeverityWarning {
//	    log.Warn(err.Error())
//	} else {
//	    log.Error(err.Error())
//	}
func SeverityOf(err error) Severity {
	if err == nil {
		return SeverityError
	}

	var e *Error
	if As(err, &e) {
		return e.Severity
	}

	return SeverityError
}

// IsRetryable checks if an error is retryable.
// Returns false if the error is not an *Error.
//
// Example:
//
//	if errors.IsRetryable(err) {
//	    time.Sleep(backoff)
//	    return retry()
//	}
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	var e *Error
	if As(err, &e) {
		return e.Retryable
	}

	return false
}

// ============================================================================
// Enhanced Sentinel Checks (combining errors.Is + Kind checks)
// ============================================================================

// IsNotFound checks if an error is a not found error.
// Checks both sentinel (fast) and kind (contextual).
func IsNotFound(err error) bool {
	return Is(err, ErrNotFound) || KindOf(err) == KindNotFound
}

// IsValidation checks if an error is a validation error.
func IsValidation(err error) bool {
	return Is(err, ErrValidation) || KindOf(err) == KindValidation
}

// IsConflict checks if an error is a conflict error.
func IsConflict(err error) bool {
	return Is(err, ErrConflict) || KindOf(err) == KindConflict
}

// IsUnauthorized checks if an error is an unauthorized error.
func IsUnauthorized(err error) bool {
	return Is(err, ErrUnauthorized) || KindOf(err) == KindUnauthorized
}

// IsForbidden checks if an error is a forbidden error.
func IsForbidden(err error) bool {
	return Is(err, ErrForbidden) || KindOf(err) == KindForbidden
}

// IsDomain checks if an error is a domain error.
func IsDomain(err error) bool {
	return Is(err, ErrDomain) || KindOf(err) == KindDomain
}

// IsInfrastructure checks if an error is an infrastructure error.
func IsInfrastructure(err error) bool {
	return Is(err, ErrInfrastructure) || KindOf(err) == KindInfrastructure
}

// IsTimeout checks if an error is a timeout error.
func IsTimeout(err error) bool {
	return Is(err, ErrTimeout) || KindOf(err) == KindTimeout
}

// IsRateLimit checks if an error is a rate limit error.
func IsRateLimit(err error) bool {
	return Is(err, ErrRateLimit) || KindOf(err) == KindRateLimit
}

// IsInternal checks if an error is an internal error.
func IsInternal(err error) bool {
	return Is(err, ErrInternal) || KindOf(err) == KindInternal
}

// ============================================================================
// Metadata Helpers
// ============================================================================

// GetMeta extracts a metadata value by key.
// Returns nil if the key doesn't exist or error is not an *Error.
//
// Example:
//
//	userID := errors.GetMeta(err, "user_id")
//	if id, ok := userID.(string); ok {
//	    log.Info("user context", "user_id", id)
//	}
func GetMeta(err error, key string) any {
	meta := MetaOf(err)
	if meta == nil {
		return nil
	}
	return meta[key]
}

// GetMetaString extracts a string metadata value.
// Returns empty string if the key doesn't exist or value is not a string.
//
// Example:
//
//	userID := errors.GetMetaString(err, "user_id")
//	if userID != "" {
//	    log.Info("user context", "user_id", userID)
//	}
func GetMetaString(err error, key string) string {
	value := GetMeta(err, key)
	if s, ok := value.(string); ok {
		return s
	}
	return ""
}

// GetMetaInt extracts an int metadata value.
// Returns 0 if the key doesn't exist or value is not an int.
func GetMetaInt(err error, key string) int {
	value := GetMeta(err, key)
	if i, ok := value.(int); ok {
		return i
	}
	return 0
}

// GetMetaBool extracts a bool metadata value.
// Returns false if the key doesn't exist or value is not a bool.
func GetMetaBool(err error, key string) bool {
	value := GetMeta(err, key)
	if b, ok := value.(bool); ok {
		return b
	}
	return false
}

// HasMeta checks if an error has a specific metadata key.
//
// Example:
//
//	if errors.HasMeta(err, "user_id") {
//	    // Error has user context
//	}
func HasMeta(err error, key string) bool {
	meta := MetaOf(err)
	if meta == nil {
		return false
	}
	_, exists := meta[key]
	return exists
}

// ============================================================================
// Error Classification
// ============================================================================

// IsClientError returns true if the error represents a client-side error
// that should not be retried without modification.
//
// Client errors include: validation, not found, unauthorized, forbidden, conflict, domain.
func IsClientError(err error) bool {
	return KindOf(err).IsClientError()
}

// IsServerError returns true if the error represents a server-side error
// that might be retryable or indicates an infrastructure problem.
//
// Server errors include: infrastructure, timeout, internal.
func IsServerError(err error) bool {
	return KindOf(err).IsServerError()
}

// ShouldAlert returns true if the error should trigger an alert.
// Based on severity level - only SeverityError triggers alerts.
func ShouldAlert(err error) bool {
	return SeverityOf(err).ShouldAlert()
}

// ShouldLog returns true if the error should be logged.
// All errors should be logged, but at different levels based on severity.
func ShouldLog(err error) bool {
	return SeverityOf(err).ShouldLog()
}

// ============================================================================
// Error Comparison
// ============================================================================

// HasKind checks if an error has a specific kind.
//
// Example:
//
//	if errors.HasKind(err, errors.KindNotFound) {
//	    return http.StatusNotFound
//	}
func HasKind(err error, kind Kind) bool {
	return KindOf(err) == kind
}

// HasCode checks if an error has a specific code.
//
// Example:
//
//	if errors.HasCode(err, "USER_EMAIL_TAKEN") {
//	    return "This email is already registered"
//	}
func HasCode(err error, code Code) bool {
	return CodeOf(err) == code
}

// HasSeverity checks if an error has a specific severity.
func HasSeverity(err error, severity Severity) bool {
	return SeverityOf(err) == severity
}

// ============================================================================
// Error Chain Helpers
// ============================================================================

// Unwrap returns the wrapped error.
// This is a convenience wrapper around errors.Unwrap.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// UnwrapAll returns all errors in the error chain.
// Returns a slice of errors from outermost to innermost.
//
// Example:
//
//	errs := errors.UnwrapAll(err)
//	for _, e := range errs {
//	    fmt.Println(e.Error())
//	}
func UnwrapAll(err error) []error {
	if err == nil {
		return nil
	}

	var errs []error
	for err != nil {
		errs = append(errs, err)

		// If this is our Error type, check if it wraps another *Error
		var appErr *Error
		if As(err, &appErr) {
			// Only continue unwrapping if the wrapped error is also an *Error
			unwrapped := errors.Unwrap(err)
			var innerAppErr *Error
			if unwrapped != nil && As(unwrapped, &innerAppErr) {
				err = unwrapped
				continue
			}
			// Stop here - don't unwrap to sentinel
			break
		}

		// For non-app errors, continue normal unwrapping
		err = errors.Unwrap(err)
	}
	return errs
}

// Root returns the root cause of an error chain.
// Returns the innermost error that doesn't wrap another error.
//
// Example:
//
//	rootErr := errors.Root(err)
//	if rootErr == sql.ErrNoRows {
//	    // Handle database not found
//	}
func Root(err error) error {
	if err == nil {
		return nil
	}

	var lastAppErr *Error
	for {
		// Check if current error is an *Error
		var appErr *Error
		if As(err, &appErr) {
			lastAppErr = appErr
		}

		unwrapped := errors.Unwrap(err)
		if unwrapped == nil {
			// If we found an app error in the chain, return it
			if lastAppErr != nil {
				return lastAppErr
			}
			// Otherwise return the last error
			return err
		}
		err = unwrapped
	}
}

// ============================================================================
// Error Formatting
// ============================================================================

// Format returns a formatted error string with all context.
// Includes: operation, message, kind, code, metadata.
//
// Example output:
//
//	[users.Repository.Create] user not found (kind=not_found, code=USER_NOT_FOUND, user_id=123)
func Format(err error) string {
	if err == nil {
		return ""
	}

	var e *Error
	if !As(err, &e) {
		return err.Error()
	}

	result := ""
	if e.Operation != "" {
		result = "[" + e.Operation + "] "
	}
	result += e.Message

	if e.Kind != KindOther {
		result += " (kind=" + e.Kind.String()
		if e.Code != "" {
			result += ", code=" + string(e.Code)
		}
		if len(e.Metadata) > 0 {
			for k, v := range e.Metadata {
				result += ", " + k + "=" + fmt.Sprint(v)
			}
		}
		result += ")"
	}

	return result
}

// Details returns a multi-line detailed error string.
// Useful for debugging and error logs.
//
// Example output:
//
//	Error: user not found
//	Kind: not_found
//	Code: USER_NOT_FOUND
//	Operation: users.Repository.Create
//	Severity: error
//	Retryable: false
//	Metadata:
//	  user_id: 123
//	  email: user@example.com
func Details(err error) string {
	if err == nil {
		return "nil"
	}

	var e *Error
	if !As(err, &e) {
		return fmt.Sprintf("Non-Error type: %v", err)
	}

	result := "Error: " + e.Message + "\n"
	result += "Kind: " + e.Kind.String() + "\n"
	result += "Code: " + string(e.Code) + "\n"
	result += "Operation: " + e.Operation + "\n"
	result += "Severity: " + e.Severity.String() + "\n"
	result += "Retryable: " + fmt.Sprint(e.Retryable)

	if len(e.Metadata) > 0 {
		result += "\nMetadata:\n"
		for k, v := range e.Metadata {
			result += "  " + k + ": " + fmt.Sprint(v) + "\n"
		}
	}

	return result
}
