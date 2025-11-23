package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// ID is a strongly-typed, globally unique identifier.
// It uses UUIDv7 (RFC 9562) which embeds a timestamp for natural ordering
// and efficient database indexing.
//
// Zero value is invalid - use NewID() to create valid IDs.
//
// Example:
//
//	userID := types.NewID()
//	fmt.Println(userID.String()) // "018e8c3a-7f1b-7c3d-9e4f-a1b2c3d4e5f6"
type ID struct {
	value uuid.UUID
}

// NewID generates a new time-ordered unique ID (UUIDv7).
// The ID embeds the current timestamp in its first 48 bits,
// enabling natural chronological sorting.
func NewID() ID {
	return ID{
		value: uuid.Must(uuid.NewV7()),
	}
}

// ParseID parses a string into an ID.
// Accepts standard UUID format (with or without hyphens).
// Returns an error if the string is not a valid UUID.
//
// Example:
//
//	id, err := types.ParseID("018e8c3a-7f1b-7c3d-9e4f-a1b2c3d4e5f6")
//	if err != nil {
//	    return err
//	}
func ParseID(s string) (ID, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return ID{}, fmt.Errorf("id cannot be empty")
	}

	parsed, err := uuid.Parse(s)
	if err != nil {
		return ID{}, fmt.Errorf("invalid id format: %w", err)
	}

	return ID{value: parsed}, nil
}

// MustParseID parses a string into an ID and panics if invalid.
// Only use this for constants where you're certain the value is valid.
func MustParseID(s string) ID {
	id, err := ParseID(s)
	if err != nil {
		panic(fmt.Sprintf("invalid id: %v", err))
	}
	return id
}

// FromUUID creates an ID from an existing UUID.
// Useful when integrating with external systems that provide UUIDs.
func FromUUID(u uuid.UUID) ID {
	return ID{value: u}
}

// String returns the canonical string representation of the ID.
// Format: "018e8c3a-7f1b-7c3d-9e4f-a1b2c3d4e5f6" (lowercase with hyphens)
func (id ID) String() string {
	return id.value.String()
}

// UUID returns the underlying UUID value.
// Useful when interfacing with libraries that expect uuid.UUID.
func (id ID) UUID() uuid.UUID {
	return id.value
}

// IsZero returns true if the ID is the zero value (invalid/uninitialized).
func (id ID) IsZero() bool {
	return id.value == uuid.Nil
}

// IsValid returns true if the ID is valid (non-zero).
func (id ID) IsValid() bool {
	return id.value != uuid.Nil
}

// Equals checks if two IDs are equal.
func (id ID) Equals(other ID) bool {
	return id.value == other.value
}

// Compare compares two IDs.
// For UUIDv7, this provides chronological ordering.
// Returns:
//   - -1 if id < other (id created before other)
//   - 0 if id == other
//   - +1 if id > other (id created after other)
func (id ID) Compare(other ID) int {
	// Compare as byte arrays for proper ordering
	for i := range 16 {
		if id.value[i] < other.value[i] {
			return -1
		}
		if id.value[i] > other.value[i] {
			return 1
		}
	}
	return 0
}

// Time returns the Unix timestamp (milliseconds) embedded in the ID.
// Only meaningful for UUIDv7 - returns 0 for other UUID versions.
//
// UUIDv7 embeds a 48-bit timestamp in the first 6 bytes.
func (id ID) Time() int64 {
	if id.IsZero() {
		return 0
	}

	// UUIDv7 timestamp is in first 48 bits (6 bytes)
	// Format: timestamp_ms (48 bits) | ver (4 bits) | rand (12 bits) | var (2 bits) | rand (62 bits)
	timestamp := int64(id.value[0])<<40 |
		int64(id.value[1])<<32 |
		int64(id.value[2])<<24 |
		int64(id.value[3])<<16 |
		int64(id.value[4])<<8 |
		int64(id.value[5])

	return timestamp
}

// Version returns the UUID version (should be 7 for UUIDv7).
func (id ID) Version() uuid.Version {
	return id.value.Version()
}

// ============================================================================
// JSON Marshaling
// ============================================================================

// MarshalJSON implements json.Marshaler.
// Encodes ID as a JSON string.
//
// Example: "018e8c3a-7f1b-7c3d-9e4f-a1b2c3d4e5f6"
func (id ID) MarshalJSON() ([]byte, error) {
	if id.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(id.String())
}

// UnmarshalJSON implements json.Unmarshaler.
// Decodes ID from JSON string.
func (id *ID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	if s == "" || s == "null" {
		*id = ID{}
		return nil
	}

	parsed, err := ParseID(s)
	if err != nil {
		return err
	}

	*id = parsed
	return nil
}

// ============================================================================
// SQL Scanning (Native UUID Support)
// ============================================================================

// Scan implements sql.Scanner.
// Supports native UUID database types and string/byte representations.
//
// PostgreSQL example:
//
//	CREATE TABLE users (id UUID PRIMARY KEY);
//	SELECT id FROM users; -- scans directly as UUID
func (id *ID) Scan(value interface{}) error {
	if value == nil {
		*id = ID{}
		return nil
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			*id = ID{}
			return nil
		}
		parsed, err := ParseID(v)
		if err != nil {
			return err
		}
		*id = parsed
		return nil

	case []byte:
		if len(v) == 0 {
			*id = ID{}
			return nil
		}

		// Try parsing as string first (most common)
		if len(v) == 36 || len(v) == 32 {
			parsed, err := ParseID(string(v))
			if err != nil {
				return err
			}
			*id = parsed
			return nil
		}

		// Try parsing as 16-byte binary UUID
		if len(v) == 16 {
			var u uuid.UUID
			copy(u[:], v)
			*id = ID{value: u}
			return nil
		}

		return fmt.Errorf("invalid byte length for UUID: %d", len(v))

	case uuid.UUID:
		*id = ID{value: v}
		return nil

	default:
		return fmt.Errorf("cannot scan %T into ID", value)
	}
}

// Value implements driver.Valuer.
// Returns UUID for native database UUID types.
//
// PostgreSQL will receive native UUID type (16 bytes).
func (id ID) Value() (driver.Value, error) {
	if id.IsZero() {
		return nil, nil
	}
	// Return the UUID directly - database/sql will handle conversion
	// For PostgreSQL, this becomes a native UUID type
	return id.value, nil
}

// ============================================================================
// Text Marshaling (for formats like YAML, TOML)
// ============================================================================

// MarshalText implements encoding.TextMarshaler.
func (id ID) MarshalText() ([]byte, error) {
	if id.IsZero() {
		return []byte{}, nil
	}
	return []byte(id.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (id *ID) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*id = ID{}
		return nil
	}

	parsed, err := ParseID(string(text))
	if err != nil {
		return err
	}

	*id = parsed
	return nil
}

// ============================================================================
// Helpers
// ============================================================================

// GoString implements fmt.GoStringer for better debugging output.
func (id ID) GoString() string {
	if id.IsZero() {
		return "ID{zero}"
	}
	return fmt.Sprintf("ID{%s}", id.String())
}

// Bytes returns the raw 16-byte representation of the UUID.
// Useful for binary protocols or custom serialization.
func (id ID) Bytes() []byte {
	return id.value[:]
}

// IsEmpty returns true if the ID is empty (zero value).
func (id ID) IsEmpty() bool {
	return id.value == uuid.Nil
}
