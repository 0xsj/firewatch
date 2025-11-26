package tenant

import (
	"encoding/json"
)

// Settings represents tenant-level configuration.
// Uses a flexible map structure to accommodate varying needs across applications.
type Settings struct {
	values map[string]any
}

// NewSettings creates a new empty Settings.
func NewSettings() Settings {
	return Settings{
		values: make(map[string]any),
	}
}

// NewSettingsFromMap creates Settings from an existing map.
// Returns a defensive copy.
func NewSettingsFromMap(m map[string]any) Settings {
	if m == nil {
		return NewSettings()
	}

	values := make(map[string]any, len(m))
	for k, v := range m {
		values[k] = v
	}

	return Settings{values: values}
}

// Get retrieves a value by key.
// Returns nil if the key does not exist.
func (s Settings) Get(key string) any {
	return s.values[key]
}

// GetString retrieves a string value by key.
// Returns empty string if the key does not exist or is not a string.
func (s Settings) GetString(key string) string {
	if v, ok := s.values[key].(string); ok {
		return v
	}
	return ""
}

// GetInt retrieves an int value by key.
// Returns 0 if the key does not exist or is not numeric.
func (s Settings) GetInt(key string) int {
	switch v := s.values[key].(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}

// GetBool retrieves a bool value by key.
// Returns false if the key does not exist or is not a bool.
func (s Settings) GetBool(key string) bool {
	if v, ok := s.values[key].(bool); ok {
		return v
	}
	return false
}

// Has checks if a key exists.
func (s Settings) Has(key string) bool {
	_, ok := s.values[key]
	return ok
}

// Set sets a value by key.
// Returns a new Settings instance (immutable pattern).
func (s Settings) Set(key string, value any) Settings {
	newValues := make(map[string]any, len(s.values)+1)
	for k, v := range s.values {
		newValues[k] = v
	}
	newValues[key] = value
	return Settings{values: newValues}
}

// Delete removes a key.
// Returns a new Settings instance (immutable pattern).
func (s Settings) Delete(key string) Settings {
	newValues := make(map[string]any, len(s.values))
	for k, v := range s.values {
		if k != key {
			newValues[k] = v
		}
	}
	return Settings{values: newValues}
}

// Merge combines two Settings, with other taking precedence.
// Returns a new Settings instance.
func (s Settings) Merge(other Settings) Settings {
	newValues := make(map[string]any, len(s.values)+len(other.values))
	for k, v := range s.values {
		newValues[k] = v
	}
	for k, v := range other.values {
		newValues[k] = v
	}
	return Settings{values: newValues}
}

// Keys returns all keys in the settings.
func (s Settings) Keys() []string {
	keys := make([]string, 0, len(s.values))
	for k := range s.values {
		keys = append(keys, k)
	}
	return keys
}

// Len returns the number of settings.
func (s Settings) Len() int {
	return len(s.values)
}

// IsEmpty returns true if there are no settings.
func (s Settings) IsEmpty() bool {
	return len(s.values) == 0
}

// ToMap returns a copy of the underlying map.
func (s Settings) ToMap() map[string]any {
	result := make(map[string]any, len(s.values))
	for k, v := range s.values {
		result[k] = v
	}
	return result
}

// Equals checks if two Settings are equal.
func (s Settings) Equals(other Settings) bool {
	if len(s.values) != len(other.values) {
		return false
	}
	for k, v := range s.values {
		if otherV, ok := other.values[k]; !ok || otherV != v {
			return false
		}
	}
	return true
}

// MarshalJSON implements json.Marshaler.
func (s Settings) MarshalJSON() ([]byte, error) {
	if s.values == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(s.values)
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *Settings) UnmarshalJSON(data []byte) error {
	if s.values == nil {
		s.values = make(map[string]any)
	}
	return json.Unmarshal(data, &s.values)
}
