package session

import (
	"fmt"
	"time"

	"github.com/0xsj/hexagonal-go/internal/identity/domain/auth"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Status represents the session status.
type Status string

const (
	StatusActive    Status = "active"
	StatusExpired   Status = "expired"
	StatusRevoked   Status = "revoked"
	StatusLoggedOut Status = "logged_out"
)

// String returns the string representation.
func (s Status) String() string {
	return string(s)
}

// Session represents an active user session.
type Session struct {
	id           types.ID
	userID       types.ID
	tenantID     string
	provider     auth.Provider
	status       Status
	refreshToken string

	// Client info
	ipAddress string
	userAgent string
	deviceID  string

	// Timestamps
	expiresAt    time.Time
	createdAt    time.Time
	updatedAt    time.Time
	lastActiveAt time.Time
	revokedAt    *time.Time

	// Events
	events []Event
}

// New creates a new active session.
func New(
	userID types.ID,
	tenantID string,
	provider auth.Provider,
	refreshToken string,
	ipAddress string,
	userAgent string,
	ttl time.Duration,
) *Session {
	now := time.Now().UTC()
	s := &Session{
		id:           types.NewID(),
		userID:       userID,
		tenantID:     tenantID,
		provider:     provider,
		status:       StatusActive,
		refreshToken: refreshToken,
		ipAddress:    ipAddress,
		userAgent:    userAgent,
		expiresAt:    now.Add(ttl),
		createdAt:    now,
		updatedAt:    now,
		lastActiveAt: now,
		events:       make([]Event, 0),
	}

	s.events = append(s.events, NewSessionCreated(s.id, userID, tenantID, provider, ipAddress, userAgent))
	return s
}

// Getters

func (s *Session) ID() types.ID            { return s.id }
func (s *Session) UserID() types.ID        { return s.userID }
func (s *Session) TenantID() string        { return s.tenantID }
func (s *Session) Provider() auth.Provider { return s.provider }
func (s *Session) Status() Status          { return s.status }
func (s *Session) RefreshToken() string    { return s.refreshToken }
func (s *Session) IPAddress() string       { return s.ipAddress }
func (s *Session) UserAgent() string       { return s.userAgent }
func (s *Session) DeviceID() string        { return s.deviceID }
func (s *Session) ExpiresAt() time.Time    { return s.expiresAt }
func (s *Session) CreatedAt() time.Time    { return s.createdAt }
func (s *Session) UpdatedAt() time.Time    { return s.updatedAt }
func (s *Session) LastActiveAt() time.Time { return s.lastActiveAt }
func (s *Session) RevokedAt() *time.Time   { return s.revokedAt }

// IsActive returns true if the session is active.
func (s *Session) IsActive() bool {
	return s.status == StatusActive && !s.IsExpired()
}

// IsExpired returns true if the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().UTC().After(s.expiresAt)
}

// Refresh refreshes the session with a new refresh token and extends expiry.
func (s *Session) Refresh(newRefreshToken string, ttl time.Duration) error {
	if !s.IsActive() {
		return fmt.Errorf("cannot refresh inactive session")
	}

	now := time.Now().UTC()
	oldToken := s.refreshToken

	s.refreshToken = newRefreshToken
	s.expiresAt = now.Add(ttl)
	s.updatedAt = now
	s.lastActiveAt = now

	s.events = append(s.events, NewSessionRefreshed(s.id, s.userID, s.tenantID, oldToken, newRefreshToken))
	return nil
}

// Touch updates the last active timestamp.
func (s *Session) Touch() {
	now := time.Now().UTC()
	s.lastActiveAt = now
	s.updatedAt = now
}

// UpdateClientInfo updates the client information.
func (s *Session) UpdateClientInfo(ipAddress, userAgent string) {
	s.ipAddress = ipAddress
	s.userAgent = userAgent
	s.updatedAt = time.Now().UTC()
}

// SetDeviceID sets the device identifier.
func (s *Session) SetDeviceID(deviceID string) {
	s.deviceID = deviceID
	s.updatedAt = time.Now().UTC()
}

// Revoke revokes the session.
func (s *Session) Revoke(reason string) error {
	if s.status == StatusRevoked {
		return fmt.Errorf("session already revoked")
	}

	now := time.Now().UTC()
	s.status = StatusRevoked
	s.revokedAt = &now
	s.updatedAt = now

	s.events = append(s.events, NewSessionRevoked(s.id, s.userID, s.tenantID, reason))
	return nil
}

// Logout logs out the session.
func (s *Session) Logout() error {
	if !s.IsActive() {
		return fmt.Errorf("session is not active")
	}

	now := time.Now().UTC()
	s.status = StatusLoggedOut
	s.updatedAt = now

	s.events = append(s.events, NewSessionLoggedOut(s.id, s.userID, s.tenantID))
	return nil
}

// Expire marks the session as expired.
func (s *Session) Expire() {
	s.status = StatusExpired
	s.updatedAt = time.Now().UTC()
}

// Events returns and clears domain events.
func (s *Session) Events() []Event {
	return s.events
}

// ClearEvents clears domain events.
func (s *Session) ClearEvents() {
	s.events = make([]Event, 0)
}

// Snapshot for persistence.
type Snapshot struct {
	ID           string
	UserID       string
	TenantID     string
	Provider     string
	Status       string
	RefreshToken string
	IPAddress    string
	UserAgent    string
	DeviceID     string
	ExpiresAt    time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastActiveAt time.Time
	RevokedAt    *time.Time
}

// ToSnapshot converts to a snapshot for persistence.
func (s *Session) ToSnapshot() Snapshot {
	return Snapshot{
		ID:           s.id.String(),
		UserID:       s.userID.String(),
		TenantID:     s.tenantID,
		Provider:     s.provider.String(),
		Status:       s.status.String(),
		RefreshToken: s.refreshToken,
		IPAddress:    s.ipAddress,
		UserAgent:    s.userAgent,
		DeviceID:     s.deviceID,
		ExpiresAt:    s.expiresAt,
		CreatedAt:    s.createdAt,
		UpdatedAt:    s.updatedAt,
		LastActiveAt: s.lastActiveAt,
		RevokedAt:    s.revokedAt,
	}
}

// FromSnapshot reconstitutes from a snapshot.
func FromSnapshot(snap Snapshot) (*Session, error) {
	id, err := types.ParseID(snap.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid session id: %w", err)
	}

	userID, err := types.ParseID(snap.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	return &Session{
		id:           id,
		userID:       userID,
		tenantID:     snap.TenantID,
		provider:     auth.Provider(snap.Provider),
		status:       Status(snap.Status),
		refreshToken: snap.RefreshToken,
		ipAddress:    snap.IPAddress,
		userAgent:    snap.UserAgent,
		deviceID:     snap.DeviceID,
		expiresAt:    snap.ExpiresAt,
		createdAt:    snap.CreatedAt,
		updatedAt:    snap.UpdatedAt,
		lastActiveAt: snap.LastActiveAt,
		revokedAt:    snap.RevokedAt,
		events:       make([]Event, 0),
	}, nil
}
