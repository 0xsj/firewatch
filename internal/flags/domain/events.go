package domain

import (
	"time"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Event is the base interface for all flag domain events.
// All flag events embed EventMetadata for common fields.
type Event interface {
	// Type returns the event type identifier
	Type() string

	// EventType returns the event type identifier (alias)
	EventType() string

	// EventTime returns when the event occurred
	EventTime() time.Time

	// AggregateID returns the flag ID
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
	FlagID     types.ID       `json:"flag_id"`
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

// AggregateID returns the flag ID.
func (m EventMetadata) AggregateID() types.ID {
	return m.FlagID
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
	EventTypeFlagCreated         = "flag.created"
	EventTypeFlagUpdated         = "flag.updated"
	EventTypeFlagDeleted         = "flag.deleted"
	EventTypeFlagEnabled         = "flag.enabled"
	EventTypeFlagDisabled        = "flag.disabled"
	EventTypeFlagVariantAdded    = "flag.variant_added"
	EventTypeFlagVariantRemoved  = "flag.variant_removed"
	EventTypeFlagRuleAdded       = "flag.rule_added"
	EventTypeFlagRuleRemoved     = "flag.rule_removed"
	EventTypeFlagOverrideSet     = "flag.override_set"
	EventTypeFlagOverrideRemoved = "flag.override_removed"
)

// ============================================================================
// Flag Lifecycle Events
// ============================================================================

// FlagCreated is emitted when a new feature flag is created.
type FlagCreated struct {
	EventMetadata
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	CreatedBy   string `json:"created_by"`
}

// NewFlagCreated creates a new FlagCreated event.
func NewFlagCreated(flagID types.ID, tenantID, key, name, description string, enabled bool, createdBy string) FlagCreated {
	return FlagCreated{
		EventMetadata: EventMetadata{
			EventType_: EventTypeFlagCreated,
			Time_:      time.Now(),
			FlagID:     flagID,
			TenantID:   tenantID,
			Version_:   1,
			Metadata:   make(map[string]any),
		},
		Key:         key,
		Name:        name,
		Description: description,
		Enabled:     enabled,
		CreatedBy:   createdBy,
	}
}

// Payload returns the event payload.
func (e FlagCreated) Payload() map[string]any {
	return map[string]any{
		"flag_id":     e.FlagID.String(),
		"tenant_id":   e.TenantID,
		"key":         e.Key,
		"name":        e.Name,
		"description": e.Description,
		"enabled":     e.Enabled,
		"created_by":  e.CreatedBy,
	}
}

// FlagUpdated is emitted when a feature flag's metadata is updated.
type FlagUpdated struct {
	EventMetadata
	Key           string   `json:"key"`
	UpdatedFields []string `json:"updated_fields"`
	UpdatedBy     string   `json:"updated_by"`
}

// NewFlagUpdated creates a new FlagUpdated event.
func NewFlagUpdated(flagID types.ID, tenantID, key string, updatedFields []string, updatedBy string, version int) FlagUpdated {
	return FlagUpdated{
		EventMetadata: EventMetadata{
			EventType_: EventTypeFlagUpdated,
			Time_:      time.Now(),
			FlagID:     flagID,
			TenantID:   tenantID,
			Version_:   version,
			Metadata:   make(map[string]any),
		},
		Key:           key,
		UpdatedFields: updatedFields,
		UpdatedBy:     updatedBy,
	}
}

// Payload returns the event payload.
func (e FlagUpdated) Payload() map[string]any {
	return map[string]any{
		"flag_id":        e.FlagID.String(),
		"tenant_id":      e.TenantID,
		"key":            e.Key,
		"updated_fields": e.UpdatedFields,
		"updated_by":     e.UpdatedBy,
	}
}

// FlagDeleted is emitted when a feature flag is deleted.
type FlagDeleted struct {
	EventMetadata
	Key       string `json:"key"`
	DeletedBy string `json:"deleted_by"`
}

// NewFlagDeleted creates a new FlagDeleted event.
func NewFlagDeleted(flagID types.ID, tenantID, key, deletedBy string, version int) FlagDeleted {
	return FlagDeleted{
		EventMetadata: EventMetadata{
			EventType_: EventTypeFlagDeleted,
			Time_:      time.Now(),
			FlagID:     flagID,
			TenantID:   tenantID,
			Version_:   version,
			Metadata:   make(map[string]any),
		},
		Key:       key,
		DeletedBy: deletedBy,
	}
}

// Payload returns the event payload.
func (e FlagDeleted) Payload() map[string]any {
	return map[string]any{
		"flag_id":    e.FlagID.String(),
		"tenant_id":  e.TenantID,
		"key":        e.Key,
		"deleted_by": e.DeletedBy,
	}
}

// ============================================================================
// Flag Status Events
// ============================================================================

// FlagEnabled is emitted when a feature flag is enabled.
type FlagEnabled struct {
	EventMetadata
	Key       string `json:"key"`
	EnabledBy string `json:"enabled_by"`
}

// NewFlagEnabled creates a new FlagEnabled event.
func NewFlagEnabled(flagID types.ID, tenantID, key, enabledBy string, version int) FlagEnabled {
	return FlagEnabled{
		EventMetadata: EventMetadata{
			EventType_: EventTypeFlagEnabled,
			Time_:      time.Now(),
			FlagID:     flagID,
			TenantID:   tenantID,
			Version_:   version,
			Metadata:   make(map[string]any),
		},
		Key:       key,
		EnabledBy: enabledBy,
	}
}

// Payload returns the event payload.
func (e FlagEnabled) Payload() map[string]any {
	return map[string]any{
		"flag_id":    e.FlagID.String(),
		"tenant_id":  e.TenantID,
		"key":        e.Key,
		"enabled_by": e.EnabledBy,
	}
}

// FlagDisabled is emitted when a feature flag is disabled.
type FlagDisabled struct {
	EventMetadata
	Key        string `json:"key"`
	DisabledBy string `json:"disabled_by"`
}

// NewFlagDisabled creates a new FlagDisabled event.
func NewFlagDisabled(flagID types.ID, tenantID, key, disabledBy string, version int) FlagDisabled {
	return FlagDisabled{
		EventMetadata: EventMetadata{
			EventType_: EventTypeFlagDisabled,
			Time_:      time.Now(),
			FlagID:     flagID,
			TenantID:   tenantID,
			Version_:   version,
			Metadata:   make(map[string]any),
		},
		Key:        key,
		DisabledBy: disabledBy,
	}
}

// Payload returns the event payload.
func (e FlagDisabled) Payload() map[string]any {
	return map[string]any{
		"flag_id":     e.FlagID.String(),
		"tenant_id":   e.TenantID,
		"key":         e.Key,
		"disabled_by": e.DisabledBy,
	}
}

// ============================================================================
// Variant Events
// ============================================================================

// FlagVariantAdded is emitted when a variant is added to a flag.
type FlagVariantAdded struct {
	EventMetadata
	Key        string `json:"key"`
	VariantKey string `json:"variant_key"`
	Value      string `json:"value"`
	Weight     int    `json:"weight"`
	AddedBy    string `json:"added_by"`
}

// NewFlagVariantAdded creates a new FlagVariantAdded event.
func NewFlagVariantAdded(flagID types.ID, tenantID, key string, variant Variant, addedBy string, version int) FlagVariantAdded {
	return FlagVariantAdded{
		EventMetadata: EventMetadata{
			EventType_: EventTypeFlagVariantAdded,
			Time_:      time.Now(),
			FlagID:     flagID,
			TenantID:   tenantID,
			Version_:   version,
			Metadata:   make(map[string]any),
		},
		Key:        key,
		VariantKey: variant.Key(),
		Value:      variant.Value(),
		Weight:     variant.Weight(),
		AddedBy:    addedBy,
	}
}

// Payload returns the event payload.
func (e FlagVariantAdded) Payload() map[string]any {
	return map[string]any{
		"flag_id":     e.FlagID.String(),
		"tenant_id":   e.TenantID,
		"key":         e.Key,
		"variant_key": e.VariantKey,
		"value":       e.Value,
		"weight":      e.Weight,
		"added_by":    e.AddedBy,
	}
}

// FlagVariantRemoved is emitted when a variant is removed from a flag.
type FlagVariantRemoved struct {
	EventMetadata
	Key        string `json:"key"`
	VariantKey string `json:"variant_key"`
	RemovedBy  string `json:"removed_by"`
}

// NewFlagVariantRemoved creates a new FlagVariantRemoved event.
func NewFlagVariantRemoved(flagID types.ID, tenantID, key, variantKey, removedBy string, version int) FlagVariantRemoved {
	return FlagVariantRemoved{
		EventMetadata: EventMetadata{
			EventType_: EventTypeFlagVariantRemoved,
			Time_:      time.Now(),
			FlagID:     flagID,
			TenantID:   tenantID,
			Version_:   version,
			Metadata:   make(map[string]any),
		},
		Key:        key,
		VariantKey: variantKey,
		RemovedBy:  removedBy,
	}
}

// Payload returns the event payload.
func (e FlagVariantRemoved) Payload() map[string]any {
	return map[string]any{
		"flag_id":     e.FlagID.String(),
		"tenant_id":   e.TenantID,
		"key":         e.Key,
		"variant_key": e.VariantKey,
		"removed_by":  e.RemovedBy,
	}
}

// ============================================================================
// Rule Events
// ============================================================================

// FlagRuleAdded is emitted when a targeting rule is added to a flag.
type FlagRuleAdded struct {
	EventMetadata
	Key        string `json:"key"`
	RuleID     string `json:"rule_id"`
	RuleType   string `json:"rule_type"`
	VariantKey string `json:"variant_key"`
	Priority   int    `json:"priority"`
	AddedBy    string `json:"added_by"`
}

// NewFlagRuleAdded creates a new FlagRuleAdded event.
func NewFlagRuleAdded(flagID types.ID, tenantID, key string, rule Rule, addedBy string, version int) FlagRuleAdded {
	return FlagRuleAdded{
		EventMetadata: EventMetadata{
			EventType_: EventTypeFlagRuleAdded,
			Time_:      time.Now(),
			FlagID:     flagID,
			TenantID:   tenantID,
			Version_:   version,
			Metadata:   make(map[string]any),
		},
		Key:        key,
		RuleID:     rule.ID().String(),
		RuleType:   rule.Type().String(),
		VariantKey: rule.VariantKey(),
		Priority:   rule.Priority(),
		AddedBy:    addedBy,
	}
}

// Payload returns the event payload.
func (e FlagRuleAdded) Payload() map[string]any {
	return map[string]any{
		"flag_id":     e.FlagID.String(),
		"tenant_id":   e.TenantID,
		"key":         e.Key,
		"rule_id":     e.RuleID,
		"rule_type":   e.RuleType,
		"variant_key": e.VariantKey,
		"priority":    e.Priority,
		"added_by":    e.AddedBy,
	}
}

// FlagRuleRemoved is emitted when a targeting rule is removed from a flag.
type FlagRuleRemoved struct {
	EventMetadata
	Key       string `json:"key"`
	RuleID    string `json:"rule_id"`
	RemovedBy string `json:"removed_by"`
}

// NewFlagRuleRemoved creates a new FlagRuleRemoved event.
func NewFlagRuleRemoved(flagID types.ID, tenantID, key, ruleID, removedBy string, version int) FlagRuleRemoved {
	return FlagRuleRemoved{
		EventMetadata: EventMetadata{
			EventType_: EventTypeFlagRuleRemoved,
			Time_:      time.Now(),
			FlagID:     flagID,
			TenantID:   tenantID,
			Version_:   version,
			Metadata:   make(map[string]any),
		},
		Key:       key,
		RuleID:    ruleID,
		RemovedBy: removedBy,
	}
}

// Payload returns the event payload.
func (e FlagRuleRemoved) Payload() map[string]any {
	return map[string]any{
		"flag_id":    e.FlagID.String(),
		"tenant_id":  e.TenantID,
		"key":        e.Key,
		"rule_id":    e.RuleID,
		"removed_by": e.RemovedBy,
	}
}

// ============================================================================
// Override Events
// ============================================================================

// FlagOverrideSet is emitted when an override is set for a specific target.
type FlagOverrideSet struct {
	EventMetadata
	Key        string `json:"key"`
	TargetType string `json:"target_type"` // "tenant" or "user"
	TargetID   string `json:"target_id"`
	VariantKey string `json:"variant_key"`
	SetBy      string `json:"set_by"`
}

// NewFlagOverrideSet creates a new FlagOverrideSet event.
func NewFlagOverrideSet(flagID types.ID, tenantID, key, targetType, targetID, variantKey, setBy string, version int) FlagOverrideSet {
	return FlagOverrideSet{
		EventMetadata: EventMetadata{
			EventType_: EventTypeFlagOverrideSet,
			Time_:      time.Now(),
			FlagID:     flagID,
			TenantID:   tenantID,
			Version_:   version,
			Metadata:   make(map[string]any),
		},
		Key:        key,
		TargetType: targetType,
		TargetID:   targetID,
		VariantKey: variantKey,
		SetBy:      setBy,
	}
}

// Payload returns the event payload.
func (e FlagOverrideSet) Payload() map[string]any {
	return map[string]any{
		"flag_id":     e.FlagID.String(),
		"tenant_id":   e.TenantID,
		"key":         e.Key,
		"target_type": e.TargetType,
		"target_id":   e.TargetID,
		"variant_key": e.VariantKey,
		"set_by":      e.SetBy,
	}
}

// FlagOverrideRemoved is emitted when an override is removed.
type FlagOverrideRemoved struct {
	EventMetadata
	Key        string `json:"key"`
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	RemovedBy  string `json:"removed_by"`
}

// NewFlagOverrideRemoved creates a new FlagOverrideRemoved event.
func NewFlagOverrideRemoved(flagID types.ID, tenantID, key, targetType, targetID, removedBy string, version int) FlagOverrideRemoved {
	return FlagOverrideRemoved{
		EventMetadata: EventMetadata{
			EventType_: EventTypeFlagOverrideRemoved,
			Time_:      time.Now(),
			FlagID:     flagID,
			TenantID:   tenantID,
			Version_:   version,
			Metadata:   make(map[string]any),
		},
		Key:        key,
		TargetType: targetType,
		TargetID:   targetID,
		RemovedBy:  removedBy,
	}
}

// Payload returns the event payload.
func (e FlagOverrideRemoved) Payload() map[string]any {
	return map[string]any{
		"flag_id":     e.FlagID.String(),
		"tenant_id":   e.TenantID,
		"key":         e.Key,
		"target_type": e.TargetType,
		"target_id":   e.TargetID,
		"removed_by":  e.RemovedBy,
	}
}
