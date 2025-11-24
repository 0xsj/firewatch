package domain

// Status represents the delivery status of a notification.
type Status string

const (
	StatusPending Status = "pending"
	StatusSent    Status = "sent"
	StatusFailed  Status = "failed"
)

// String returns the string representation.
func (s Status) String() string {
	return string(s)
}

// IsValid checks if the status is valid.
func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusSent, StatusFailed:
		return true
	default:
		return false
	}
}

// IsTerminal returns true if the status is final (no more transitions).
func (s Status) IsTerminal() bool {
	return s == StatusSent || s == StatusFailed
}
