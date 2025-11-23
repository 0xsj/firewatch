package user

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Status represents the current state of a user account.
// Controls whether a user can authenticate and access the system.
type Status string

const (
	// StatusActive indicates the user account is active and can be used.
	StatusActive Status = "active"

	// StatusPending indicates the user account is awaiting email verification.
	StatusPending Status = "pending"

	// StatusSuspended indicates the user account has been suspended by an admin.
	// User cannot login but account is not deleted.
	StatusSuspended Status = "suspended"

	// StatusLocked indicates the user account is temporarily locked.
	// Usually due to multiple failed login attempts.
	StatusLocked Status = "locked"

	// StatusDeleted indicates the user account has been soft-deleted.
	// User cannot login and data may be scheduled for hard deletion.
	StatusDeleted Status = "deleted"
)

// AllStatuses returns all valid user statuses.
func AllStatuses() []Status {
	return []Status{
		StatusActive,
		StatusPending,
		StatusSuspended,
		StatusLocked,
		StatusDeleted,
	}
}

// IsValid returns true if the status is a valid user status.
func (s Status) IsValid() bool {
	switch s {
	case StatusActive, StatusPending, StatusSuspended, StatusLocked, StatusDeleted:
		return true
	default:
		return false
	}
}

// Validate returns an error if the status is invalid.
func (s Status) Validate() error {
	if !s.IsValid() {
		return fmt.Errorf("invalid user status: %s", s)
	}
	return nil
}

// String returns the string representation of the status.
func (s Status) String() string {
	return string(s)
}

// ============================================================================
// Status Checks
// ============================================================================

// IsActive returns true if the user account is active.
func (s Status) IsActive() bool {
	return s == StatusActive
}

// IsPending returns true if the user account is pending verification.
func (s Status) IsPending() bool {
	return s == StatusPending
}

// IsSuspended returns true if the user account is suspended.
func (s Status) IsSuspended() bool {
	return s == StatusSuspended
}

// IsLocked returns true if the user account is locked.
func (s Status) IsLocked() bool {
	return s == StatusLocked
}

// IsDeleted returns true if the user account is deleted.
func (s Status) IsDeleted() bool {
	return s == StatusDeleted
}

// CanLogin returns true if the user can attempt to login.
// Only active users can login.
func (s Status) CanLogin() bool {
	return s == StatusActive
}

// CanBeActivated returns true if the status can transition to active.
// Pending and locked accounts can be activated.
func (s Status) CanBeActivated() bool {
	return s == StatusPending || s == StatusLocked
}

// CanBeSuspended returns true if the status can transition to suspended.
// Only active accounts can be suspended.
func (s Status) CanBeSuspended() bool {
	return s == StatusActive
}

// CanBeLocked returns true if the status can transition to locked.
// Only active and pending accounts can be locked.
func (s Status) CanBeLocked() bool {
	return s == StatusActive || s == StatusPending
}

// CanBeDeleted returns true if the status can transition to deleted.
// Any non-deleted account can be deleted.
func (s Status) CanBeDeleted() bool {
	return s != StatusDeleted
}

// ============================================================================
// Status Transitions
// ============================================================================

// CanTransitionTo checks if the status can transition to the target status.
// Returns an error if the transition is not allowed.
func (s Status) CanTransitionTo(target Status) error {
	// Validate target status
	if err := target.Validate(); err != nil {
		return err
	}

	// No-op transition
	if s == target {
		return nil
	}

	// Check valid transitions based on business rules
	switch target {
	case StatusActive:
		if !s.CanBeActivated() {
			return fmt.Errorf("cannot activate user with status: %s", s)
		}

	case StatusSuspended:
		if !s.CanBeSuspended() {
			return fmt.Errorf("cannot suspend user with status: %s", s)
		}

	case StatusLocked:
		if !s.CanBeLocked() {
			return fmt.Errorf("cannot lock user with status: %s", s)
		}

	case StatusDeleted:
		if !s.CanBeDeleted() {
			return fmt.Errorf("cannot delete user with status: %s", s)
		}

	case StatusPending:
		// Generally, you can't transition back to pending
		// This would only happen during user creation
		return fmt.Errorf("cannot transition to pending status from: %s", s)
	}

	return nil
}

// ============================================================================
// Parsing
// ============================================================================

// ParseStatus parses a string into a Status.
// Returns an error if the string is not a valid status.
func ParseStatus(s string) (Status, error) {
	status := Status(s)
	if err := status.Validate(); err != nil {
		return "", err
	}
	return status, nil
}

// MustParseStatus parses a string into a Status and panics if invalid.
// Only use this for constants where you're certain the value is valid.
func MustParseStatus(s string) Status {
	status, err := ParseStatus(s)
	if err != nil {
		panic(fmt.Sprintf("invalid status: %v", err))
	}
	return status
}

// ============================================================================
// Database Marshaling
// ============================================================================

// Scan implements sql.Scanner for reading from database.
func (s *Status) Scan(value interface{}) error {
	if value == nil {
		*s = ""
		return nil
	}

	var str string
	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("cannot scan %T into Status", value)
	}

	status, err := ParseStatus(str)
	if err != nil {
		return err
	}

	*s = status
	return nil
}

// Value implements driver.Valuer for writing to database.
func (s Status) Value() (driver.Value, error) {
	if s == "" {
		return nil, nil
	}

	if err := s.Validate(); err != nil {
		return nil, err
	}

	return string(s), nil
}

// ============================================================================
// JSON Marshaling
// ============================================================================

// MarshalJSON implements json.Marshaler.
func (s Status) MarshalJSON() ([]byte, error) {
	if s == "" {
		return []byte("null"), nil
	}
	return fmt.Appendf(nil, "%q", s), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *Status) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	if str == "" {
		*s = ""
		return nil
	}

	status, err := ParseStatus(str)
	if err != nil {
		return err
	}

	*s = status
	return nil
}
