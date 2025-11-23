package dto

import (
	"time"
)

// UserDTO is the data transfer object for user responses.
// Used for API responses and never exposes sensitive data like password hashes.
//
// Design principles:
//   - Contains only data needed for API responses
//   - Never includes domain logic
//   - Never exposes sensitive fields (password, internal state)
//   - Uses primitive types (string, int, time.Time) not domain types
type UserDTO struct {
	ID              string     `json:"id"`
	TenantID        string     `json:"tenant_id"`
	Email           string     `json:"email"`
	Status          string     `json:"status"`
	Role            string     `json:"role"`
	EmailVerified   bool       `json:"email_verified"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
}

// UserSummaryDTO is a lightweight user representation for list views.
// Contains only essential fields to reduce payload size.
type UserSummaryDTO struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}
