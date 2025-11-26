package tenant

import "fmt"

// Status represents the lifecycle state of a tenant.
type Status string

const (
	// StatusTrialing indicates the tenant is in a trial period.
	StatusTrialing Status = "trialing"

	// StatusActive indicates the tenant is active and in good standing.
	StatusActive Status = "active"

	// StatusSuspended indicates the tenant has been suspended.
	// Typically due to billing issues or policy violations.
	StatusSuspended Status = "suspended"

	// StatusCancelled indicates the tenant has cancelled their subscription.
	// Data is retained for a grace period before deletion.
	StatusCancelled Status = "cancelled"

	// StatusDeleted indicates the tenant has been soft-deleted.
	StatusDeleted Status = "deleted"
)

// String returns the string representation of the status.
func (s Status) String() string {
	return string(s)
}

// Validate checks if the status is a valid value.
func (s Status) Validate() error {
	switch s {
	case StatusTrialing, StatusActive, StatusSuspended, StatusCancelled, StatusDeleted:
		return nil
	default:
		return fmt.Errorf("invalid tenant status: %s", s)
	}
}

// IsActive returns true if the tenant can operate normally.
func (s Status) IsActive() bool {
	return s == StatusActive || s == StatusTrialing
}

// IsAccessible returns true if the tenant's data can be accessed.
// Suspended and cancelled tenants may have read-only access.
func (s Status) IsAccessible() bool {
	return s != StatusDeleted
}

// CanTransitionTo validates if a status transition is allowed.
//
// Valid transitions:
//
//	trialing  → active, suspended, cancelled, deleted
//	active    → suspended, cancelled, deleted
//	suspended → active, cancelled, deleted
//	cancelled → deleted
//	deleted   → (none)
func (s Status) CanTransitionTo(target Status) error {
	if s == target {
		return nil // No-op transitions are allowed
	}

	allowed := s.allowedTransitions()
	for _, t := range allowed {
		if t == target {
			return nil
		}
	}

	return fmt.Errorf("cannot transition from %s to %s", s, target)
}

// allowedTransitions returns the list of valid target statuses.
func (s Status) allowedTransitions() []Status {
	switch s {
	case StatusTrialing:
		return []Status{StatusActive, StatusSuspended, StatusCancelled, StatusDeleted}
	case StatusActive:
		return []Status{StatusSuspended, StatusCancelled, StatusDeleted}
	case StatusSuspended:
		return []Status{StatusActive, StatusCancelled, StatusDeleted}
	case StatusCancelled:
		return []Status{StatusDeleted}
	case StatusDeleted:
		return []Status{} // Terminal state
	default:
		return []Status{}
	}
}

// AllStatuses returns all valid tenant statuses.
func AllStatuses() []Status {
	return []Status{
		StatusTrialing,
		StatusActive,
		StatusSuspended,
		StatusCancelled,
		StatusDeleted,
	}
}

// ParseStatus parses a string into a Status.
func ParseStatus(s string) (Status, error) {
	status := Status(s)
	if err := status.Validate(); err != nil {
		return "", err
	}
	return status, nil
}