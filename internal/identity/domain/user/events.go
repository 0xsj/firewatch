package user

import (
	"time"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Event is the base interface for all user domain events.
// All user events embed EventMetadata for common fields.
type Event interface {
	// Type returns the event type identifier
	Type() string

	// EventType returns the event type identifier (alias)
	EventType() string

	// EventTime returns when the event occurred
	EventTime() time.Time

	// AggregateID returns the user ID
	AggregateID() types.ID

	// AggregateTenantID returns the tenant ID
	AggregateTenantID() string

	// Payload returns the event data as a map
	Payload() map[string]any

	// Version returns the aggregate version
	Version() int
}

// EventMetadata contains common fields for all domain events.
type EventMetadata struct {
	EventType_ string         `json:"type"`
	Time_      time.Time      `json:"time"`
	UserID     types.ID       `json:"user_id"`
	TenantID   string         `json:"tenant_id"`
	Version_   int            `json:"version"`  // Aggregate version
	Metadata   map[string]any `json:"metadata"` // Additional context
}

// Type returns the event type.
func (m EventMetadata) Type() string {
	return m.EventType_
}

// EventType returns the event type.
func (m EventMetadata) EventType() string {
	return m.EventType_
}

// EventTime returns when the event occurred.
func (m EventMetadata) EventTime() time.Time {
	return m.Time_
}

// AggregateID returns the user ID.
func (m EventMetadata) AggregateID() types.ID {
	return m.UserID
}

// AggregateTenantID returns the tenant ID.
func (m EventMetadata) AggregateTenantID() string {
	return m.TenantID
}

// Version returns the aggregate version.
func (m EventMetadata) Version() int {
	return m.Version_
}

// ============================================================================
// Event Type Constants
// ============================================================================

const (
	EventTypeUserRegistered         = "user.registered"
	EventTypeUserEmailVerified      = "user.email_verified"
	EventTypeUserPasswordChanged    = "user.password_changed"
	EventTypeUserPasswordReset      = "user.password_reset"
	EventTypeUserLoggedIn           = "user.logged_in"
	EventTypeUserLoggedOut          = "user.logged_out"
	EventTypeUserLoginFailed        = "user.login_failed"
	EventTypeUserAccountLocked      = "user.account_locked"
	EventTypeUserAccountUnlocked    = "user.account_unlocked"
	EventTypeUserAccountSuspended   = "user.account_suspended"
	EventTypeUserAccountReactivated = "user.account_reactivated"
	EventTypeUserAccountDeleted     = "user.account_deleted"
	EventTypeUserRoleChanged        = "user.role_changed"
	EventTypeUserProfileUpdated     = "user.profile_updated"
)

// ============================================================================
// User Registration Events
// ============================================================================

// UserRegistered is emitted when a new user registers.
type UserRegistered struct {
	EventMetadata
	Email         string `json:"email"`
	Role          string `json:"role"`
	HasPassword   bool   `json:"has_password"`   // false for passwordless/SSO
	EmailVerified bool   `json:"email_verified"` // true for SSO, false for password
	Source        string `json:"source"`         // "password", "google", "magic_link", etc.
}

// NewUserRegistered creates a new UserRegistered event.
func NewUserRegistered(userID types.ID, tenantID string, email Email, role Role, hasPassword bool, source string) UserRegistered {
	return UserRegistered{
		EventMetadata: EventMetadata{
			EventType_: EventTypeUserRegistered,
			Time_:      time.Now(),
			UserID:     userID,
			TenantID:   tenantID,
			Version_:   1,
			Metadata:   make(map[string]any),
		},
		Email:         email.String(),
		Role:          role.String(),
		HasPassword:   hasPassword,
		EmailVerified: !hasPassword, // SSO/magic link users are pre-verified
		Source:        source,
	}
}

// Payload returns the event payload.
func (e UserRegistered) Payload() map[string]any {
	return map[string]any{
		"user_id":        e.UserID.String(),
		"tenant_id":      e.TenantID,
		"email":          e.Email,
		"role":           e.Role,
		"has_password":   e.HasPassword,
		"email_verified": e.EmailVerified,
		"source":         e.Source,
	}
}

// ============================================================================
// Email Verification Events
// ============================================================================

// UserEmailVerified is emitted when a user verifies their email.
type UserEmailVerified struct {
	EventMetadata
	Email      string    `json:"email"`
	VerifiedAt time.Time `json:"verified_at"`
}

// NewUserEmailVerified creates a new UserEmailVerified event.
func NewUserEmailVerified(userID types.ID, tenantID string, email Email) UserEmailVerified {
	return UserEmailVerified{
		EventMetadata: EventMetadata{
			EventType_: EventTypeUserEmailVerified,
			Time_:      time.Now(),
			UserID:     userID,
			TenantID:   tenantID,
			Metadata:   make(map[string]any),
		},
		Email:      email.String(),
		VerifiedAt: time.Now(),
	}
}

// Payload returns the event payload.
func (e UserEmailVerified) Payload() map[string]any {
	return map[string]any{
		"user_id":     e.UserID.String(),
		"tenant_id":   e.TenantID,
		"email":       e.Email,
		"verified_at": e.VerifiedAt,
	}
}

// ============================================================================
// Password Events
// ============================================================================

// UserPasswordChanged is emitted when a user changes their password.
type UserPasswordChanged struct {
	EventMetadata
	Email     string `json:"email"`
	ChangedBy string `json:"changed_by"` // "user" or "admin"
}

// NewUserPasswordChanged creates a new UserPasswordChanged event.
func NewUserPasswordChanged(userID types.ID, tenantID string, email Email, changedBy string) UserPasswordChanged {
	return UserPasswordChanged{
		EventMetadata: EventMetadata{
			EventType_: EventTypeUserPasswordChanged,
			Time_:      time.Now(),
			UserID:     userID,
			TenantID:   tenantID,
			Metadata:   make(map[string]any),
		},
		Email:     email.String(),
		ChangedBy: changedBy,
	}
}

// Payload returns the event payload.
func (e UserPasswordChanged) Payload() map[string]any {
	return map[string]any{
		"user_id":    e.UserID.String(),
		"tenant_id":  e.TenantID,
		"email":      e.Email,
		"changed_by": e.ChangedBy,
	}
}

// UserPasswordReset is emitted when a user resets their password via email.
type UserPasswordReset struct {
	EventMetadata
	Email     string `json:"email"`
	IPAddress string `json:"ip_address,omitempty"`
}

// NewUserPasswordReset creates a new UserPasswordReset event.
func NewUserPasswordReset(userID types.ID, tenantID string, email Email, ipAddress string) UserPasswordReset {
	return UserPasswordReset{
		EventMetadata: EventMetadata{
			EventType_: EventTypeUserPasswordReset,
			Time_:      time.Now(),
			UserID:     userID,
			TenantID:   tenantID,
			Metadata:   make(map[string]any),
		},
		Email:     email.String(),
		IPAddress: ipAddress,
	}
}

// Payload returns the event payload.
func (e UserPasswordReset) Payload() map[string]any {
	return map[string]any{
		"user_id":    e.UserID.String(),
		"tenant_id":  e.TenantID,
		"email":      e.Email,
		"ip_address": e.IPAddress,
	}
}

// ============================================================================
// Authentication Events
// ============================================================================

// UserLoggedIn is emitted when a user successfully logs in.
type UserLoggedIn struct {
	EventMetadata
	Email     string `json:"email"`
	Method    string `json:"method"` // "password", "magic_link", "google", etc.
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

// NewUserLoggedIn creates a new UserLoggedIn event.
func NewUserLoggedIn(userID types.ID, tenantID string, email Email, method, ipAddress, userAgent, sessionID string) UserLoggedIn {
	return UserLoggedIn{
		EventMetadata: EventMetadata{
			EventType_: EventTypeUserLoggedIn,
			Time_:      time.Now(),
			UserID:     userID,
			TenantID:   tenantID,
			Metadata:   make(map[string]any),
		},
		Email:     email.String(),
		Method:    method,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		SessionID: sessionID,
	}
}

// Payload returns the event payload.
func (e UserLoggedIn) Payload() map[string]any {
	return map[string]any{
		"user_id":    e.UserID.String(),
		"tenant_id":  e.TenantID,
		"email":      e.Email,
		"method":     e.Method,
		"ip_address": e.IPAddress,
		"user_agent": e.UserAgent,
		"session_id": e.SessionID,
	}
}

// UserLoggedOut is emitted when a user logs out.
type UserLoggedOut struct {
	EventMetadata
	Email     string `json:"email"`
	SessionID string `json:"session_id,omitempty"`
	Reason    string `json:"reason,omitempty"` // "user_initiated", "session_expired", "forced"
}

// NewUserLoggedOut creates a new UserLoggedOut event.
func NewUserLoggedOut(userID types.ID, tenantID string, email Email, sessionID, reason string) UserLoggedOut {
	return UserLoggedOut{
		EventMetadata: EventMetadata{
			EventType_: EventTypeUserLoggedOut,
			Time_:      time.Now(),
			UserID:     userID,
			TenantID:   tenantID,
			Metadata:   make(map[string]any),
		},
		Email:     email.String(),
		SessionID: sessionID,
		Reason:    reason,
	}
}

// Payload returns the event payload.
func (e UserLoggedOut) Payload() map[string]any {
	return map[string]any{
		"user_id":    e.UserID.String(),
		"tenant_id":  e.TenantID,
		"email":      e.Email,
		"session_id": e.SessionID,
		"reason":     e.Reason,
	}
}

// UserLoginFailed is emitted when a login attempt fails.
type UserLoginFailed struct {
	EventMetadata
	Email      string `json:"email"`
	Reason     string `json:"reason"` // "invalid_credentials", "account_locked", etc.
	IPAddress  string `json:"ip_address,omitempty"`
	UserAgent  string `json:"user_agent,omitempty"`
	AttemptNum int    `json:"attempt_num"` // Failed attempt number
}

// NewUserLoginFailed creates a new UserLoginFailed event.
func NewUserLoginFailed(userID types.ID, tenantID string, email Email, reason, ipAddress, userAgent string, attemptNum int) UserLoginFailed {
	return UserLoginFailed{
		EventMetadata: EventMetadata{
			EventType_: EventTypeUserLoginFailed,
			Time_:      time.Now(),
			UserID:     userID,
			TenantID:   tenantID,
			Metadata:   make(map[string]any),
		},
		Email:      email.String(),
		Reason:     reason,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		AttemptNum: attemptNum,
	}
}

// Payload returns the event payload.
func (e UserLoginFailed) Payload() map[string]any {
	return map[string]any{
		"user_id":     e.UserID.String(),
		"tenant_id":   e.TenantID,
		"email":       e.Email,
		"reason":      e.Reason,
		"ip_address":  e.IPAddress,
		"user_agent":  e.UserAgent,
		"attempt_num": e.AttemptNum,
	}
}

// ============================================================================
// Account Status Events
// ============================================================================

// UserAccountLocked is emitted when a user account is locked.
type UserAccountLocked struct {
	EventMetadata
	Email       string     `json:"email"`
	Reason      string     `json:"reason"` // "too_many_attempts", "admin_action", "suspicious_activity"
	LockedUntil *time.Time `json:"locked_until,omitempty"`
	LockedBy    string     `json:"locked_by,omitempty"` // "system" or admin user ID
}

// NewUserAccountLocked creates a new UserAccountLocked event.
func NewUserAccountLocked(userID types.ID, tenantID string, email Email, reason string, lockedUntil *time.Time, lockedBy string) UserAccountLocked {
	return UserAccountLocked{
		EventMetadata: EventMetadata{
			EventType_: EventTypeUserAccountLocked,
			Time_:      time.Now(),
			UserID:     userID,
			TenantID:   tenantID,
			Metadata:   make(map[string]any),
		},
		Email:       email.String(),
		Reason:      reason,
		LockedUntil: lockedUntil,
		LockedBy:    lockedBy,
	}
}

// Payload returns the event payload.
func (e UserAccountLocked) Payload() map[string]any {
	payload := map[string]any{
		"user_id":   e.UserID.String(),
		"tenant_id": e.TenantID,
		"email":     e.Email,
		"reason":    e.Reason,
		"locked_by": e.LockedBy,
	}
	if e.LockedUntil != nil {
		payload["locked_until"] = e.LockedUntil
	}
	return payload
}

// UserAccountUnlocked is emitted when a locked account is unlocked.
type UserAccountUnlocked struct {
	EventMetadata
	Email      string `json:"email"`
	UnlockedBy string `json:"unlocked_by"` // "system" or admin user ID
}

// NewUserAccountUnlocked creates a new UserAccountUnlocked event.
func NewUserAccountUnlocked(userID types.ID, tenantID string, email Email, unlockedBy string) UserAccountUnlocked {
	return UserAccountUnlocked{
		EventMetadata: EventMetadata{
			EventType_: EventTypeUserAccountUnlocked,
			Time_:      time.Now(),
			UserID:     userID,
			TenantID:   tenantID,
			Metadata:   make(map[string]any),
		},
		Email:      email.String(),
		UnlockedBy: unlockedBy,
	}
}

// Payload returns the event payload.
func (e UserAccountUnlocked) Payload() map[string]any {
	return map[string]any{
		"user_id":     e.UserID.String(),
		"tenant_id":   e.TenantID,
		"email":       e.Email,
		"unlocked_by": e.UnlockedBy,
	}
}

// UserAccountSuspended is emitted when an account is suspended.
type UserAccountSuspended struct {
	EventMetadata
	Email       string `json:"email"`
	Reason      string `json:"reason"`
	SuspendedBy string `json:"suspended_by"` // Admin user ID
}

// NewUserAccountSuspended creates a new UserAccountSuspended event.
func NewUserAccountSuspended(userID types.ID, tenantID string, email Email, reason, suspendedBy string) UserAccountSuspended {
	return UserAccountSuspended{
		EventMetadata: EventMetadata{
			EventType_: EventTypeUserAccountSuspended,
			Time_:      time.Now(),
			UserID:     userID,
			TenantID:   tenantID,
			Metadata:   make(map[string]any),
		},
		Email:       email.String(),
		Reason:      reason,
		SuspendedBy: suspendedBy,
	}
}

// Payload returns the event payload.
func (e UserAccountSuspended) Payload() map[string]any {
	return map[string]any{
		"user_id":      e.UserID.String(),
		"tenant_id":    e.TenantID,
		"email":        e.Email,
		"reason":       e.Reason,
		"suspended_by": e.SuspendedBy,
	}
}

// UserAccountReactivated is emitted when a suspended account is reactivated.
type UserAccountReactivated struct {
	EventMetadata
	Email         string `json:"email"`
	ReactivatedBy string `json:"reactivated_by"` // Admin user ID
}

// NewUserAccountReactivated creates a new UserAccountReactivated event.
func NewUserAccountReactivated(userID types.ID, tenantID string, email Email, reactivatedBy string) UserAccountReactivated {
	return UserAccountReactivated{
		EventMetadata: EventMetadata{
			EventType_: EventTypeUserAccountReactivated,
			Time_:      time.Now(),
			UserID:     userID,
			TenantID:   tenantID,
			Metadata:   make(map[string]any),
		},
		Email:         email.String(),
		ReactivatedBy: reactivatedBy,
	}
}

// Payload returns the event payload.
func (e UserAccountReactivated) Payload() map[string]any {
	return map[string]any{
		"user_id":        e.UserID.String(),
		"tenant_id":      e.TenantID,
		"email":          e.Email,
		"reactivated_by": e.ReactivatedBy,
	}
}

// UserAccountDeleted is emitted when an account is deleted.
type UserAccountDeleted struct {
	EventMetadata
	Email     string `json:"email"`
	Reason    string `json:"reason,omitempty"`
	DeletedBy string `json:"deleted_by"` // "user" or admin user ID
}

// NewUserAccountDeleted creates a new UserAccountDeleted event.
func NewUserAccountDeleted(userID types.ID, tenantID string, email Email, reason, deletedBy string) UserAccountDeleted {
	return UserAccountDeleted{
		EventMetadata: EventMetadata{
			EventType_: EventTypeUserAccountDeleted,
			Time_:      time.Now(),
			UserID:     userID,
			TenantID:   tenantID,
			Metadata:   make(map[string]any),
		},
		Email:     email.String(),
		Reason:    reason,
		DeletedBy: deletedBy,
	}
}

// Payload returns the event payload.
func (e UserAccountDeleted) Payload() map[string]any {
	return map[string]any{
		"user_id":    e.UserID.String(),
		"tenant_id":  e.TenantID,
		"email":      e.Email,
		"reason":     e.Reason,
		"deleted_by": e.DeletedBy,
	}
}

// ============================================================================
// Role & Permission Events
// ============================================================================

// UserRoleChanged is emitted when a user's role changes.
type UserRoleChanged struct {
	EventMetadata
	Email     string `json:"email"`
	OldRole   string `json:"old_role"`
	NewRole   string `json:"new_role"`
	ChangedBy string `json:"changed_by"` // Admin user ID
	Reason    string `json:"reason,omitempty"`
}

// NewUserRoleChanged creates a new UserRoleChanged event.
func NewUserRoleChanged(userID types.ID, tenantID string, email Email, oldRole, newRole Role, changedBy, reason string) UserRoleChanged {
	return UserRoleChanged{
		EventMetadata: EventMetadata{
			EventType_: EventTypeUserRoleChanged,
			Time_:      time.Now(),
			UserID:     userID,
			TenantID:   tenantID,
			Metadata:   make(map[string]any),
		},
		Email:     email.String(),
		OldRole:   oldRole.String(),
		NewRole:   newRole.String(),
		ChangedBy: changedBy,
		Reason:    reason,
	}
}

// Payload returns the event payload.
func (e UserRoleChanged) Payload() map[string]any {
	return map[string]any{
		"user_id":    e.UserID.String(),
		"tenant_id":  e.TenantID,
		"email":      e.Email,
		"old_role":   e.OldRole,
		"new_role":   e.NewRole,
		"changed_by": e.ChangedBy,
		"reason":     e.Reason,
	}
}

// ============================================================================
// Profile Events
// ============================================================================

// UserProfileUpdated is emitted when user profile information changes.
type UserProfileUpdated struct {
	EventMetadata
	Email         string   `json:"email"`
	UpdatedFields []string `json:"updated_fields"` // List of fields that changed
}

// NewUserProfileUpdated creates a new UserProfileUpdated event.
func NewUserProfileUpdated(userID types.ID, tenantID string, email Email, updatedFields []string) UserProfileUpdated {
	return UserProfileUpdated{
		EventMetadata: EventMetadata{
			EventType_: EventTypeUserProfileUpdated,
			Time_:      time.Now(),
			UserID:     userID,
			TenantID:   tenantID,
			Metadata:   make(map[string]any),
		},
		Email:         email.String(),
		UpdatedFields: updatedFields,
	}
}

// Payload returns the event payload.
func (e UserProfileUpdated) Payload() map[string]any {
	return map[string]any{
		"user_id":        e.UserID.String(),
		"tenant_id":      e.TenantID,
		"email":          e.Email,
		"updated_fields": e.UpdatedFields,
	}
}
