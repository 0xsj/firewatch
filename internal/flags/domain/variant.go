package domain

import (
	"regexp"
	"strings"
)

// Variant represents a possible value for a feature flag.
// Used for A/B testing and multivariate flags.
//
// Example variants for a "checkout_flow" flag:
//   - { Key: "control", Value: "v1", Weight: 50 }
//   - { Key: "treatment", Value: "v2", Weight: 50 }
type Variant struct {
	key         string
	value       string
	weight      int
	description string
}

// Variant key constraints.
const (
	VariantKeyMinLength = 1
	VariantKeyMaxLength = 64
	VariantWeightMin    = 0
	VariantWeightMax    = 100
)

// variantKeyPattern defines valid variant key format.
// Allows lowercase letters, numbers, underscores, and hyphens.
var variantKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

// NewVariant creates a new Variant with validation.
func NewVariant(key, value string, weight int, description string) (Variant, error) {
	const op = "Variant.New"

	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)

	if err := validateVariantKey(op, key); err != nil {
		return Variant{}, err
	}

	if weight < VariantWeightMin || weight > VariantWeightMax {
		return Variant{}, ErrRuleInvalid(op, "weight must be between 0 and 100")
	}

	return Variant{
		key:         key,
		value:       value,
		weight:      weight,
		description: description,
	}, nil
}

// Key returns the variant key.
func (v Variant) Key() string {
	return v.key
}

// Value returns the variant value.
func (v Variant) Value() string {
	return v.value
}

// Weight returns the variant weight (0-100).
func (v Variant) Weight() int {
	return v.weight
}

// Description returns the variant description.
func (v Variant) Description() string {
	return v.description
}

// IsEmpty returns true if the variant is empty.
func (v Variant) IsEmpty() bool {
	return v.key == ""
}

// Equals checks if two variants are equal by key.
func (v Variant) Equals(other Variant) bool {
	return v.key == other.key
}

// validateVariantKey validates a variant key.
func validateVariantKey(op, key string) error {
	if len(key) < VariantKeyMinLength {
		return ErrVariantKeyInvalid(op, key, "variant key is required")
	}

	if len(key) > VariantKeyMaxLength {
		return ErrVariantKeyInvalid(op, key, "variant key must be at most 64 characters")
	}

	if !variantKeyPattern.MatchString(key) {
		return ErrVariantKeyInvalid(op, key, "variant key must start with a letter and contain only lowercase letters, numbers, underscores, and hyphens")
	}

	return nil
}

// ============================================================================
// Predefined Variants
// ============================================================================

// Boolean flag variants.
var (
	VariantEnabled  = Variant{key: "enabled", value: "true", weight: 100}
	VariantDisabled = Variant{key: "disabled", value: "false", weight: 100}
)

// DefaultBooleanVariants returns the default variants for a boolean flag.
func DefaultBooleanVariants() []Variant {
	return []Variant{
		{key: "enabled", value: "true", weight: 50},
		{key: "disabled", value: "false", weight: 50},
	}
}
