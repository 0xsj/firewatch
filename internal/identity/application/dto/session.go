package dto

import (
	"time"

	"github.com/0xsj/hexagonal-go/internal/identity/domain/session"
)

// SessionDTO represents a session for API responses.
type SessionDTO struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	TenantID     string     `json:"tenant_id"`
	Provider     string     `json:"provider"`
	Status       string     `json:"status"`
	IPAddress    string     `json:"ip_address,omitempty"`
	UserAgent    string     `json:"user_agent,omitempty"`
	DeviceID     string     `json:"device_id,omitempty"`
	ExpiresAt    time.Time  `json:"expires_at"`
	CreatedAt    time.Time  `json:"created_at"`
	LastActiveAt time.Time  `json:"last_active_at"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`
}

// NewSessionResponse creates a SessionDTO from a session aggregate.
func NewSessionResponse(s *session.Session) *SessionDTO {
	return &SessionDTO{
		ID:           s.ID().String(),
		UserID:       s.UserID().String(),
		TenantID:     s.TenantID(),
		Provider:     s.Provider().String(),
		Status:       s.Status().String(),
		IPAddress:    s.IPAddress(),
		UserAgent:    s.UserAgent(),
		DeviceID:     s.DeviceID(),
		ExpiresAt:    s.ExpiresAt(),
		CreatedAt:    s.CreatedAt(),
		LastActiveAt: s.LastActiveAt(),
		RevokedAt:    s.RevokedAt(),
	}
}
