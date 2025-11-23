package user

import (
	"time"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Event is the base interface for all user domain events.
// All user events embed EventMetadata for common fields.
type Event interface {
	// EventType returns the event type identifier
	EventType() string

	// EventTime returns when the event occurred
	EventTime() time.Time

	// AggregateID returns the user ID
	AggregateID() types.ID

	// AggregateTenantID returns the tenant ID
	AggregateTenantID() string
}

// EventMetadata contains common fields for all domain events.
type EventMetadata struct {
	Type     string         `json:"type"`
	Time     time.Time      `json:"time"`
	UserID   types.ID       `json:"user_id"`
	TenantID string         `json:"tenant_id"`
	Version  int            `json:"version"`  // Aggregate version
	Metadata map[string]any `json:"metadata"` // Additional context
}

// EventType returns the event type.
func (m EventMetadata) EventType() string {
	return m.Type
}

// EventTime returns when the event occurred.
func (m EventMetadata) EventTime() time.Time {
	return m.Time
}

// AggregateID returns the user ID.
func (m EventMetadata) AggregateID() types.ID {
	return m.UserID
}

// AggregateTenantID returns the tenant ID.
func (m EventMetadata) AggregateTenantID() string {
	return m.TenantID
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
			Type:     EventTypeUserRegistered,
			Time:     time.Now(),
			UserID:   userID,
			TenantID: tenantID,
			Version:  1,
			Metadata: make(map[string]any),
		},
		Email:         email.String(),
		Role:          role.String(),
		HasPassword:   hasPassword,
		EmailVerified: !hasPassword, // SSO/magic link users are pre-verified
		Source:        source,
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
			Type:     EventTypeUserEmailVerified,
			Time:     time.Now(),
			UserID:   userID,
			TenantID: tenantID,
			Metadata: make(map[string]any),
		},
		Email:      email.String(),
		VerifiedAt: time.Now(),
	}
}

// ============================================================================
// Password Events
// ============================================================================

// UserPasswordChanged is emitted when a user changes their password.
type UserPasswordChanged struct {
	EventMetadata
	ChangedBy string `json:"changed_by"` // "user" or "admin"
}

// NewUserPasswordChanged creates a new UserPasswordChanged event.
func NewUserPasswordChanged(userID types.ID, tenantID string, changedBy string) UserPasswordChanged {
	return UserPasswordChanged{
		EventMetadata: EventMetadata{
			Type:     EventTypeUserPasswordChanged,
			Time:     time.Now(),
			UserID:   userID,
			TenantID: tenantID,
			Metadata: make(map[string]any),
		},
		ChangedBy: changedBy,
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
			Type:     EventTypeUserPasswordReset,
			Time:     time.Now(),
			UserID:   userID,
			TenantID: tenantID,
			Metadata: make(map[string]any),
		},
		Email:     email.String(),
		IPAddress: ipAddress,
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
			Type:     EventTypeUserLoggedIn,
			Time:     time.Now(),
			UserID:   userID,
			TenantID: tenantID,
			Metadata: make(map[string]any),
		},
		Email:     email.String(),
		Method:    method,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		SessionID: sessionID,
	}
}

// UserLoggedOut is emitted when a user logs out.
type UserLoggedOut struct {
	EventMetadata
	SessionID string `json:"session_id,omitempty"`
	Reason    string `json:"reason,omitempty"` // "user_initiated", "session_expired", "forced"
}

// NewUserLoggedOut creates a new UserLoggedOut event.
func NewUserLoggedOut(userID types.ID, tenantID string, sessionID, reason string) UserLoggedOut {
	return UserLoggedOut{
		EventMetadata: EventMetadata{
			Type:     EventTypeUserLoggedOut,
			Time:     time.Now(),
			UserID:   userID,
			TenantID: tenantID,
			Metadata: make(map[string]any),
		},
		SessionID: sessionID,
		Reason:    reason,
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
			Type:     EventTypeUserLoginFailed,
			Time:     time.Now(),
			UserID:   userID,
			TenantID: tenantID,
			Metadata: make(map[string]any),
		},
		Email:      email.String(),
		Reason:     reason,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		AttemptNum: attemptNum,
	}
}

// ============================================================================
// Account Status Events
// ============================================================================

// UserAccountLocked is emitted when a user account is locked.
type UserAccountLocked struct {
	EventMetadata
	Reason      string     `json:"reason"` // "too_many_attempts", "admin_action", "suspicious_activity"
	LockedUntil *time.Time `json:"locked_until,omitempty"`
	LockedBy    string     `json:"locked_by,omitempty"` // "system" or admin user ID
}

// NewUserAccountLocked creates a new UserAccountLocked event.
func NewUserAccountLocked(userID types.ID, tenantID string, reason string, lockedUntil *time.Time, lockedBy string) UserAccountLocked {
	return UserAccountLocked{
		EventMetadata: EventMetadata{
			Type:     EventTypeUserAccountLocked,
			Time:     time.Now(),
			UserID:   userID,
			TenantID: tenantID,
			Metadata: make(map[string]any),
		},
		Reason:      reason,
		LockedUntil: lockedUntil,
		LockedBy:    lockedBy,
	}
}

// UserAccountUnlocked is emitted when a locked account is unlocked.
type UserAccountUnlocked struct {
	EventMetadata
	UnlockedBy string `json:"unlocked_by"` // "system" or admin user ID
}

// NewUserAccountUnlocked creates a new UserAccountUnlocked event.
func NewUserAccountUnlocked(userID types.ID, tenantID string, unlockedBy string) UserAccountUnlocked {
	return UserAccountUnlocked{
		EventMetadata: EventMetadata{
			Type:     EventTypeUserAccountUnlocked,
			Time:     time.Now(),
			UserID:   userID,
			TenantID: tenantID,
			Metadata: make(map[string]any),
		},
		UnlockedBy: unlockedBy,
	}
}

// UserAccountSuspended is emitted when an account is suspended.
type UserAccountSuspended struct {
	EventMetadata
	Reason      string `json:"reason"`
	SuspendedBy string `json:"suspended_by"` // Admin user ID
}

// NewUserAccountSuspended creates a new UserAccountSuspended event.
func NewUserAccountSuspended(userID types.ID, tenantID string, reason, suspendedBy string) UserAccountSuspended {
	return UserAccountSuspended{
		EventMetadata: EventMetadata{
			Type:     EventTypeUserAccountSuspended,
			Time:     time.Now(),
			UserID:   userID,
			TenantID: tenantID,
			Metadata: make(map[string]any),
		},
		Reason:      reason,
		SuspendedBy: suspendedBy,
	}
}

// UserAccountReactivated is emitted when a suspended account is reactivated.
type UserAccountReactivated struct {
	EventMetadata
	ReactivatedBy string `json:"reactivated_by"` // Admin user ID
}

// NewUserAccountReactivated creates a new UserAccountReactivated event.
func NewUserAccountReactivated(userID types.ID, tenantID string, reactivatedBy string) UserAccountReactivated {
	return UserAccountReactivated{
		EventMetadata: EventMetadata{
			Type:     EventTypeUserAccountReactivated,
			Time:     time.Now(),
			UserID:   userID,
			TenantID: tenantID,
			Metadata: make(map[string]any),
		},
		ReactivatedBy: reactivatedBy,
	}
}

// UserAccountDeleted is emitted when an account is deleted.
type UserAccountDeleted struct {
	EventMetadata
	Reason    string `json:"reason,omitempty"`
	DeletedBy string `json:"deleted_by"` // "user" or admin user ID
}

// NewUserAccountDeleted creates a new UserAccountDeleted event.
func NewUserAccountDeleted(userID types.ID, tenantID string, reason, deletedBy string) UserAccountDeleted {
	return UserAccountDeleted{
		EventMetadata: EventMetadata{
			Type:     EventTypeUserAccountDeleted,
			Time:     time.Now(),
			UserID:   userID,
			TenantID: tenantID,
			Metadata: make(map[string]any),
		},
		Reason:    reason,
		DeletedBy: deletedBy,
	}
}

// ============================================================================
// Role & Permission Events
// ============================================================================

// UserRoleChanged is emitted when a user's role changes.
type UserRoleChanged struct {
	EventMetadata
	OldRole   string `json:"old_role"`
	NewRole   string `json:"new_role"`
	ChangedBy string `json:"changed_by"` // Admin user ID
	Reason    string `json:"reason,omitempty"`
}

// NewUserRoleChanged creates a new UserRoleChanged event.
func NewUserRoleChanged(userID types.ID, tenantID string, oldRole, newRole Role, changedBy, reason string) UserRoleChanged {
	return UserRoleChanged{
		EventMetadata: EventMetadata{
			Type:     EventTypeUserRoleChanged,
			Time:     time.Now(),
			UserID:   userID,
			TenantID: tenantID,
			Metadata: make(map[string]any),
		},
		OldRole:   oldRole.String(),
		NewRole:   newRole.String(),
		ChangedBy: changedBy,
		Reason:    reason,
	}
}

// ============================================================================
// Profile Events
// ============================================================================

// UserProfileUpdated is emitted when user profile information changes.
type UserProfileUpdated struct {
	EventMetadata
	UpdatedFields []string `json:"updated_fields"` // List of fields that changed
}

// NewUserProfileUpdated creates a new UserProfileUpdated event.
func NewUserProfileUpdated(userID types.ID, tenantID string, updatedFields []string) UserProfileUpdated {
	return UserProfileUpdated{
		EventMetadata: EventMetadata{
			Type:     EventTypeUserProfileUpdated,
			Time:     time.Now(),
			UserID:   userID,
			TenantID: tenantID,
			Metadata: make(map[string]any),
		},
		UpdatedFields: updatedFields,
	}
}
