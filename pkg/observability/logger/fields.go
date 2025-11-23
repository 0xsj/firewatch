package logger

import (
	"fmt"
	"time"
)

// Field is a key-value pair for structured logging.
type Field struct {
	Key   string
	Value any
}

// ============================================================================
// Field Constructors
// ============================================================================

// String creates a string field.
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates an int field.
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 creates an int64 field.
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Uint creates a uint field.
func Uint(key string, value uint) Field {
	return Field{Key: key, Value: value}
}

// Float64 creates a float64 field.
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool creates a bool field.
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Time creates a time field.
func Time(key string, value time.Time) Field {
	return Field{Key: key, Value: value}
}

// Duration creates a duration field.
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

// Error creates an error field.
func Error(err error) Field {
	if err == nil {
		return Field{Key: "error", Value: nil}
	}
	return Field{Key: "error", Value: err.Error()}
}

// Err is an alias for Error.
func Err(err error) Field {
	return Error(err)
}

// Any creates a field with any type.
// The value will be formatted using fmt.Sprintf("%+v", value).
func Any(key string, value any) Field {
	return Field{Key: key, Value: value}
}

// Stringer creates a field from a fmt.Stringer.
func Stringer(key string, value fmt.Stringer) Field {
	if value == nil {
		return Field{Key: key, Value: nil}
	}
	return Field{Key: key, Value: value.String()}
}

// ============================================================================
// Common Field Names (conventions)
// ============================================================================

// These are recommended field names for consistency.
const (
	FieldTenantID      = "tenant_id"
	FieldUserID        = "user_id"
	FieldRequestID     = "request_id"
	FieldCorrelationID = "correlation_id"
	FieldSessionID     = "session_id"
	FieldOperation     = "operation"
	FieldComponent     = "component"
	FieldMethod        = "method"
	FieldPath          = "path"
	FieldStatusCode    = "status_code"
	FieldDuration      = "duration"
	FieldError         = "error"
	FieldStackTrace    = "stack_trace"
)

// ============================================================================
// Field Helpers
// ============================================================================

// fieldsToMap converts a slice of fields to a map.
func fieldsToMap(fields []Field) map[string]any {
	m := make(map[string]any, len(fields))
	for _, f := range fields {
		m[f.Key] = f.Value
	}
	return m
}

// mergeFields merges two field slices (later fields override earlier ones).
func mergeFields(base, additional []Field) []Field {
	if len(base) == 0 {
		return additional
	}
	if len(additional) == 0 {
		return base
	}

	// Use map to deduplicate by key
	m := make(map[string]any, len(base)+len(additional))
	keys := make([]string, 0, len(base)+len(additional))

	for _, f := range base {
		if _, exists := m[f.Key]; !exists {
			keys = append(keys, f.Key)
		}
		m[f.Key] = f.Value
	}

	for _, f := range additional {
		if _, exists := m[f.Key]; !exists {
			keys = append(keys, f.Key)
		}
		m[f.Key] = f.Value
	}

	// Convert back to fields slice preserving order
	result := make([]Field, 0, len(keys))
	for _, key := range keys {
		result = append(result, Field{Key: key, Value: m[key]})
	}

	return result
}
