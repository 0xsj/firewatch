package domain

import (
	"regexp"
	"strings"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Flag is the aggregate root for feature flags.
// It encapsulates all flag state and enforces business rules.
type Flag struct {
	id             types.ID
	tenantID       string
	key            string
	name           string
	description    string
	enabled        bool
	variants       []Variant
	defaultVariant string
	rules          []Rule
	overrides      map[string]Override
	createdAt      types.Timestamp
	updatedAt      types.Timestamp
	version        int
	events         []Event
}

// Flag key constraints.
const (
	FlagKeyMinLength = 1
	FlagKeyMaxLength = 128
)

// flagKeyPattern defines valid flag key format.
// Allows lowercase letters, numbers, underscores, and hyphens.
var flagKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

// Override represents a forced variant for a specific target.
type Override struct {
	TargetType string // "tenant" or "user"
	TargetID   string
	VariantKey string
}

// ============================================================================
// Getters
// ============================================================================

// ID returns the flag ID.
func (f *Flag) ID() types.ID {
	return f.id
}

// TenantID returns the tenant ID.
func (f *Flag) TenantID() string {
	return f.tenantID
}

// Key returns the flag key.
func (f *Flag) Key() string {
	return f.key
}

// Name returns the flag name.
func (f *Flag) Name() string {
	return f.name
}

// Description returns the flag description.
func (f *Flag) Description() string {
	return f.description
}

// Enabled returns whether the flag is enabled.
func (f *Flag) Enabled() bool {
	return f.enabled
}

// Variants returns the flag variants.
func (f *Flag) Variants() []Variant {
	return f.variants
}

// DefaultVariant returns the default variant key.
func (f *Flag) DefaultVariant() string {
	return f.defaultVariant
}

// Rules returns the targeting rules.
func (f *Flag) Rules() []Rule {
	return f.rules
}

// Overrides returns the overrides map.
func (f *Flag) Overrides() map[string]Override {
	return f.overrides
}

// CreatedAt returns the creation timestamp.
func (f *Flag) CreatedAt() types.Timestamp {
	return f.createdAt
}

// UpdatedAt returns the last update timestamp.
func (f *Flag) UpdatedAt() types.Timestamp {
	return f.updatedAt
}

// Version returns the aggregate version.
func (f *Flag) Version() int {
	return f.version
}

// Events returns uncommitted domain events.
func (f *Flag) Events() []Event {
	return f.events
}

// ClearEvents clears uncommitted domain events.
func (f *Flag) ClearEvents() {
	f.events = nil
}

// ============================================================================
// Factory Methods
// ============================================================================

// Create creates a new feature flag.
// Emits FlagCreated event.
func Create(
	id types.ID,
	tenantID string,
	key string,
	name string,
	description string,
	enabled bool,
	createdBy string,
) (*Flag, error) {
	const op = "Flag.Create"

	// Validate key
	if err := validateFlagKey(op, key); err != nil {
		return nil, err
	}

	// Validate name
	name = strings.TrimSpace(name)
	if name == "" {
		name = key // Default name to key
	}

	now := types.Now()

	flag := &Flag{
		id:             id,
		tenantID:       tenantID,
		key:            key,
		name:           name,
		description:    description,
		enabled:        enabled,
		variants:       DefaultBooleanVariants(),
		defaultVariant: "disabled",
		rules:          make([]Rule, 0),
		overrides:      make(map[string]Override),
		createdAt:      now,
		updatedAt:      now,
		version:        1,
		events:         make([]Event, 0),
	}

	flag.addEvent(NewFlagCreated(id, tenantID, key, name, description, enabled, createdBy))

	return flag, nil
}

// Reconstitute recreates a flag from stored state (used by repository).
// Does NOT emit events - only for loading from database.
func Reconstitute(
	id types.ID,
	tenantID string,
	key string,
	name string,
	description string,
	enabled bool,
	variants []Variant,
	defaultVariant string,
	rules []Rule,
	overrides map[string]Override,
	createdAt types.Timestamp,
	updatedAt types.Timestamp,
	version int,
) *Flag {
	return &Flag{
		id:             id,
		tenantID:       tenantID,
		key:            key,
		name:           name,
		description:    description,
		enabled:        enabled,
		variants:       variants,
		defaultVariant: defaultVariant,
		rules:          rules,
		overrides:      overrides,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
		version:        version,
		events:         make([]Event, 0),
	}
}

// ============================================================================
// Commands
// ============================================================================

// Update updates the flag's metadata.
// Emits FlagUpdated event.
func (f *Flag) Update(name, description string, updatedBy string) error {
	var updatedFields []string

	name = strings.TrimSpace(name)
	if name != "" && name != f.name {
		f.name = name
		updatedFields = append(updatedFields, "name")
	}

	description = strings.TrimSpace(description)
	if description != f.description {
		f.description = description
		updatedFields = append(updatedFields, "description")
	}

	if len(updatedFields) > 0 {
		f.updatedAt = types.Now()
		f.version++
		f.addEvent(NewFlagUpdated(f.id, f.tenantID, f.key, updatedFields, updatedBy, f.version))
	}

	return nil
}

// Enable enables the flag.
// Emits FlagEnabled event.
func (f *Flag) Enable(enabledBy string) error {
	if f.enabled {
		return nil // Already enabled
	}

	f.enabled = true
	f.updatedAt = types.Now()
	f.version++
	f.addEvent(NewFlagEnabled(f.id, f.tenantID, f.key, enabledBy, f.version))

	return nil
}

// Disable disables the flag.
// Emits FlagDisabled event.
func (f *Flag) Disable(disabledBy string) error {
	if !f.enabled {
		return nil // Already disabled
	}

	f.enabled = false
	f.updatedAt = types.Now()
	f.version++
	f.addEvent(NewFlagDisabled(f.id, f.tenantID, f.key, disabledBy, f.version))

	return nil
}

// AddVariant adds a variant to the flag.
// Emits FlagVariantAdded event.
func (f *Flag) AddVariant(variant Variant, addedBy string) error {
	const op = "Flag.AddVariant"

	// Check for duplicate key
	for _, v := range f.variants {
		if v.Key() == variant.Key() {
			return ErrVariantKeyInvalid(op, variant.Key(), "variant key already exists")
		}
	}

	f.variants = append(f.variants, variant)
	f.updatedAt = types.Now()
	f.version++
	f.addEvent(NewFlagVariantAdded(f.id, f.tenantID, f.key, variant, addedBy, f.version))

	return nil
}

// RemoveVariant removes a variant from the flag.
// Emits FlagVariantRemoved event.
func (f *Flag) RemoveVariant(variantKey string, removedBy string) error {
	const op = "Flag.RemoveVariant"

	// Cannot remove default variant
	if variantKey == f.defaultVariant {
		return ErrVariantKeyInvalid(op, variantKey, "cannot remove default variant")
	}

	// Find and remove
	found := false
	for i, v := range f.variants {
		if v.Key() == variantKey {
			f.variants = append(f.variants[:i], f.variants[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return ErrVariantNotFound(op, f.key, variantKey)
	}

	f.updatedAt = types.Now()
	f.version++
	f.addEvent(NewFlagVariantRemoved(f.id, f.tenantID, f.key, variantKey, removedBy, f.version))

	return nil
}

// SetDefaultVariant sets the default variant.
func (f *Flag) SetDefaultVariant(variantKey string) error {
	const op = "Flag.SetDefaultVariant"

	// Verify variant exists
	found := false
	for _, v := range f.variants {
		if v.Key() == variantKey {
			found = true
			break
		}
	}

	if !found {
		return ErrVariantNotFound(op, f.key, variantKey)
	}

	f.defaultVariant = variantKey
	f.updatedAt = types.Now()
	f.version++

	return nil
}

// AddRule adds a targeting rule to the flag.
// Emits FlagRuleAdded event.
func (f *Flag) AddRule(rule Rule, addedBy string) error {
	const op = "Flag.AddRule"

	// Verify variant exists
	found := false
	for _, v := range f.variants {
		if v.Key() == rule.VariantKey() {
			found = true
			break
		}
	}

	if !found {
		return ErrVariantNotFound(op, f.key, rule.VariantKey())
	}

	f.rules = append(f.rules, rule)
	f.updatedAt = types.Now()
	f.version++
	f.addEvent(NewFlagRuleAdded(f.id, f.tenantID, f.key, rule, addedBy, f.version))

	return nil
}

// RemoveRule removes a targeting rule from the flag.
// Emits FlagRuleRemoved event.
func (f *Flag) RemoveRule(ruleID types.ID, removedBy string) error {
	const op = "Flag.RemoveRule"

	// Find and remove
	found := false
	for i, r := range f.rules {
		if r.ID() == ruleID {
			f.rules = append(f.rules[:i], f.rules[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return ErrRuleNotFound(op, f.key, ruleID.String())
	}

	f.updatedAt = types.Now()
	f.version++
	f.addEvent(NewFlagRuleRemoved(f.id, f.tenantID, f.key, ruleID.String(), removedBy, f.version))

	return nil
}

// SetOverride sets an override for a specific target.
// Emits FlagOverrideSet event.
func (f *Flag) SetOverride(targetType, targetID, variantKey, setBy string) error {
	const op = "Flag.SetOverride"

	// Validate target type
	if targetType != "tenant" && targetType != "user" {
		return ErrRuleInvalid(op, "target type must be 'tenant' or 'user'")
	}

	// Verify variant exists
	found := false
	for _, v := range f.variants {
		if v.Key() == variantKey {
			found = true
			break
		}
	}

	if !found {
		return ErrVariantNotFound(op, f.key, variantKey)
	}

	overrideKey := targetType + ":" + targetID
	f.overrides[overrideKey] = Override{
		TargetType: targetType,
		TargetID:   targetID,
		VariantKey: variantKey,
	}

	f.updatedAt = types.Now()
	f.version++
	f.addEvent(NewFlagOverrideSet(f.id, f.tenantID, f.key, targetType, targetID, variantKey, setBy, f.version))

	return nil
}

// RemoveOverride removes an override for a specific target.
// Emits FlagOverrideRemoved event.
func (f *Flag) RemoveOverride(targetType, targetID, removedBy string) error {
	const op = "Flag.RemoveOverride"

	overrideKey := targetType + ":" + targetID
	if _, exists := f.overrides[overrideKey]; !exists {
		return ErrOverrideNotFound(op, f.key, targetType, targetID)
	}

	delete(f.overrides, overrideKey)

	f.updatedAt = types.Now()
	f.version++
	f.addEvent(NewFlagOverrideRemoved(f.id, f.tenantID, f.key, targetType, targetID, removedBy, f.version))

	return nil
}

// MarkDeleted prepares the flag for deletion.
// Emits FlagDeleted event.
func (f *Flag) MarkDeleted(deletedBy string) {
	f.version++
	f.addEvent(NewFlagDeleted(f.id, f.tenantID, f.key, deletedBy, f.version))
}

// ============================================================================
// Query Methods
// ============================================================================

// GetVariant returns a variant by key.
func (f *Flag) GetVariant(key string) (Variant, bool) {
	for _, v := range f.variants {
		if v.Key() == key {
			return v, true
		}
	}
	return Variant{}, false
}

// GetDefaultVariantValue returns the default variant.
func (f *Flag) GetDefaultVariantValue() Variant {
	for _, v := range f.variants {
		if v.Key() == f.defaultVariant {
			return v
		}
	}
	return VariantDisabled
}

// GetOverride returns an override for a specific target.
func (f *Flag) GetOverride(targetType, targetID string) (Override, bool) {
	overrideKey := targetType + ":" + targetID
	override, exists := f.overrides[overrideKey]
	return override, exists
}

// ============================================================================
// Internal
// ============================================================================

// addEvent adds a domain event.
func (f *Flag) addEvent(event Event) {
	f.events = append(f.events, event)
}

// validateFlagKey validates a flag key.
func validateFlagKey(op, key string) error {
	key = strings.TrimSpace(key)

	if len(key) < FlagKeyMinLength {
		return ErrFlagKeyInvalid(op, key, "flag key is required")
	}

	if len(key) > FlagKeyMaxLength {
		return ErrFlagKeyInvalid(op, key, "flag key must be at most 128 characters")
	}

	if !flagKeyPattern.MatchString(key) {
		return ErrFlagKeyInvalid(op, key, "flag key must start with a letter and contain only lowercase letters, numbers, underscores, and hyphens")
	}

	return nil
}
