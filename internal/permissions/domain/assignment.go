package domain

import (
	"time"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Assignment represents a role assignment to a user.
// It links a user to a role within a specific scope.
type Assignment struct {
	id         types.ID
	userID     types.ID
	roleID     types.ID
	tenantID   string
	resourceID string     // Optional: specific resource scope (e.g., project ID)
	expiresAt  *time.Time // Optional: when the assignment expires
	createdAt  types.Timestamp
	createdBy  string
	events     []Event
}

// ============================================================================
// Getters
// ============================================================================

// ID returns the assignment ID.
func (a *Assignment) ID() types.ID {
	return a.id
}

// UserID returns the user ID.
func (a *Assignment) UserID() types.ID {
	return a.userID
}

// RoleID returns the role ID.
func (a *Assignment) RoleID() types.ID {
	return a.roleID
}

// TenantID returns the tenant ID.
func (a *Assignment) TenantID() string {
	return a.tenantID
}

// ResourceID returns the optional resource ID scope.
func (a *Assignment) ResourceID() string {
	return a.resourceID
}

// ExpiresAt returns the optional expiration time.
func (a *Assignment) ExpiresAt() *time.Time {
	return a.expiresAt
}

// CreatedAt returns the creation timestamp.
func (a *Assignment) CreatedAt() types.Timestamp {
	return a.createdAt
}

// CreatedBy returns who created the assignment.
func (a *Assignment) CreatedBy() string {
	return a.createdBy
}

// Events returns uncommitted domain events.
func (a *Assignment) Events() []Event {
	return a.events
}

// ClearEvents clears uncommitted domain events.
func (a *Assignment) ClearEvents() {
	a.events = nil
}

// ============================================================================
// Factory Methods
// ============================================================================

// AssignRole creates a new role assignment.
// Emits RoleAssigned event.
func AssignRole(
	id types.ID,
	userID types.ID,
	roleID types.ID,
	tenantID string,
	roleName string,
	resourceID string,
	expiresAt *time.Time,
	assignedBy string,
) *Assignment {
	now := types.Now()

	assignment := &Assignment{
		id:         id,
		userID:     userID,
		roleID:     roleID,
		tenantID:   tenantID,
		resourceID: resourceID,
		expiresAt:  expiresAt,
		createdAt:  now,
		createdBy:  assignedBy,
		events:     make([]Event, 0),
	}

	assignment.addEvent(NewRoleAssigned(
		id,
		userID,
		roleID,
		tenantID,
		roleName,
		resourceID,
		expiresAt,
		assignedBy,
	))

	return assignment
}

// ReconstituteAssignment recreates an assignment from stored state.
// Does NOT emit events - only for loading from database.
func ReconstituteAssignment(
	id types.ID,
	userID types.ID,
	roleID types.ID,
	tenantID string,
	resourceID string,
	expiresAt *time.Time,
	createdAt types.Timestamp,
	createdBy string,
) *Assignment {
	return &Assignment{
		id:         id,
		userID:     userID,
		roleID:     roleID,
		tenantID:   tenantID,
		resourceID: resourceID,
		expiresAt:  expiresAt,
		createdAt:  createdAt,
		createdBy:  createdBy,
		events:     make([]Event, 0),
	}
}

// ============================================================================
// Commands
// ============================================================================

// Revoke marks the assignment as revoked.
// Emits RoleRevoked event.
func (a *Assignment) Revoke(roleName, revokedBy, reason string) {
	a.addEvent(NewRoleRevoked(
		a.id,
		a.userID,
		a.roleID,
		a.tenantID,
		roleName,
		revokedBy,
		reason,
	))
}

// ============================================================================
// Query Methods
// ============================================================================

// IsExpired checks if the assignment has expired.
func (a *Assignment) IsExpired() bool {
	if a.expiresAt == nil {
		return false
	}
	return time.Now().After(*a.expiresAt)
}

// IsActive checks if the assignment is currently active (not expired).
func (a *Assignment) IsActive() bool {
	return !a.IsExpired()
}

// MatchesScope checks if the assignment matches the given scope.
func (a *Assignment) MatchesScope(tenantID, resourceID string) bool {
	// Tenant must match (or assignment is for all tenants)
	if a.tenantID != "*" && a.tenantID != tenantID {
		return false
	}

	// If assignment has no resource restriction, it matches
	if a.resourceID == "" {
		return true
	}

	// Otherwise, resource must match
	return a.resourceID == resourceID
}

// ============================================================================
// Internal
// ============================================================================

// addEvent adds a domain event.
func (a *Assignment) addEvent(event Event) {
	a.events = append(a.events, event)
}
