package errors

import (
	"fmt"
	"regexp"
	"strings"
)

// Code is a machine-readable error code used to identify specific error conditions.
//
// Codes are meant to be stable across versions and used by clients for error handling.
// They follow the format: DOMAIN_RESOURCE_CONDITION
//
// Examples:
//   - USER_EMAIL_TAKEN
//   - ORDER_ALREADY_SHIPPED
//   - PAYMENT_INSUFFICIENT_FUNDS
//   - AUTH_TOKEN_EXPIRED
//
// Codes must:
//   - Be UPPER_SNAKE_CASE
//   - Contain only A-Z, 0-9, and underscore
//   - Start with a letter
//   - Be between 3 and 64 characters
//   - Be registered before use (see registry.go)
type Code string

// Common error codes used across domains.
// Domain-specific codes should be defined in their respective domain packages.
const (
	// Generic codes (avoid using these, prefer domain-specific codes)
	CodeUnknown     Code = "UNKNOWN"
	CodeInternal    Code = "INTERNAL_ERROR"
	CodeUnavailable Code = "SERVICE_UNAVAILABLE"

	// Common validation codes
	CodeInvalidInput  Code = "INVALID_INPUT"
	CodeMissingField  Code = "MISSING_REQUIRED_FIELD"
	CodeInvalidFormat Code = "INVALID_FORMAT"
	CodeOutOfRange    Code = "VALUE_OUT_OF_RANGE"
	CodeTooLong       Code = "VALUE_TOO_LONG"
	CodeTooShort      Code = "VALUE_TOO_SHORT"

	// Common resource codes
	CodeNotFound      Code = "RESOURCE_NOT_FOUND"
	CodeAlreadyExists Code = "RESOURCE_ALREADY_EXISTS"
	CodeConflict      Code = "RESOURCE_CONFLICT"

	// Common auth codes
	CodeUnauthenticated         Code = "UNAUTHENTICATED"
	CodeUnauthorized            Code = "UNAUTHORIZED"
	CodeInsufficientPermissions Code = "INSUFFICIENT_PERMISSIONS"
	CodeTokenExpired            Code = "TOKEN_EXPIRED"
	CodeTokenInvalid            Code = "TOKEN_INVALID"

	// Common rate limit codes
	CodeRateLimitExceeded Code = "RATE_LIMIT_EXCEEDED"
	CodeQuotaExceeded     Code = "QUOTA_EXCEEDED"

	// Common timeout codes
	CodeTimeout          Code = "OPERATION_TIMEOUT"
	CodeDeadlineExceeded Code = "DEADLINE_EXCEEDED"
)

// codeRegex defines the valid format for error codes.
// Must be UPPER_SNAKE_CASE: starts with letter, contains only A-Z, 0-9, underscore.
var codeRegex = regexp.MustCompile(`^[A-Z][A-Z0-9_]*[A-Z0-9]$`)

const (
	minCodeLength = 3
	maxCodeLength = 64
)

// IsValid checks if a code follows the required format.
func (c Code) IsValid() bool {
	s := string(c)

	// Check length
	if len(s) < minCodeLength || len(s) > maxCodeLength {
		return false
	}

	// Check format
	return codeRegex.MatchString(s)
}

// Validate returns an error if the code is invalid.
func (c Code) Validate() error {
	s := string(c)

	if len(s) < minCodeLength {
		return fmt.Errorf("error code too short: must be at least %d characters", minCodeLength)
	}

	if len(s) > maxCodeLength {
		return fmt.Errorf("error code too long: must be at most %d characters", maxCodeLength)
	}

	if !codeRegex.MatchString(s) {
		return fmt.Errorf("error code invalid format: must be UPPER_SNAKE_CASE (got: %s)", s)
	}

	return nil
}

// String returns the string representation of the code.
func (c Code) String() string {
	return string(c)
}

// Domain extracts the domain prefix from a code.
// For "USER_EMAIL_TAKEN", returns "USER".
// For "ORDER_ITEM_OUT_OF_STOCK", returns "ORDER".
// Returns empty string if code has no underscore.
func (c Code) Domain() string {
	s := string(c)
	idx := strings.Index(s, "_")
	if idx == -1 {
		return ""
	}
	return s[:idx]
}

// ParseCode converts a string to a Code and validates it.
// Returns an error if the code format is invalid.
func ParseCode(s string) (Code, error) {
	c := Code(s)
	if err := c.Validate(); err != nil {
		return "", err
	}
	return c, nil
}

// MustParseCode converts a string to a Code and panics if invalid.
// Only use this for constants where you're certain the format is correct.
func MustParseCode(s string) Code {
	c, err := ParseCode(s)
	if err != nil {
		panic(fmt.Sprintf("invalid error code: %v", err))
	}
	return c
}

// NormalizeCode converts a string to valid Code format.
// It uppercases the string and replaces invalid characters with underscores.
// Useful for generating codes from strings, but prefer explicit constants.
//
// Examples:
//   - "user email taken" -> "USER_EMAIL_TAKEN"
//   - "order-not-found"  -> "ORDER_NOT_FOUND"
//   - "invalid@email"    -> "INVALID_EMAIL"
func NormalizeCode(s string) Code {
	// Convert to uppercase
	s = strings.ToUpper(s)

	// Replace spaces, hyphens, and invalid chars with underscores
	s = strings.Map(func(r rune) rune {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '_'
	}, s)

	// Remove duplicate underscores
	for strings.Contains(s, "__") {
		s = strings.ReplaceAll(s, "__", "_")
	}

	// Trim leading/trailing underscores
	s = strings.Trim(s, "_")

	return Code(s)
}

// Equals compares two codes for equality (case-insensitive).
func (c Code) Equals(other Code) bool {
	return strings.EqualFold(string(c), string(other))
}

// IsEmpty returns true if the code is empty.
func (c Code) IsEmpty() bool {
	return len(c) == 0
}

// WithSuffix appends a suffix to the code with an underscore.
// For example: USER_NOT_FOUND.WithSuffix("ARCHIVED") -> "USER_NOT_FOUND_ARCHIVED"
func (c Code) WithSuffix(suffix string) Code {
	if c.IsEmpty() {
		return NormalizeCode(suffix)
	}
	return Code(string(c) + "_" + strings.ToUpper(suffix))
}

// WithPrefix prepends a prefix to the code with an underscore.
// For example: NOT_FOUND.WithPrefix("USER") -> "USER_NOT_FOUND"
func (c Code) WithPrefix(prefix string) Code {
	if c.IsEmpty() {
		return NormalizeCode(prefix)
	}
	return Code(strings.ToUpper(prefix) + "_" + string(c))
}
