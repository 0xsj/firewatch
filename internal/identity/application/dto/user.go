package dto

import (
	"time"

	"github.com/0xsj/hexagonal-go/internal/identity/domain/user"
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

// ============================================================================
// Constructors - Convert Domain to DTO
// ============================================================================

// NewUserResponse creates a UserDTO from a User aggregate.
func NewUserResponse(u *user.User) *UserDTO {
	dto := &UserDTO{
		ID:            u.ID().String(),
		TenantID:      u.TenantID(),
		Email:         u.Email().String(),
		Status:        u.Status().String(),
		Role:          u.Role().String(),
		EmailVerified: u.EmailVerified(),
		CreatedAt:     u.CreatedAt().Time(),
		UpdatedAt:     u.UpdatedAt().Time(),
	}

	// Optional fields
	if u.EmailVerifiedAt() != nil {
		t := u.EmailVerifiedAt().Time()
		dto.EmailVerifiedAt = &t
	}

	if u.LastLoginAt() != nil {
		t := u.LastLoginAt().Time()
		dto.LastLoginAt = &t
	}

	return dto
}

// NewUserSummary creates a UserSummaryDTO from a User aggregate.
func NewUserSummary(u *user.User) *UserSummaryDTO {
	return &UserSummaryDTO{
		ID:        u.ID().String(),
		Email:     u.Email().String(),
		Status:    u.Status().String(),
		Role:      u.Role().String(),
		CreatedAt: u.CreatedAt().Time(),
	}
}

// ToUserResponse converts UserDTO to the response format.
// (Alias for consistency with API naming)
type UserResponse = UserDTO

// ToUserSummary converts UserSummaryDTO to the response format.
// (Alias for consistency with API naming)
type UserSummary = UserSummaryDTO
