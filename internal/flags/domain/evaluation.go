package domain

import (
	"hash/fnv"
)

// EvaluationContext contains the context for evaluating a feature flag.
// This is passed to the evaluator to determine which variant to serve.
type EvaluationContext struct {
	tenantID   string
	userID     string
	attributes map[string]string
}

// NewEvaluationContext creates a new EvaluationContext.
func NewEvaluationContext(tenantID, userID string) *EvaluationContext {
	return &EvaluationContext{
		tenantID:   tenantID,
		userID:     userID,
		attributes: make(map[string]string),
	}
}

// TenantID returns the tenant ID.
func (ec *EvaluationContext) TenantID() string {
	return ec.tenantID
}

// UserID returns the user ID.
func (ec *EvaluationContext) UserID() string {
	return ec.userID
}

// Attributes returns all custom attributes.
func (ec *EvaluationContext) Attributes() map[string]string {
	return ec.attributes
}

// GetAttribute returns a custom attribute value.
func (ec *EvaluationContext) GetAttribute(key string) (string, bool) {
	val, ok := ec.attributes[key]
	return val, ok
}

// WithAttribute adds a custom attribute to the context.
func (ec *EvaluationContext) WithAttribute(key, value string) *EvaluationContext {
	ec.attributes[key] = value
	return ec
}

// WithAttributes adds multiple custom attributes to the context.
func (ec *EvaluationContext) WithAttributes(attrs map[string]string) *EvaluationContext {
	for k, v := range attrs {
		ec.attributes[k] = v
	}
	return ec
}

// HashKey returns a consistent hash key for percentage-based targeting.
// Uses the flag key + user ID (or tenant ID if no user) for deterministic bucketing.
func (ec *EvaluationContext) HashKey(flagKey string) string {
	if ec.userID != "" {
		return flagKey + ":" + ec.userID
	}
	if ec.tenantID != "" {
		return flagKey + ":" + ec.tenantID
	}
	return flagKey + ":anonymous"
}

// HashBucket returns a bucket number (0-99) for percentage-based targeting.
// The same context + flag key will always return the same bucket.
func (ec *EvaluationContext) HashBucket(flagKey string) int {
	h := fnv.New32a()
	h.Write([]byte(ec.HashKey(flagKey)))
	return int(h.Sum32() % 100)
}

// IsEmpty returns true if the context has no identifying information.
func (ec *EvaluationContext) IsEmpty() bool {
	return ec.tenantID == "" && ec.userID == "" && len(ec.attributes) == 0
}

// ============================================================================
// Evaluation Result
// ============================================================================

// EvaluationResult contains the result of evaluating a feature flag.
type EvaluationResult struct {
	flagKey     string
	enabled     bool
	variant     Variant
	reason      EvaluationReason
	matchedRule *Rule
}

// EvaluationReason explains why a particular variant was returned.
type EvaluationReason string

const (
	// ReasonDefault indicates the default variant was returned.
	ReasonDefault EvaluationReason = "default"

	// ReasonDisabled indicates the flag is disabled.
	ReasonDisabled EvaluationReason = "disabled"

	// ReasonOverride indicates an override was applied.
	ReasonOverride EvaluationReason = "override"

	// ReasonRule indicates a targeting rule matched.
	ReasonRule EvaluationReason = "rule"

	// ReasonPercentage indicates percentage-based targeting.
	ReasonPercentage EvaluationReason = "percentage"

	// ReasonError indicates an error occurred during evaluation.
	ReasonError EvaluationReason = "error"
)

// NewEvaluationResult creates a new EvaluationResult.
func NewEvaluationResult(flagKey string, enabled bool, variant Variant, reason EvaluationReason) *EvaluationResult {
	return &EvaluationResult{
		flagKey: flagKey,
		enabled: enabled,
		variant: variant,
		reason:  reason,
	}
}

// FlagKey returns the flag key that was evaluated.
func (er *EvaluationResult) FlagKey() string {
	return er.flagKey
}

// Enabled returns true if the flag is enabled.
func (er *EvaluationResult) Enabled() bool {
	return er.enabled
}

// Variant returns the selected variant.
func (er *EvaluationResult) Variant() Variant {
	return er.variant
}

// VariantKey returns the selected variant key.
func (er *EvaluationResult) VariantKey() string {
	return er.variant.Key()
}

// VariantValue returns the selected variant value.
func (er *EvaluationResult) VariantValue() string {
	return er.variant.Value()
}

// Reason returns why this variant was selected.
func (er *EvaluationResult) Reason() EvaluationReason {
	return er.reason
}

// MatchedRule returns the rule that matched (if any).
func (er *EvaluationResult) MatchedRule() *Rule {
	return er.matchedRule
}

// WithMatchedRule sets the matched rule.
func (er *EvaluationResult) WithMatchedRule(rule *Rule) *EvaluationResult {
	er.matchedRule = rule
	return er
}

// IsOn returns true if the flag evaluated to an enabled state.
// Convenience method for boolean flags.
func (er *EvaluationResult) IsOn() bool {
	return er.enabled && er.variant.Value() == "true"
}

// IsOff returns true if the flag evaluated to a disabled state.
// Convenience method for boolean flags.
func (er *EvaluationResult) IsOff() bool {
	return !er.enabled || er.variant.Value() == "false"
}

// ============================================================================
// Disabled Result Factory
// ============================================================================

// NewDisabledResult creates a result for a disabled flag.
func NewDisabledResult(flagKey string) *EvaluationResult {
	return &EvaluationResult{
		flagKey: flagKey,
		enabled: false,
		variant: VariantDisabled,
		reason:  ReasonDisabled,
	}
}

// NewErrorResult creates a result for an evaluation error.
func NewErrorResult(flagKey string) *EvaluationResult {
	return &EvaluationResult{
		flagKey: flagKey,
		enabled: false,
		variant: VariantDisabled,
		reason:  ReasonError,
	}
}
