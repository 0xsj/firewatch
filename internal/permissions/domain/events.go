package domain

import (
	"time"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Event is the base interface for all permission domain events.
type Event interface {
	// Type returns the event type identifier
	Type() string

	// EventType returns the event type identifier (alias)
	EventType() string

	// EventTime returns when the event occurred
	EventTime() time.Time

	// AggregateID returns the aggregate ID (role or assignment)
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
	ID         types.ID       `json:"id"`
	TenantID   string         `json:"tenant_id"`
	Version_   int            `json:"version"`
	Metadata   map[string]any `json:"metadata"`
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

// AggregateID returns the aggregate ID.
func (m EventMetadata) AggregateID() types.ID {
	return m.ID
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
	// Role events
	EventTypeRoleCreated           = "role.created"
	EventTypeRoleUpdated           = "role.updated"
	EventTypeRoleDeleted           = "role.deleted"
	EventTypeRolePermissionAdded   = "role.permission_added"
	EventTypeRolePermissionRemoved = "role.permission_removed"

	// Assignment events
	EventTypeRoleAssigned = "role.assigned"
	EventTypeRoleRevoked  = "role.revoked"
)

// ============================================================================
// Role Events
// ============================================================================

// RoleCreated is emitted when a new role is created.
type RoleCreated struct {
	EventMetadata
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	IsSystem    bool     `json:"is_system"`
	CreatedBy   string   `json:"created_by"`
}

// NewRoleCreated creates a new RoleCreated event.
func NewRoleCreated(roleID types.ID, tenantID, name, description string, permissions PermissionSet, isSystem bool, createdBy string) RoleCreated {
	return RoleCreated{
		EventMetadata: EventMetadata{
			EventType_: EventTypeRoleCreated,
			Time_:      time.Now(),
			ID:         roleID,
			TenantID:   tenantID,
			Version_:   1,
			Metadata:   make(map[string]any),
		},
		Name:        name,
		Description: description,
		Permissions: permissions.Strings(),
		IsSystem:    isSystem,
		CreatedBy:   createdBy,
	}
}

// Payload returns the event payload.
func (e RoleCreated) Payload() map[string]any {
	return map[string]any{
		"role_id":     e.ID.String(),
		"tenant_id":   e.TenantID,
		"name":        e.Name,
		"description": e.Description,
		"permissions": e.Permissions,
		"is_system":   e.IsSystem,
		"created_by":  e.CreatedBy,
	}
}

// RoleUpdated is emitted when a role is updated.
type RoleUpdated struct {
	EventMetadata
	Name          string   `json:"name"`
	UpdatedFields []string `json:"updated_fields"`
	UpdatedBy     string   `json:"updated_by"`
}

// NewRoleUpdated creates a new RoleUpdated event.
func NewRoleUpdated(roleID types.ID, tenantID, name string, updatedFields []string, updatedBy string, version int) RoleUpdated {
	return RoleUpdated{
		EventMetadata: EventMetadata{
			EventType_: EventTypeRoleUpdated,
			Time_:      time.Now(),
			ID:         roleID,
			TenantID:   tenantID,
			Version_:   version,
			Metadata:   make(map[string]any),
		},
		Name:          name,
		UpdatedFields: updatedFields,
		UpdatedBy:     updatedBy,
	}
}

// Payload returns the event payload.
func (e RoleUpdated) Payload() map[string]any {
	return map[string]any{
		"role_id":        e.ID.String(),
		"tenant_id":      e.TenantID,
		"name":           e.Name,
		"updated_fields": e.UpdatedFields,
		"updated_by":     e.UpdatedBy,
	}
}

// RoleDeleted is emitted when a role is deleted.
type RoleDeleted struct {
	EventMetadata
	Name      string `json:"name"`
	DeletedBy string `json:"deleted_by"`
}

// NewRoleDeleted creates a new RoleDeleted event.
func NewRoleDeleted(roleID types.ID, tenantID, name, deletedBy string, version int) RoleDeleted {
	return RoleDeleted{
		EventMetadata: EventMetadata{
			EventType_: EventTypeRoleDeleted,
			Time_:      time.Now(),
			ID:         roleID,
			TenantID:   tenantID,
			Version_:   version,
			Metadata:   make(map[string]any),
		},
		Name:      name,
		DeletedBy: deletedBy,
	}
}

// Payload returns the event payload.
func (e RoleDeleted) Payload() map[string]any {
	return map[string]any{
		"role_id":    e.ID.String(),
		"tenant_id":  e.TenantID,
		"name":       e.Name,
		"deleted_by": e.DeletedBy,
	}
}

// RolePermissionAdded is emitted when a permission is added to a role.
type RolePermissionAdded struct {
	EventMetadata
	Name       string `json:"name"`
	Permission string `json:"permission"`
	AddedBy    string `json:"added_by"`
}

// NewRolePermissionAdded creates a new RolePermissionAdded event.
func NewRolePermissionAdded(roleID types.ID, tenantID, name string, permission Permission, addedBy string, version int) RolePermissionAdded {
	return RolePermissionAdded{
		EventMetadata: EventMetadata{
			EventType_: EventTypeRolePermissionAdded,
			Time_:      time.Now(),
			ID:         roleID,
			TenantID:   tenantID,
			Version_:   version,
			Metadata:   make(map[string]any),
		},
		Name:       name,
		Permission: permission.String(),
		AddedBy:    addedBy,
	}
}

// Payload returns the event payload.
func (e RolePermissionAdded) Payload() map[string]any {
	return map[string]any{
		"role_id":    e.ID.String(),
		"tenant_id":  e.TenantID,
		"name":       e.Name,
		"permission": e.Permission,
		"added_by":   e.AddedBy,
	}
}

// RolePermissionRemoved is emitted when a permission is removed from a role.
type RolePermissionRemoved struct {
	EventMetadata
	Name       string `json:"name"`
	Permission string `json:"permission"`
	RemovedBy  string `json:"removed_by"`
}

// NewRolePermissionRemoved creates a new RolePermissionRemoved event.
func NewRolePermissionRemoved(roleID types.ID, tenantID, name string, permission Permission, removedBy string, version int) RolePermissionRemoved {
	return RolePermissionRemoved{
		EventMetadata: EventMetadata{
			EventType_: EventTypeRolePermissionRemoved,
			Time_:      time.Now(),
			ID:         roleID,
			TenantID:   tenantID,
			Version_:   version,
			Metadata:   make(map[string]any),
		},
		Name:       name,
		Permission: permission.String(),
		RemovedBy:  removedBy,
	}
}

// Payload returns the event payload.
func (e RolePermissionRemoved) Payload() map[string]any {
	return map[string]any{
		"role_id":    e.ID.String(),
		"tenant_id":  e.TenantID,
		"name":       e.Name,
		"permission": e.Permission,
		"removed_by": e.RemovedBy,
	}
}

// ============================================================================
// Assignment Events
// ============================================================================

// RoleAssigned is emitted when a role is assigned to a user.
type RoleAssigned struct {
	EventMetadata
	UserID     types.ID   `json:"user_id"`
	RoleID     types.ID   `json:"role_id"`
	RoleName   string     `json:"role_name"`
	ResourceID string     `json:"resource_id,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	AssignedBy string     `json:"assigned_by"`
}

// NewRoleAssigned creates a new RoleAssigned event.
func NewRoleAssigned(assignmentID, userID, roleID types.ID, tenantID, roleName, resourceID string, expiresAt *time.Time, assignedBy string) RoleAssigned {
	return RoleAssigned{
		EventMetadata: EventMetadata{
			EventType_: EventTypeRoleAssigned,
			Time_:      time.Now(),
			ID:         assignmentID,
			TenantID:   tenantID,
			Version_:   1,
			Metadata:   make(map[string]any),
		},
		UserID:     userID,
		RoleID:     roleID,
		RoleName:   roleName,
		ResourceID: resourceID,
		ExpiresAt:  expiresAt,
		AssignedBy: assignedBy,
	}
}

// Payload returns the event payload.
func (e RoleAssigned) Payload() map[string]any {
	payload := map[string]any{
		"assignment_id": e.ID.String(),
		"user_id":       e.UserID.String(),
		"role_id":       e.RoleID.String(),
		"role_name":     e.RoleName,
		"tenant_id":     e.TenantID,
		"assigned_by":   e.AssignedBy,
	}
	if e.ResourceID != "" {
		payload["resource_id"] = e.ResourceID
	}
	if e.ExpiresAt != nil {
		payload["expires_at"] = e.ExpiresAt
	}
	return payload
}

// RoleRevoked is emitted when a role is revoked from a user.
type RoleRevoked struct {
	EventMetadata
	UserID    types.ID `json:"user_id"`
	RoleID    types.ID `json:"role_id"`
	RoleName  string   `json:"role_name"`
	RevokedBy string   `json:"revoked_by"`
	Reason    string   `json:"reason,omitempty"`
}

// NewRoleRevoked creates a new RoleRevoked event.
func NewRoleRevoked(assignmentID, userID, roleID types.ID, tenantID, roleName, revokedBy, reason string) RoleRevoked {
	return RoleRevoked{
		EventMetadata: EventMetadata{
			EventType_: EventTypeRoleRevoked,
			Time_:      time.Now(),
			ID:         assignmentID,
			TenantID:   tenantID,
			Version_:   1,
			Metadata:   make(map[string]any),
		},
		UserID:    userID,
		RoleID:    roleID,
		RoleName:  roleName,
		RevokedBy: revokedBy,
		Reason:    reason,
	}
}

// Payload returns the event payload.
func (e RoleRevoked) Payload() map[string]any {
	payload := map[string]any{
		"assignment_id": e.ID.String(),
		"user_id":       e.UserID.String(),
		"role_id":       e.RoleID.String(),
		"role_name":     e.RoleName,
		"tenant_id":     e.TenantID,
		"revoked_by":    e.RevokedBy,
	}
	if e.Reason != "" {
		payload["reason"] = e.Reason
	}
	return payload
}
