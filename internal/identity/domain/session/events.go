package session

import (
	"time"

	"github.com/0xsj/hexagonal-go/internal/identity/domain/auth"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Event types for session domain.
const (
	EventTypeSessionCreated   = "session.created"
	EventTypeSessionRefreshed = "session.refreshed"
	EventTypeSessionRevoked   = "session.revoked"
	EventTypeSessionLoggedOut = "session.logged_out"
	EventTypeSessionExpired   = "session.expired"
)

// Event is the interface for session domain events.
// Aligns with messaging.DomainEvent for consistent event publishing.
type Event interface {
	Type() string
	EventTime() time.Time
	AggregateID() types.ID
	AggregateTenantID() string
	Payload() map[string]any
	Version() int
}

// BaseEvent contains common event fields.
type BaseEvent struct {
	eventType string
	eventTime time.Time
	sessionID types.ID
	userID    types.ID
	tenantID  string
	version   int
}

func (e BaseEvent) Type() string              { return e.eventType }
func (e BaseEvent) EventTime() time.Time      { return e.eventTime }
func (e BaseEvent) AggregateID() types.ID     { return e.sessionID }
func (e BaseEvent) AggregateTenantID() string { return e.tenantID }
func (e BaseEvent) Version() int              { return e.version }

// ============================================================================
// Session Created
// ============================================================================

// SessionCreated is emitted when a new session is created.
type SessionCreated struct {
	BaseEvent
	Provider  auth.Provider
	IPAddress string
	UserAgent string
}

// NewSessionCreated creates a new SessionCreated event.
func NewSessionCreated(
	sessionID types.ID,
	userID types.ID,
	tenantID string,
	provider auth.Provider,
	ipAddress string,
	userAgent string,
) SessionCreated {
	return SessionCreated{
		BaseEvent: BaseEvent{
			eventType: EventTypeSessionCreated,
			eventTime: time.Now().UTC(),
			sessionID: sessionID,
			userID:    userID,
			tenantID:  tenantID,
			version:   1,
		},
		Provider:  provider,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
}

// Payload returns the event payload.
func (e SessionCreated) Payload() map[string]any {
	return map[string]any{
		"session_id": e.sessionID.String(),
		"user_id":    e.userID.String(),
		"tenant_id":  e.tenantID,
		"provider":   e.Provider.String(),
		"ip_address": e.IPAddress,
		"user_agent": e.UserAgent,
	}
}

// ============================================================================
// Session Refreshed
// ============================================================================

// SessionRefreshed is emitted when a session is refreshed.
type SessionRefreshed struct {
	BaseEvent
	OldRefreshToken string
	NewRefreshToken string
}

// NewSessionRefreshed creates a new SessionRefreshed event.
func NewSessionRefreshed(
	sessionID types.ID,
	userID types.ID,
	tenantID string,
	oldToken string,
	newToken string,
) SessionRefreshed {
	return SessionRefreshed{
		BaseEvent: BaseEvent{
			eventType: EventTypeSessionRefreshed,
			eventTime: time.Now().UTC(),
			sessionID: sessionID,
			userID:    userID,
			tenantID:  tenantID,
			version:   1,
		},
		OldRefreshToken: oldToken,
		NewRefreshToken: newToken,
	}
}

// Payload returns the event payload.
// Note: We don't include actual tokens in payload for security.
func (e SessionRefreshed) Payload() map[string]any {
	return map[string]any{
		"session_id":    e.sessionID.String(),
		"user_id":       e.userID.String(),
		"tenant_id":     e.tenantID,
		"token_rotated": true,
	}
}

// ============================================================================
// Session Revoked
// ============================================================================

// SessionRevoked is emitted when a session is revoked.
type SessionRevoked struct {
	BaseEvent
	Reason string
}

// NewSessionRevoked creates a new SessionRevoked event.
func NewSessionRevoked(
	sessionID types.ID,
	userID types.ID,
	tenantID string,
	reason string,
) SessionRevoked {
	return SessionRevoked{
		BaseEvent: BaseEvent{
			eventType: EventTypeSessionRevoked,
			eventTime: time.Now().UTC(),
			sessionID: sessionID,
			userID:    userID,
			tenantID:  tenantID,
			version:   1,
		},
		Reason: reason,
	}
}

// Payload returns the event payload.
func (e SessionRevoked) Payload() map[string]any {
	return map[string]any{
		"session_id": e.sessionID.String(),
		"user_id":    e.userID.String(),
		"tenant_id":  e.tenantID,
		"reason":     e.Reason,
	}
}

// ============================================================================
// Session Logged Out
// ============================================================================

// SessionLoggedOut is emitted when a user logs out.
type SessionLoggedOut struct {
	BaseEvent
}

// NewSessionLoggedOut creates a new SessionLoggedOut event.
func NewSessionLoggedOut(
	sessionID types.ID,
	userID types.ID,
	tenantID string,
) SessionLoggedOut {
	return SessionLoggedOut{
		BaseEvent: BaseEvent{
			eventType: EventTypeSessionLoggedOut,
			eventTime: time.Now().UTC(),
			sessionID: sessionID,
			userID:    userID,
			tenantID:  tenantID,
			version:   1,
		},
	}
}

// Payload returns the event payload.
func (e SessionLoggedOut) Payload() map[string]any {
	return map[string]any{
		"session_id": e.sessionID.String(),
		"user_id":    e.userID.String(),
		"tenant_id":  e.tenantID,
	}
}

// ============================================================================
// Session Expired
// ============================================================================

// SessionExpired is emitted when a session expires.
type SessionExpired struct {
	BaseEvent
}

// NewSessionExpired creates a new SessionExpired event.
func NewSessionExpired(
	sessionID types.ID,
	userID types.ID,
	tenantID string,
) SessionExpired {
	return SessionExpired{
		BaseEvent: BaseEvent{
			eventType: EventTypeSessionExpired,
			eventTime: time.Now().UTC(),
			sessionID: sessionID,
			userID:    userID,
			tenantID:  tenantID,
			version:   1,
		},
	}
}

// Payload returns the event payload.
func (e SessionExpired) Payload() map[string]any {
	return map[string]any{
		"session_id": e.sessionID.String(),
		"user_id":    e.userID.String(),
		"tenant_id":  e.tenantID,
	}
}
