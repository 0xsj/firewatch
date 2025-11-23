package errors

import (
	"encoding/json"
	"fmt"
)

// ============================================================================
// Metadata Operations
// ============================================================================

// MergeMeta merges metadata from multiple errors into a single map.
// Later errors override earlier ones for duplicate keys.
//
// Example:
//
//	err1 := errors.NotFound(op, "user").WithMeta("user_id", "123")
//	err2 := errors.Validation(op, "invalid").WithMeta("field", "email")
//	meta := errors.MergeMeta(err1, err2)
//	// Result: {"user_id": "123", "field": "email"}
func MergeMeta(errs ...error) map[string]any {
	result := make(map[string]any)

	for _, err := range errs {
		if err == nil {
			continue
		}

		meta := MetaOf(err)
		for k, v := range meta {
			result[k] = v
		}
	}

	return result
}

// CopyMeta creates a deep copy of error metadata.
// This is useful when you need to modify metadata without affecting the original error.
//
// Example:
//
//	meta := errors.CopyMeta(err)
//	meta["new_field"] = "value"  // Doesn't affect original error
func CopyMeta(err error) map[string]any {
	meta := MetaOf(err)
	if meta == nil {
		return nil
	}

	result := make(map[string]any, len(meta))
	for k, v := range meta {
		result[k] = deepCopyValue(v)
	}

	return result
}

// deepCopyValue attempts to deep copy a value.
// Handles maps and slices recursively, other types are copied by value.
func deepCopyValue(v any) any {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case map[string]any:
		result := make(map[string]any, len(val))
		for k, v := range val {
			result[k] = deepCopyValue(v)
		}
		return result

	case []any:
		result := make([]any, len(val))
		for i, v := range val {
			result[i] = deepCopyValue(v)
		}
		return result

	default:
		// For primitive types, structs, etc., just return as-is
		// Go will copy by value
		return val
	}
}

// FilterMeta filters metadata by keeping only specified keys.
//
// Example:
//
//	meta := errors.FilterMeta(err, "user_id", "request_id")
//	// Only includes user_id and request_id if they exist
func FilterMeta(err error, keys ...string) map[string]any {
	meta := MetaOf(err)
	if meta == nil {
		return nil
	}

	result := make(map[string]any)
	for _, key := range keys {
		if value, exists := meta[key]; exists {
			result[key] = value
		}
	}

	return result
}

// OmitMeta filters metadata by removing specified keys.
// This is useful for removing sensitive information before logging or returning errors.
//
// Example:
//
//	meta := errors.OmitMeta(err, "password", "token", "api_key")
//	// Returns all metadata except sensitive fields
func OmitMeta(err error, keys ...string) map[string]any {
	meta := MetaOf(err)
	if meta == nil {
		return nil
	}

	// Create set of keys to omit
	omitSet := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		omitSet[key] = struct{}{}
	}

	result := make(map[string]any)
	for k, v := range meta {
		if _, shouldOmit := omitSet[k]; !shouldOmit {
			result[k] = v
		}
	}

	return result
}

// SanitizeMeta removes sensitive metadata fields.
// Default sensitive fields: password, token, secret, api_key, auth, credentials
//
// Example:
//
//	meta := errors.SanitizeMeta(err)
//	// Safe to log or return to client
func SanitizeMeta(err error) map[string]any {
	return OmitMeta(err,
		"password",
		"token",
		"access_token",
		"refresh_token",
		"api_key",
		"secret",
		"auth",
		"authorization",
		"credentials",
		"private_key",
		"session_id",
		"cookie",
	)
}

// MetaKeys returns all metadata keys from an error.
//
// Example:
//
//	keys := errors.MetaKeys(err)
//	// ["user_id", "email", "request_id"]
func MetaKeys(err error) []string {
	meta := MetaOf(err)
	if meta == nil {
		return nil
	}

	keys := make([]string, 0, len(meta))
	for k := range meta {
		keys = append(keys, k)
	}

	return keys
}

// MetaSize returns the number of metadata entries.
func MetaSize(err error) int {
	meta := MetaOf(err)
	if meta == nil {
		return 0
	}
	return len(meta)
}

// HasAnyMeta checks if an error has any of the specified metadata keys.
//
// Example:
//
//	if errors.HasAnyMeta(err, "user_id", "session_id") {
//	    // Error has user context
//	}
func HasAnyMeta(err error, keys ...string) bool {
	meta := MetaOf(err)
	if meta == nil {
		return false
	}

	for _, key := range keys {
		if _, exists := meta[key]; exists {
			return true
		}
	}

	return false
}

// HasAllMeta checks if an error has all of the specified metadata keys.
//
// Example:
//
//	if errors.HasAllMeta(err, "user_id", "email") {
//	    // Error has complete user context
//	}
func HasAllMeta(err error, keys ...string) bool {
	meta := MetaOf(err)
	if meta == nil {
		return len(keys) == 0
	}

	for _, key := range keys {
		if _, exists := meta[key]; !exists {
			return false
		}
	}

	return true
}

// ============================================================================
// Metadata Transformation
// ============================================================================

// MetaToJSON converts error metadata to JSON string.
// Returns empty string if metadata is nil or marshaling fails.
//
// Example:
//
//	json := errors.MetaToJSON(err)
//	// {"user_id":"123","email":"user@example.com"}
func MetaToJSON(err error) string {
	meta := MetaOf(err)
	if meta == nil {
		return ""
	}

	bytes, err := json.Marshal(meta)
	if err != nil {
		return ""
	}

	return string(bytes)
}

// MetaToJSONPretty converts error metadata to pretty-printed JSON.
//
// Example:
//
//	json := errors.MetaToJSONPretty(err)
//	// {
//	//   "user_id": "123",
//	//   "email": "user@example.com"
//	// }
func MetaToJSONPretty(err error) string {
	meta := MetaOf(err)
	if meta == nil {
		return ""
	}

	bytes, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return ""
	}

	return string(bytes)
}

// MetaFromJSON parses JSON string into metadata map.
// Returns nil if parsing fails.
//
// Example:
//
//	meta := errors.MetaFromJSON(`{"user_id":"123"}`)
func MetaFromJSON(jsonStr string) map[string]any {
	var meta map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &meta); err != nil {
		return nil
	}
	return meta
}

// ============================================================================
// Metadata Validation
// ============================================================================

// ValidateMeta checks if metadata contains required keys.
// Returns an error if any required key is missing.
//
// Example:
//
//	if err := errors.ValidateMeta(err, "user_id", "request_id"); err != nil {
//	    // Missing required metadata
//	}
func ValidateMeta(err error, requiredKeys ...string) error {
	meta := MetaOf(err)

	var missing []string
	for _, key := range requiredKeys {
		if meta == nil {
			missing = append(missing, key)
			continue
		}
		if _, exists := meta[key]; !exists {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required metadata keys: %v", missing)
	}

	return nil
}
