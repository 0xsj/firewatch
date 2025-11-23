package errors

import "strings"

// Severity represents the severity level of an error.
// This is used for:
//   - Logging: Map severity to log levels (Error, Warn, Info)
//   - Alerting: Only alert on SeverityError, ignore warnings
//   - Metrics: Track error distribution by severity
//   - Client display: Show warnings differently from errors
type Severity uint8

const (
	// SeverityError represents a critical error that prevented the operation.
	// The request failed and could not be completed.
	//
	// Use when:
	//   - Operation completely failed
	//   - Data integrity is at risk
	//   - User action was blocked
	//
	// Examples:
	//   - Database write failed
	//   - Payment processing failed
	//   - Authentication failed
	//
	// Logging: ERROR level
	// Alerting: Yes
	// Client: Show error message
	SeverityError Severity = iota

	// SeverityWarning represents a non-critical issue.
	// The operation succeeded but with caveats or degraded functionality.
	//
	// Use when:
	//   - Operation completed but not optimally
	//   - Fallback mechanism was used
	//   - Deprecated feature was used
	//   - Partial failure in batch operation
	//
	// Examples:
	//   - Cache miss, fell back to database
	//   - Email sending failed but user was created
	//   - Using default value because config is missing
	//   - 3 out of 5 items processed successfully
	//
	// Logging: WARN level
	// Alerting: No (but monitor trends)
	// Client: Optional warning indicator
	SeverityWarning

	// SeverityInfo represents informational context, not an actual error.
	// Used for expected conditions that should be tracked.
	//
	// Use when:
	//   - Expected business condition (not an error)
	//   - Tracking for analytics
	//   - Audit trail entries
	//
	// Examples:
	//   - User not found (expected for new signups)
	//   - Feature flag disabled
	//   - Rate limit approaching threshold
	//
	// Logging: INFO level
	// Alerting: No
	// Client: May not show at all
	SeverityInfo
)

// String returns the human-readable name of the severity level.
func (s Severity) String() string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warning"
	case SeverityInfo:
		return "info"
	default:
		return "unknown"
	}
}

// IsError returns true if the severity is Error level.
func (s Severity) IsError() bool {
	return s == SeverityError
}

// IsWarning returns true if the severity is Warning level.
func (s Severity) IsWarning() bool {
	return s == SeverityWarning
}

// IsInfo returns true if the severity is Info level.
func (s Severity) IsInfo() bool {
	return s == SeverityInfo
}

// ShouldAlert returns true if errors of this severity should trigger alerts.
// Typically only SeverityError should trigger alerts.
func (s Severity) ShouldAlert() bool {
	return s == SeverityError
}

// ShouldLog returns true if errors of this severity should be logged.
// All severities should be logged, but at different levels.
func (s Severity) ShouldLog() bool {
	return true
}

// Level returns a string suitable for log level mapping.
// This can be used to map severity to your logger's level constants.
//
// Returns:
//   - "error" for SeverityError
//   - "warn" for SeverityWarning
//   - "info" for SeverityInfo
func (s Severity) Level() string {
	return s.String()
}

// ParseSeverity converts a string to a Severity.
// It is case-insensitive and accepts: "error", "warning", "warn", "info".
// Returns SeverityError if the string is not recognized.
func ParseSeverity(s string) Severity {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "error", "err":
		return SeverityError
	case "warning", "warn":
		return SeverityWarning
	case "info", "information":
		return SeverityInfo
	default:
		return SeverityError // Default to error for unknown severities
	}
}

// MustParseSeverity converts a string to a Severity and panics if invalid.
// Only use this for constants where you're certain the value is correct.
func MustParseSeverity(s string) Severity {
	severity := ParseSeverity(s)
	// Since ParseSeverity has a default, we only panic on truly invalid strings
	normalized := strings.ToLower(strings.TrimSpace(s))
	switch normalized {
	case "error", "err", "warning", "warn", "info", "information":
		return severity
	default:
		panic("invalid severity: " + s)
	}
}

// Compare returns:
//
//	-1 if s is less severe than other
//	 0 if s equals other
//	+1 if s is more severe than other
//
// Severity order: Info < Warning < Error
func (s Severity) Compare(other Severity) int {
	if s < other {
		return -1
	}
	if s > other {
		return 1
	}
	return 0
}

// IsMoreSevereThan returns true if s is more severe than other.
func (s Severity) IsMoreSevereThan(other Severity) bool {
	return s > other
}

// IsLessSevereThan returns true if s is less severe than other.
func (s Severity) IsLessSevereThan(other Severity) bool {
	return s < other
}

// MaxSeverity returns the more severe of two severity levels.
func MaxSeverity(a, b Severity) Severity {
	if a > b {
		return a
	}
	return b
}

// MinSeverity returns the less severe of two severity levels.
func MinSeverity(a, b Severity) Severity {
	if a < b {
		return a
	}
	return b
}
