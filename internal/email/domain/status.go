package domain

import "fmt"

// Status represents the status of an email template.
type Status string

const (
	StatusDraft    Status = "draft"
	StatusActive   Status = "active"
	StatusArchived Status = "archived"
)

// String returns the string representation of the status.
func (s Status) String() string {
	return string(s)
}

// Validate validates the status value.
func (s Status) Validate() error {
	switch s {
	case StatusDraft, StatusActive, StatusArchived:
		return nil
	default:
		return fmt.Errorf("invalid template status: %s", s)
	}
}

// IsValid returns true if the status is valid.
func (s Status) IsValid() bool {
	return s.Validate() == nil
}

// IsDraft returns true if the template is a draft.
func (s Status) IsDraft() bool {
	return s == StatusDraft
}

// IsActive returns true if the template is active.
func (s Status) IsActive() bool {
	return s == StatusActive
}

// IsArchived returns true if the template is archived.
func (s Status) IsArchived() bool {
	return s == StatusArchived
}

// CanActivate returns true if the template can be activated from this status.
func (s Status) CanActivate() bool {
	return s == StatusDraft || s == StatusArchived
}

// CanArchive returns true if the template can be archived from this status.
func (s Status) CanArchive() bool {
	return s == StatusActive || s == StatusDraft
}

// CanEdit returns true if the template can be edited in this status.
func (s Status) CanEdit() bool {
	return s == StatusDraft
}

// ParseStatus parses a string into a Status.
func ParseStatus(s string) (Status, error) {
	status := Status(s)
	if err := status.Validate(); err != nil {
		return "", err
	}
	return status, nil
}
