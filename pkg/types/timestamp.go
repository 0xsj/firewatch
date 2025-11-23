package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Timestamp represents a point in time with millisecond precision.
// Always stored in UTC. Immutable value object.
//
// Zero value represents the Unix epoch (1970-01-01 00:00:00 UTC).
// Use IsZero() to check for uninitialized timestamps.
//
// Example:
//
//	ts := types.Now()
//	fmt.Println(ts.String()) // "2024-01-15T10:30:45.123Z"
type Timestamp struct {
	value time.Time
}

// Now returns the current timestamp in UTC with millisecond precision.
func Now() Timestamp {
	return Timestamp{
		value: time.Now().UTC().Truncate(time.Millisecond),
	}
}

// NewTimestamp creates a timestamp from a time.Time value.
// The time is normalized to UTC and truncated to millisecond precision.
func NewTimestamp(t time.Time) Timestamp {
	return Timestamp{
		value: t.UTC().Truncate(time.Millisecond),
	}
}

// TimestampFromTime is an alias for NewTimestamp for consistency.
// Creates a timestamp from a time.Time value.
func TimestampFromTime(t time.Time) Timestamp {
	return NewTimestamp(t)
}

// FromUnixMilli creates a timestamp from Unix milliseconds.
//
// Example:
//
//	ts := types.FromUnixMilli(1705318245123)
func FromUnixMilli(ms int64) Timestamp {
	return Timestamp{
		value: time.UnixMilli(ms).UTC(),
	}
}

// FromUnix creates a timestamp from Unix seconds.
//
// Example:
//
//	ts := types.FromUnix(1705318245)
func FromUnix(sec int64) Timestamp {
	return Timestamp{
		value: time.Unix(sec, 0).UTC(),
	}
}

// ParseTimestamp parses an RFC 3339 / ISO 8601 timestamp string.
//
// Supported formats:
//   - "2024-01-15T10:30:45Z"
//   - "2024-01-15T10:30:45.123Z"
//   - "2024-01-15T10:30:45+00:00"
//
// Example:
//
//	ts, err := types.ParseTimestamp("2024-01-15T10:30:45.123Z")
func ParseTimestamp(s string) (Timestamp, error) {
	if s == "" {
		return Timestamp{}, fmt.Errorf("timestamp cannot be empty")
	}

	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return Timestamp{}, fmt.Errorf("invalid timestamp format: %w", err)
	}

	return NewTimestamp(t), nil
}

// MustParseTimestamp parses a timestamp and panics if invalid.
// Only use for constants where you're certain the value is valid.
func MustParseTimestamp(s string) Timestamp {
	ts, err := ParseTimestamp(s)
	if err != nil {
		panic(fmt.Sprintf("invalid timestamp: %v", err))
	}
	return ts
}

// Time returns the underlying time.Time value in UTC.
func (ts Timestamp) Time() time.Time {
	return ts.value
}

// UnixMilli returns the Unix timestamp in milliseconds.
//
// Example:
//
//	ts.UnixMilli() // 1705318245123
func (ts Timestamp) UnixMilli() int64 {
	return ts.value.UnixMilli()
}

// Unix returns the Unix timestamp in seconds.
//
// Example:
//
//	ts.Unix() // 1705318245
func (ts Timestamp) Unix() int64 {
	return ts.value.Unix()
}

// String returns the ISO 8601 / RFC 3339 representation.
//
// Example:
//
//	ts.String() // "2024-01-15T10:30:45.123Z"
func (ts Timestamp) String() string {
	if ts.IsZero() {
		return ""
	}
	return ts.value.Format(time.RFC3339Nano)
}

// IsZero returns true if the timestamp is the zero value (Unix epoch).
func (ts Timestamp) IsZero() bool {
	return ts.value.IsZero()
}

// IsValid returns true if the timestamp is non-zero.
func (ts Timestamp) IsValid() bool {
	return !ts.value.IsZero()
}

// Equals checks if two timestamps are equal.
func (ts Timestamp) Equals(other Timestamp) bool {
	return ts.value.Equal(other.value)
}

// Compare compares two timestamps.
// Returns:
//   - -1 if ts < other (ts is earlier)
//   - 0 if ts == other
//   - +1 if ts > other (ts is later)
func (ts Timestamp) Compare(other Timestamp) int {
	if ts.value.Before(other.value) {
		return -1
	}
	if ts.value.After(other.value) {
		return 1
	}
	return 0
}

// Before returns true if ts is before other.
func (ts Timestamp) Before(other Timestamp) bool {
	return ts.value.Before(other.value)
}

// After returns true if ts is after other.
func (ts Timestamp) After(other Timestamp) bool {
	return ts.value.After(other.value)
}

// Add adds a duration to the timestamp.
//
// Example:
//
//	expires := ts.Add(24 * time.Hour) // 24 hours from now
func (ts Timestamp) Add(d time.Duration) Timestamp {
	return Timestamp{
		value: ts.value.Add(d).UTC().Truncate(time.Millisecond),
	}
}

// Sub returns the duration between two timestamps (ts - other).
//
// Example:
//
//	duration := ts.Sub(earlier) // time.Duration
func (ts Timestamp) Sub(other Timestamp) time.Duration {
	return ts.value.Sub(other.value)
}

// ============================================================================
// Convenience Time Checkers
// ============================================================================

// IsPast returns true if the timestamp is in the past.
func (ts Timestamp) IsPast() bool {
	return ts.Before(Now())
}

// IsFuture returns true if the timestamp is in the future.
func (ts Timestamp) IsFuture() bool {
	return ts.After(Now())
}

// IsToday returns true if the timestamp is today (in UTC).
func (ts Timestamp) IsToday() bool {
	now := time.Now().UTC()
	y1, m1, d1 := ts.value.Date()
	y2, m2, d2 := now.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// Age returns the duration since the timestamp.
// Returns 0 if the timestamp is in the future.
//
// Example:
//
//	age := createdAt.Age() // 2h30m
func (ts Timestamp) Age() time.Duration {
	now := Now()
	if ts.After(now) {
		return 0
	}
	return now.Sub(ts)
}

// Until returns the duration until the timestamp.
// Returns 0 if the timestamp is in the past.
//
// Example:
//
//	remaining := expiresAt.Until() // 1h30m
func (ts Timestamp) Until() time.Duration {
	now := Now()
	if ts.Before(now) {
		return 0
	}
	return ts.Sub(now)
}

// ============================================================================
// JSON Marshaling (ISO 8601 format)
// ============================================================================

// MarshalJSON implements json.Marshaler.
// Encodes timestamp as ISO 8601 / RFC 3339 string.
//
// Example: "2024-01-15T10:30:45.123Z"
func (ts Timestamp) MarshalJSON() ([]byte, error) {
	if ts.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(ts.value.Format(time.RFC3339Nano))
}

// UnmarshalJSON implements json.Unmarshaler.
// Decodes timestamp from ISO 8601 / RFC 3339 string.
func (ts *Timestamp) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		// Try unmarshaling as null
		var null *string
		if err := json.Unmarshal(data, &null); err == nil && null == nil {
			*ts = Timestamp{}
			return nil
		}
		return err
	}

	if s == "" || s == "null" {
		*ts = Timestamp{}
		return nil
	}

	parsed, err := ParseTimestamp(s)
	if err != nil {
		return err
	}

	*ts = parsed
	return nil
}

// ============================================================================
// SQL Scanning
// ============================================================================

// Scan implements sql.Scanner.
// Supports time.Time and string representations.
func (ts *Timestamp) Scan(value interface{}) error {
	if value == nil {
		*ts = Timestamp{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		*ts = NewTimestamp(v)
		return nil

	case string:
		if v == "" {
			*ts = Timestamp{}
			return nil
		}
		parsed, err := ParseTimestamp(v)
		if err != nil {
			return err
		}
		*ts = parsed
		return nil

	case []byte:
		if len(v) == 0 {
			*ts = Timestamp{}
			return nil
		}
		parsed, err := ParseTimestamp(string(v))
		if err != nil {
			return err
		}
		*ts = parsed
		return nil

	default:
		return fmt.Errorf("cannot scan %T into Timestamp", value)
	}
}

// Value implements driver.Valuer.
// Returns time.Time for database storage.
func (ts Timestamp) Value() (driver.Value, error) {
	if ts.IsZero() {
		return nil, nil
	}
	return ts.value, nil
}

// ============================================================================
// Text Marshaling
// ============================================================================

// MarshalText implements encoding.TextMarshaler.
func (ts Timestamp) MarshalText() ([]byte, error) {
	if ts.IsZero() {
		return []byte{}, nil
	}
	return []byte(ts.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (ts *Timestamp) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*ts = Timestamp{}
		return nil
	}

	parsed, err := ParseTimestamp(string(text))
	if err != nil {
		return err
	}

	*ts = parsed
	return nil
}

// ============================================================================
// Helpers
// ============================================================================

// GoString implements fmt.GoStringer for debugging.
func (ts Timestamp) GoString() string {
	if ts.IsZero() {
		return "Timestamp{zero}"
	}
	return fmt.Sprintf("Timestamp{%s}", ts.String())
}

// Format formats the timestamp using a custom layout.
// Uses Go's time.Format layout syntax.
//
// Example:
//
//	ts.Format("2006-01-02")           // "2024-01-15"
//	ts.Format("15:04:05")             // "10:30:45"
//	ts.Format("Jan 02, 2006 3:04 PM") // "Jan 15, 2024 10:30 AM"
func (ts Timestamp) Format(layout string) string {
	if ts.IsZero() {
		return ""
	}
	return ts.value.Format(layout)
}

// Date returns the year, month, and day of the timestamp.
func (ts Timestamp) Date() (year int, month time.Month, day int) {
	return ts.value.Date()
}

// Clock returns the hour, minute, and second of the timestamp.
func (ts Timestamp) Clock() (hour, min, sec int) {
	return ts.value.Clock()
}

// Year returns the year.
func (ts Timestamp) Year() int {
	return ts.value.Year()
}

// Month returns the month.
func (ts Timestamp) Month() time.Month {
	return ts.value.Month()
}

// Day returns the day of the month.
func (ts Timestamp) Day() int {
	return ts.value.Day()
}

// Weekday returns the day of the week.
func (ts Timestamp) Weekday() time.Weekday {
	return ts.value.Weekday()
}

// ============================================================================
// Null Timestamp Support
// ============================================================================

// NullTimestamp represents a timestamp that can be null/nil.
// Used for optional timestamp fields in databases.
type NullTimestamp struct {
	Timestamp Timestamp
	Valid     bool // Valid is true if Timestamp is not null
}

// NewNullTimestamp creates a valid NullTimestamp.
func NewNullTimestamp(ts Timestamp) NullTimestamp {
	return NullTimestamp{
		Timestamp: ts,
		Valid:     true,
	}
}

// Scan implements sql.Scanner for NullTimestamp.
func (nt *NullTimestamp) Scan(value interface{}) error {
	if value == nil {
		nt.Timestamp = Timestamp{}
		nt.Valid = false
		return nil
	}

	if err := nt.Timestamp.Scan(value); err != nil {
		return err
	}

	nt.Valid = true
	return nil
}

// Value implements driver.Valuer for NullTimestamp.
func (nt NullTimestamp) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Timestamp.Value()
}

// MarshalJSON implements json.Marshaler for NullTimestamp.
func (nt NullTimestamp) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return []byte("null"), nil
	}
	return nt.Timestamp.MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler for NullTimestamp.
func (nt *NullTimestamp) UnmarshalJSON(data []byte) error {
	var s *string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	if s == nil || *s == "" {
		nt.Valid = false
		nt.Timestamp = Timestamp{}
		return nil
	}

	ts, err := ParseTimestamp(*s)
	if err != nil {
		return err
	}

	nt.Timestamp = ts
	nt.Valid = true
	return nil
}
