package user

import (
	"fmt"
	"time"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// User is the aggregate root for user identity.
// Encapsulates all business logic related to user accounts.
//
// Design principles:
//   - Aggregate root enforces invariants
//   - No public setters - use methods that express intent
//   - Collects domain events for publishing
//   - Supports both password and passwordless authentication
type User struct {
	// Identity
	id       types.ID
	tenantID string
	email    Email

	// Authentication (optional for passwordless users)
	password *Password

	// Account state
	status Status
	role   Role

	// Email verification
	emailVerified   bool
	emailVerifiedAt *types.Timestamp

	// Account security
	failedLoginAttempts int
	lockedUntil         *time.Time

	// Timestamps
	createdAt          types.Timestamp
	updatedAt          types.Timestamp
	lastLoginAt        *types.Timestamp
	lastLoginIP        *string
	lastLoginUserAgent *string

	// Domain events (uncommitted)
	events []Event

	// Aggregate version (for optimistic locking)
	version int
}

// ============================================================================
// Aggregate Getters (Read-only access)
// ============================================================================

func (u *User) ID() types.ID                      { return u.id }
func (u *User) TenantID() string                  { return u.tenantID }
func (u *User) Email() Email                      { return u.email }
func (u *User) Password() *Password               { return u.password }
func (u *User) Status() Status                    { return u.status }
func (u *User) Role() Role                        { return u.role }
func (u *User) EmailVerified() bool               { return u.emailVerified }
func (u *User) EmailVerifiedAt() *types.Timestamp { return u.emailVerifiedAt }
func (u *User) FailedLoginAttempts() int          { return u.failedLoginAttempts }
func (u *User) LockedUntil() *time.Time           { return u.lockedUntil }
func (u *User) CreatedAt() types.Timestamp        { return u.createdAt }
func (u *User) UpdatedAt() types.Timestamp        { return u.updatedAt }
func (u *User) LastLoginAt() *types.Timestamp     { return u.lastLoginAt }
func (u *User) LastLoginIP() *string              { return u.lastLoginIP }
func (u *User) LastLoginUserAgent() *string       { return u.lastLoginUserAgent }
func (u *User) Version() int                      { return u.version }

// HasPassword returns true if the user has password authentication enabled.
func (u *User) HasPassword() bool {
	return u.password != nil
}

// IsLocked returns true if the account is currently locked.
func (u *User) IsLocked() bool {
	if u.lockedUntil == nil {
		return false
	}
	return time.Now().Before(*u.lockedUntil)
}

// ============================================================================
// Factory Methods (Aggregate Creation)
// ============================================================================

// Register creates a new user with password authentication.
// Emits UserRegistered event.
//
// Example:
//
//	user, err := user.Register(
//	    id, tenantID, email, password,
//	    user.RoleUser,
//	)
func Register(
	id types.ID,
	tenantID string,
	email Email,
	password *Password,
	role Role,
) (*User, error) {
	const op = "user.Register"

	// Validate inputs
	if id.IsEmpty() {
		return nil, fmt.Errorf("%s: user id is required", op)
	}
	if tenantID == "" {
		return nil, fmt.Errorf("%s: tenant id is required", op)
	}
	if email.IsEmpty() {
		return nil, fmt.Errorf("%s: email is required", op)
	}
	if password == nil {
		return nil, fmt.Errorf("%s: password is required", op)
	}
	if err := role.Validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	now := types.Now()

	user := &User{
		id:                  id,
		tenantID:            tenantID,
		email:               email,
		password:            password,
		status:              StatusPending, // Requires email verification
		role:                role,
		emailVerified:       false,
		emailVerifiedAt:     nil,
		failedLoginAttempts: 0,
		lockedUntil:         nil,
		createdAt:           now,
		updatedAt:           now,
		lastLoginAt:         nil,
		lastLoginIP:         nil,
		lastLoginUserAgent:  nil,
		events:              make([]Event, 0),
		version:             1,
	}

	// Emit domain event
	user.addEvent(NewUserRegistered(
		id, tenantID, email, role,
		true, // hasPassword
		"password",
	))

	return user, nil
}

// RegisterPasswordless creates a new user without password authentication.
// Used for magic link, OAuth, or SSO authentication.
// Emits UserRegistered event.
//
// Example:
//
//	user, err := user.RegisterPasswordless(
//	    id, tenantID, email,
//	    user.RoleUser,
//	    "magic_link", // or "google", "github", etc.
//	)
func RegisterPasswordless(
	id types.ID,
	tenantID string,
	email Email,
	role Role,
	source string, // "magic_link", "google", "github", etc.
) (*User, error) {
	const op = "user.RegisterPasswordless"

	// Validate inputs
	if id.IsEmpty() {
		return nil, fmt.Errorf("%s: user id is required", op)
	}
	if tenantID == "" {
		return nil, fmt.Errorf("%s: tenant id is required", op)
	}
	if email.IsEmpty() {
		return nil, fmt.Errorf("%s: email is required", op)
	}
	if err := role.Validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if source == "" {
		return nil, fmt.Errorf("%s: source is required", op)
	}

	now := types.Now()

	user := &User{
		id:                  id,
		tenantID:            tenantID,
		email:               email,
		password:            nil,          // No password
		status:              StatusActive, // Passwordless users are active immediately
		role:                role,
		emailVerified:       true, // Assume verified (OAuth/SSO)
		emailVerifiedAt:     &now,
		failedLoginAttempts: 0,
		lockedUntil:         nil,
		createdAt:           now,
		updatedAt:           now,
		lastLoginAt:         nil,
		lastLoginIP:         nil,
		lastLoginUserAgent:  nil,
		events:              make([]Event, 0),
		version:             1,
	}

	// Emit domain event
	user.addEvent(NewUserRegistered(
		id, tenantID, email, role,
		false, // hasPassword
		source,
	))

	return user, nil
}

// Reconstitute recreates a user from stored state (used by repository).
// Does NOT emit events - only for loading from database.
func Reconstitute(
	id types.ID,
	tenantID string,
	email Email,
	password *Password,
	status Status,
	role Role,
	emailVerified bool,
	emailVerifiedAt *types.Timestamp,
	failedLoginAttempts int,
	lockedUntil *time.Time,
	createdAt types.Timestamp,
	updatedAt types.Timestamp,
	lastLoginAt *types.Timestamp,
	version int,
) *User {
	return &User{
		id:                  id,
		tenantID:            tenantID,
		email:               email,
		password:            password,
		status:              status,
		role:                role,
		emailVerified:       emailVerified,
		emailVerifiedAt:     emailVerifiedAt,
		failedLoginAttempts: failedLoginAttempts,
		lockedUntil:         lockedUntil,
		createdAt:           createdAt,
		updatedAt:           updatedAt,
		lastLoginAt:         lastLoginAt,
		lastLoginIP:         nil,
		lastLoginUserAgent:  nil,
		events:              make([]Event, 0),
		version:             version,
	}
}

// ============================================================================
// Email Verification
// ============================================================================

// VerifyEmail marks the user's email as verified.
// Activates pending accounts.
// Emits UserEmailVerified event.
func (u *User) VerifyEmail() error {
	const op = "user.VerifyEmail"

	// Already verified
	if u.emailVerified {
		return nil
	}

	// Mark as verified
	u.emailVerified = true
	now := types.Now()
	u.emailVerifiedAt = &now
	u.updatedAt = now

	// Activate pending accounts
	if u.status == StatusPending {
		u.status = StatusActive
	}

	// Emit event
	u.addEvent(NewUserEmailVerified(u.id, u.tenantID, u.email))

	return nil
}

// ============================================================================
// Authentication
// ============================================================================

// Authenticate verifies the provided password.
// Returns error if authentication fails.
// Handles failed attempt tracking and account locking.
func (u *User) Authenticate(password string, ipAddress, userAgent string) error {
	const op = "user.Authenticate"

	// Check if account can login
	if err := u.canLogin(); err != nil {
		return err
	}

	// Check if user has password authentication
	if u.password == nil {
		u.recordFailedLogin(ipAddress, userAgent, "no_password_auth")
		return ErrInvalidCredentials(op)
	}

	// Verify password
	if !u.password.Matches(password) {
		u.recordFailedLogin(ipAddress, userAgent, "invalid_password")
		return ErrInvalidCredentials(op)
	}

	// Authentication successful
	u.recordSuccessfulLogin(ipAddress, userAgent, "password")

	return nil
}

// AuthenticatePasswordless authenticates via passwordless method (magic link, OAuth).
// Skips password verification.
func (u *User) AuthenticatePasswordless(method, ipAddress, userAgent string) error {
	// Check if account can login
	if err := u.canLogin(); err != nil {
		return err
	}

	// Record successful login
	u.recordSuccessfulLogin(ipAddress, userAgent, method)

	return nil
}

// canLogin checks if the user can attempt to login.
func (u *User) canLogin() error {
	const op = "user.canLogin"

	// Check status
	if !u.status.CanLogin() {
		switch u.status {
		case StatusPending:
			return ErrEmailNotVerified(op, u.email.String())
		case StatusLocked:
			if u.IsLocked() {
				return ErrAccountLocked(op, u.lockedUntil.Format(time.RFC3339))
			}
			// Lock expired, unlock account
			u.unlock()
		case StatusSuspended:
			return ErrAccountSuspended(op, "account suspended by administrator")
		case StatusDeleted:
			return ErrUserNotFound(op, u.id.String())
		default:
			return fmt.Errorf("%s: account status does not allow login: %s", op, u.status)
		}
	}

	// Check if locked
	if u.IsLocked() {
		return ErrAccountLocked(op, u.lockedUntil.Format(time.RFC3339))
	}

	return nil
}

// recordSuccessfulLogin records a successful authentication.
func (u *User) recordSuccessfulLogin(ipAddress, userAgent, method string) {
	// Reset failed attempts
	u.failedLoginAttempts = 0

	// Update last login info
	now := types.Now()
	u.lastLoginAt = &now
	u.lastLoginIP = &ipAddress
	u.lastLoginUserAgent = &userAgent
	u.updatedAt = now

	// Emit event
	u.addEvent(NewUserLoggedIn(
		u.id, u.tenantID, u.email,
		method, ipAddress, userAgent, "",
	))
}

// recordFailedLogin records a failed authentication attempt.
func (u *User) recordFailedLogin(ipAddress, userAgent, reason string) {
	u.failedLoginAttempts++
	u.updatedAt = types.Now()

	// Lock account after too many attempts (configurable threshold)
	const maxAttempts = 5
	if u.failedLoginAttempts >= maxAttempts {
		u.lock("too_many_attempts", 15*time.Minute)
	}

	// Emit event
	u.addEvent(NewUserLoginFailed(
		u.id, u.tenantID, u.email,
		reason, ipAddress, userAgent,
		u.failedLoginAttempts,
	))
}

// ============================================================================
// Password Management
// ============================================================================

// ChangePassword changes the user's password.
// Emits UserPasswordChanged event.
func (u *User) ChangePassword(oldPassword string, newPassword *Password, changedBy string) error {
	const op = "user.ChangePassword"

	// If user has password, verify old password
	if u.password != nil {
		if !u.password.Matches(oldPassword) {
			return ErrPasswordInvalid(op)
		}
	}

	// Set new password
	u.password = newPassword
	u.updatedAt = types.Now()

	// Emit event
	u.addEvent(NewUserPasswordChanged(u.id, u.tenantID, u.email, changedBy))

	return nil
}

// ResetPassword resets the password (via email reset flow).
// Does not require old password.
// Emits UserPasswordReset event.
func (u *User) ResetPassword(newPassword *Password, ipAddress string) error {
	u.password = newPassword
	u.failedLoginAttempts = 0 // Reset failed attempts
	u.updatedAt = types.Now()

	// Unlock if locked
	if u.IsLocked() {
		u.unlock()
	}

	// Emit event
	u.addEvent(NewUserPasswordReset(u.id, u.tenantID, u.email, ipAddress))

	return nil
}

// SetPassword sets a password for a passwordless user.
// Allows adding password authentication to passwordless accounts.
func (u *User) SetPassword(password *Password) error {
	const op = "user.SetPassword"

	if password == nil {
		return fmt.Errorf("%s: password is required", op)
	}

	u.password = password
	u.updatedAt = types.Now()

	// Emit event
	u.addEvent(NewUserPasswordChanged(u.id, u.tenantID, u.email, "user"))

	return nil
}

// RemovePassword removes password authentication.
// Converts to passwordless-only account.
func (u *User) RemovePassword() error {
	const op = "user.RemovePassword"

	if u.password == nil {
		return nil // Already passwordless
	}

	u.password = nil
	u.updatedAt = types.Now()

	return nil
}

// ============================================================================
// Account Status Management
// ============================================================================

// Suspend suspends the user account.
// Emits UserAccountSuspended event.
func (u *User) Suspend(reason, suspendedBy string) error {
	const op = "user.Suspend"

	if err := u.status.CanTransitionTo(StatusSuspended); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	u.status = StatusSuspended
	u.updatedAt = types.Now()

	// Emit event
	u.addEvent(NewUserAccountSuspended(u.id, u.tenantID, u.email, reason, suspendedBy))

	return nil
}

// Reactivate reactivates a suspended account.
// Emits UserAccountReactivated event.
func (u *User) Reactivate(reactivatedBy string) error {
	const op = "user.Reactivate"

	if u.status != StatusSuspended {
		return fmt.Errorf("%s: only suspended accounts can be reactivated", op)
	}

	u.status = StatusActive
	u.updatedAt = types.Now()

	// Emit event
	u.addEvent(NewUserAccountReactivated(u.id, u.tenantID, u.email, reactivatedBy))

	return nil
}

// lock locks the account for a duration.
func (u *User) lock(reason string, duration time.Duration) {
	lockedUntil := time.Now().Add(duration)
	u.status = StatusLocked
	u.lockedUntil = &lockedUntil
	u.updatedAt = types.Now()

	// Emit event
	u.addEvent(NewUserAccountLocked(u.id, u.tenantID, u.email, reason, &lockedUntil, "system"))
}

// unlock unlocks the account.
func (u *User) unlock() {
	if u.status != StatusLocked {
		return
	}

	u.status = StatusActive
	u.lockedUntil = nil
	u.failedLoginAttempts = 0
	u.updatedAt = types.Now()

	// Emit event
	u.addEvent(NewUserAccountUnlocked(u.id, u.tenantID, u.email, "system"))
}

// Delete soft-deletes the user account.
// Emits UserAccountDeleted event.
func (u *User) Delete(reason, deletedBy string) error {
	const op = "user.Delete"

	if err := u.status.CanTransitionTo(StatusDeleted); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	u.status = StatusDeleted
	u.updatedAt = types.Now()

	// Emit event
	u.addEvent(NewUserAccountDeleted(u.id, u.tenantID, u.email, reason, deletedBy))

	return nil
}

// ============================================================================
// Role Management
// ============================================================================

// ChangeRole changes the user's role.
// Emits UserRoleChanged event.
func (u *User) ChangeRole(newRole Role, changedBy, reason string) error {
	const op = "user.ChangeRole"

	if err := newRole.Validate(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if u.role == newRole {
		return nil // No change
	}

	oldRole := u.role
	u.role = newRole
	u.updatedAt = types.Now()

	// Emit event
	u.addEvent(NewUserRoleChanged(u.id, u.tenantID, u.email, oldRole, newRole, changedBy, reason))

	return nil
}

// ============================================================================
// Event Management
// ============================================================================

// Events returns all uncommitted domain events.
func (u *User) Events() []Event {
	return u.events
}

// ClearEvents clears all uncommitted events.
// Called after events are published.
func (u *User) ClearEvents() {
	u.events = make([]Event, 0)
}

// addEvent adds a domain event to the uncommitted events list.
func (u *User) addEvent(event Event) {
	u.events = append(u.events, event)
}

// ============================================================================
// Version Management (Optimistic Locking)
// ============================================================================

// IncrementVersion increments the aggregate version.
// Used for optimistic locking in the repository.
func (u *User) IncrementVersion() {
	u.version++
	u.updatedAt = types.Now()
}

// ============================================================================
// Snapshot for Persistence
// ============================================================================

// Snapshot represents the complete state of a User.
// Used for persistence and reconstruction.
type Snapshot struct {
	ID                  string
	TenantID            string
	Email               string
	PasswordHash        *string
	Status              string
	Role                string
	EmailVerified       bool
	EmailVerifiedAt     *types.Timestamp
	FailedLoginAttempts int
	LockedUntil         *time.Time
	LastLoginAt         *types.Timestamp
	LastLoginIP         *string
	LastLoginUserAgent  *string
	CreatedAt           types.Timestamp
	UpdatedAt           types.Timestamp
}

// Snapshot returns a snapshot of the user's current state.
func (u *User) Snapshot() Snapshot {
	return Snapshot{
		ID:                  u.id.String(),
		TenantID:            u.tenantID,
		Email:               u.email.String(),
		PasswordHash:        u.passwordHash(),
		Status:              string(u.status),
		Role:                string(u.role),
		EmailVerified:       u.emailVerified,
		EmailVerifiedAt:     u.emailVerifiedAt,
		FailedLoginAttempts: u.failedLoginAttempts,
		LockedUntil:         u.lockedUntil,
		LastLoginAt:         u.lastLoginAt,
		LastLoginIP:         u.lastLoginIP,
		LastLoginUserAgent:  u.lastLoginUserAgent,
		CreatedAt:           u.createdAt,
		UpdatedAt:           u.updatedAt,
	}
}

// passwordHash returns the password hash or nil if passwordless.
func (u *User) passwordHash() *string {
	if u.password == nil {
		return nil
	}
	hash := u.password.Hash()
	return &hash
}

// LoadFromSnapshot reconstructs a User from a snapshot.
func LoadFromSnapshot(snapshot Snapshot) (*User, error) {
	const op = "User.LoadFromSnapshot"

	id, err := types.ParseID(snapshot.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid ID: %w", op, err)
	}

	email, err := NewEmail(snapshot.Email)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var password *Password
	if snapshot.PasswordHash != nil {
		password = NewPasswordFromHash(*snapshot.PasswordHash)
	}

	status := Status(snapshot.Status)
	role := Role(snapshot.Role)

	user := &User{
		id:                  id,
		tenantID:            snapshot.TenantID,
		email:               email,
		password:            password,
		status:              status,
		role:                role,
		emailVerified:       snapshot.EmailVerified,
		emailVerifiedAt:     snapshot.EmailVerifiedAt,
		failedLoginAttempts: snapshot.FailedLoginAttempts,
		lockedUntil:         snapshot.LockedUntil,
		lastLoginAt:         snapshot.LastLoginAt,
		lastLoginIP:         snapshot.LastLoginIP,
		lastLoginUserAgent:  snapshot.LastLoginUserAgent,
		createdAt:           snapshot.CreatedAt,
		updatedAt:           snapshot.UpdatedAt,
		events:              make([]Event, 0),
		version:             1,
	}

	return user, nil
}
