package tenant

import (
	"time"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Event is the base interface for all tenant domain events.
// All tenant events embed EventMetadata for common fields.
type Event interface {
	// Type returns the event type identifier
	Type() string

	// EventType returns the event type identifier (alias)
	EventType() string

	// EventTime returns when the event occurred
	EventTime() time.Time

	// AggregateID returns the tenant ID
	AggregateID() types.ID

	// Payload returns the event data as a map
	Payload() map[string]any

	// Version returns the aggregate version
	Version() int
}

// EventMetadata contains common fields for all domain events.
type EventMetadata struct {
	EventType_ string         `json:"type"`
	Time_      time.Time      `json:"time"`
	TenantID   types.ID       `json:"tenant_id"`
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

// AggregateID returns the tenant ID.
func (m EventMetadata) AggregateID() types.ID {
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
	EventTypeTenantCreated         = "tenant.created"
	EventTypeTenantUpdated         = "tenant.updated"
	EventTypeTenantSettingsUpdated = "tenant.settings_updated"
	EventTypeTenantSuspended       = "tenant.suspended"
	EventTypeTenantReactivated     = "tenant.reactivated"
	EventTypeTenantPlanChanged     = "tenant.plan_changed"
	EventTypeTenantDeleted         = "tenant.deleted"
)

// ============================================================================
// Tenant Created Event
// ============================================================================

// TenantCreated is emitted when a new tenant is created.
type TenantCreated struct {
	EventMetadata
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	Plan      string `json:"plan"`
	OwnerID   string `json:"owner_id"`
	CreatedBy string `json:"created_by"`
}

// NewTenantCreated creates a new TenantCreated event.
func NewTenantCreated(tenantID types.ID, slug Slug, name string, plan Plan, ownerID types.ID, createdBy string) TenantCreated {
	return TenantCreated{
		EventMetadata: EventMetadata{
			EventType_: EventTypeTenantCreated,
			Time_:      time.Now(),
			TenantID:   tenantID,
			Version_:   1,
			Metadata:   make(map[string]any),
		},
		Slug:      slug.String(),
		Name:      name,
		Plan:      plan.String(),
		OwnerID:   ownerID.String(),
		CreatedBy: createdBy,
	}
}

// Payload returns the event payload.
func (e TenantCreated) Payload() map[string]any {
	return map[string]any{
		"tenant_id":  e.TenantID.String(),
		"slug":       e.Slug,
		"name":       e.Name,
		"plan":       e.Plan,
		"owner_id":   e.OwnerID,
		"created_by": e.CreatedBy,
	}
}

// ============================================================================
// Tenant Updated Event
// ============================================================================

// TenantUpdated is emitted when tenant details are updated.
type TenantUpdated struct {
	EventMetadata
	UpdatedFields []string `json:"updated_fields"`
	UpdatedBy     string   `json:"updated_by"`
}

// NewTenantUpdated creates a new TenantUpdated event.
func NewTenantUpdated(tenantID types.ID, updatedFields []string, updatedBy string) TenantUpdated {
	return TenantUpdated{
		EventMetadata: EventMetadata{
			EventType_: EventTypeTenantUpdated,
			Time_:      time.Now(),
			TenantID:   tenantID,
			Metadata:   make(map[string]any),
		},
		UpdatedFields: updatedFields,
		UpdatedBy:     updatedBy,
	}
}

// Payload returns the event payload.
func (e TenantUpdated) Payload() map[string]any {
	return map[string]any{
		"tenant_id":      e.TenantID.String(),
		"updated_fields": e.UpdatedFields,
		"updated_by":     e.UpdatedBy,
	}
}

// ============================================================================
// Tenant Settings Updated Event
// ============================================================================

// TenantSettingsUpdated is emitted when tenant settings are updated.
type TenantSettingsUpdated struct {
	EventMetadata
	ChangedKeys []string `json:"changed_keys"`
	UpdatedBy   string   `json:"updated_by"`
}

// NewTenantSettingsUpdated creates a new TenantSettingsUpdated event.
func NewTenantSettingsUpdated(tenantID types.ID, changedKeys []string, updatedBy string) TenantSettingsUpdated {
	return TenantSettingsUpdated{
		EventMetadata: EventMetadata{
			EventType_: EventTypeTenantSettingsUpdated,
			Time_:      time.Now(),
			TenantID:   tenantID,
			Metadata:   make(map[string]any),
		},
		ChangedKeys: changedKeys,
		UpdatedBy:   updatedBy,
	}
}

// Payload returns the event payload.
func (e TenantSettingsUpdated) Payload() map[string]any {
	return map[string]any{
		"tenant_id":    e.TenantID.String(),
		"changed_keys": e.ChangedKeys,
		"updated_by":   e.UpdatedBy,
	}
}

// ============================================================================
// Tenant Suspended Event
// ============================================================================

// TenantSuspended is emitted when a tenant is suspended.
type TenantSuspended struct {
	EventMetadata
	Reason      string `json:"reason"`
	SuspendedBy string `json:"suspended_by"`
}

// NewTenantSuspended creates a new TenantSuspended event.
func NewTenantSuspended(tenantID types.ID, reason string, suspendedBy string) TenantSuspended {
	return TenantSuspended{
		EventMetadata: EventMetadata{
			EventType_: EventTypeTenantSuspended,
			Time_:      time.Now(),
			TenantID:   tenantID,
			Metadata:   make(map[string]any),
		},
		Reason:      reason,
		SuspendedBy: suspendedBy,
	}
}

// Payload returns the event payload.
func (e TenantSuspended) Payload() map[string]any {
	return map[string]any{
		"tenant_id":    e.TenantID.String(),
		"reason":       e.Reason,
		"suspended_by": e.SuspendedBy,
	}
}

// ============================================================================
// Tenant Reactivated Event
// ============================================================================

// TenantReactivated is emitted when a suspended tenant is reactivated.
type TenantReactivated struct {
	EventMetadata
	ReactivatedBy string `json:"reactivated_by"`
}

// NewTenantReactivated creates a new TenantReactivated event.
func NewTenantReactivated(tenantID types.ID, reactivatedBy string) TenantReactivated {
	return TenantReactivated{
		EventMetadata: EventMetadata{
			EventType_: EventTypeTenantReactivated,
			Time_:      time.Now(),
			TenantID:   tenantID,
			Metadata:   make(map[string]any),
		},
		ReactivatedBy: reactivatedBy,
	}
}

// Payload returns the event payload.
func (e TenantReactivated) Payload() map[string]any {
	return map[string]any{
		"tenant_id":      e.TenantID.String(),
		"reactivated_by": e.ReactivatedBy,
	}
}

// ============================================================================
// Tenant Plan Changed Event
// ============================================================================

// TenantPlanChanged is emitted when a tenant's plan changes.
type TenantPlanChanged struct {
	EventMetadata
	OldPlan   string `json:"old_plan"`
	NewPlan   string `json:"new_plan"`
	ChangedBy string `json:"changed_by"`
	Reason    string `json:"reason,omitempty"`
}

// NewTenantPlanChanged creates a new TenantPlanChanged event.
func NewTenantPlanChanged(tenantID types.ID, oldPlan Plan, newPlan Plan, changedBy string, reason string) TenantPlanChanged {
	return TenantPlanChanged{
		EventMetadata: EventMetadata{
			EventType_: EventTypeTenantPlanChanged,
			Time_:      time.Now(),
			TenantID:   tenantID,
			Metadata:   make(map[string]any),
		},
		OldPlan:   oldPlan.String(),
		NewPlan:   newPlan.String(),
		ChangedBy: changedBy,
		Reason:    reason,
	}
}

// Payload returns the event payload.
func (e TenantPlanChanged) Payload() map[string]any {
	return map[string]any{
		"tenant_id":  e.TenantID.String(),
		"old_plan":   e.OldPlan,
		"new_plan":   e.NewPlan,
		"changed_by": e.ChangedBy,
		"reason":     e.Reason,
	}
}

// ============================================================================
// Tenant Deleted Event
// ============================================================================

// TenantDeleted is emitted when a tenant is deleted.
type TenantDeleted struct {
	EventMetadata
	Reason    string `json:"reason,omitempty"`
	DeletedBy string `json:"deleted_by"`
}

// NewTenantDeleted creates a new TenantDeleted event.
func NewTenantDeleted(tenantID types.ID, reason string, deletedBy string) TenantDeleted {
	return TenantDeleted{
		EventMetadata: EventMetadata{
			EventType_: EventTypeTenantDeleted,
			Time_:      time.Now(),
			TenantID:   tenantID,
			Metadata:   make(map[string]any),
		},
		Reason:    reason,
		DeletedBy: deletedBy,
	}
}

// Payload returns the event payload.
func (e TenantDeleted) Payload() map[string]any {
	return map[string]any{
		"tenant_id":  e.TenantID.String(),
		"reason":     e.Reason,
		"deleted_by": e.DeletedBy,
	}
}
