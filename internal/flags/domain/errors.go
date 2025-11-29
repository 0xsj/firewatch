package domain

import "github.com/0xsj/hexagonal-go/pkg/errors"

// Domain-specific error codes for the Flag aggregate.
// These codes are stable across versions and used by clients.
//
// Naming convention: FLAG_RESOURCE_CONDITION
const (
	// Flag errors
	ErrCodeFlagNotFound      = errors.Code("FLAG_NOT_FOUND")
	ErrCodeFlagAlreadyExists = errors.Code("FLAG_ALREADY_EXISTS")
	ErrCodeFlagDisabled      = errors.Code("FLAG_DISABLED")
	ErrCodeFlagKeyInvalid    = errors.Code("FLAG_KEY_INVALID")

	// Variant errors
	ErrCodeVariantNotFound   = errors.Code("FLAG_VARIANT_NOT_FOUND")
	ErrCodeVariantKeyInvalid = errors.Code("FLAG_VARIANT_KEY_INVALID")

	// Rule errors
	ErrCodeRuleInvalid  = errors.Code("FLAG_RULE_INVALID")
	ErrCodeRuleNotFound = errors.Code("FLAG_RULE_NOT_FOUND")

	// Override errors
	ErrCodeOverrideNotFound = errors.Code("FLAG_OVERRIDE_NOT_FOUND")

	// Evaluation errors
	ErrCodeEvaluationFailed = errors.Code("FLAG_EVALUATION_FAILED")
)

// Register domain error codes with the global registry.
// This happens once when the package is imported.
func init() {
	// Flag errors
	errors.Register(
		ErrCodeFlagNotFound,
		errors.KindNotFound,
		"feature flag not found",
	)

	errors.Register(
		ErrCodeFlagAlreadyExists,
		errors.KindConflict,
		"feature flag already exists",
	)

	errors.Register(
		ErrCodeFlagDisabled,
		errors.KindDomain,
		"feature flag is disabled",
	)

	errors.Register(
		ErrCodeFlagKeyInvalid,
		errors.KindValidation,
		"invalid feature flag key",
	)

	// Variant errors
	errors.Register(
		ErrCodeVariantNotFound,
		errors.KindNotFound,
		"variant not found",
	)

	errors.Register(
		ErrCodeVariantKeyInvalid,
		errors.KindValidation,
		"invalid variant key",
	)

	// Rule errors
	errors.Register(
		ErrCodeRuleInvalid,
		errors.KindValidation,
		"invalid targeting rule",
	)

	errors.Register(
		ErrCodeRuleNotFound,
		errors.KindNotFound,
		"targeting rule not found",
	)

	// Override errors
	errors.Register(
		ErrCodeOverrideNotFound,
		errors.KindNotFound,
		"override not found",
	)

	// Evaluation errors
	errors.Register(
		ErrCodeEvaluationFailed,
		errors.KindDomain,
		"flag evaluation failed",
	)
}

// ============================================================================
// Error Constructor Helpers
// ============================================================================

// ErrFlagNotFound creates a flag not found error.
func ErrFlagNotFound(operation string, key string) error {
	return errors.NotFound(operation, "flag").
		WithCode(ErrCodeFlagNotFound).
		WithMeta("flag_key", key)
}

// ErrFlagAlreadyExists creates a flag already exists error.
func ErrFlagAlreadyExists(operation string, key string) error {
	return errors.Conflict(operation, "flag").
		WithCode(ErrCodeFlagAlreadyExists).
		WithMeta("flag_key", key)
}

// ErrFlagDisabled creates a flag disabled error.
func ErrFlagDisabled(operation string, key string) error {
	return errors.Domain(operation, "flag is disabled").
		WithCode(ErrCodeFlagDisabled).
		WithMeta("flag_key", key)
}

// ErrFlagKeyInvalid creates a flag key invalid error.
func ErrFlagKeyInvalid(operation string, key string, reason string) error {
	return errors.Validation(operation, reason).
		WithCode(ErrCodeFlagKeyInvalid).
		WithMeta("flag_key", key)
}

// ErrVariantNotFound creates a variant not found error.
func ErrVariantNotFound(operation string, flagKey string, variantKey string) error {
	return errors.NotFound(operation, "variant").
		WithCode(ErrCodeVariantNotFound).
		WithMeta("flag_key", flagKey).
		WithMeta("variant_key", variantKey)
}

// ErrVariantKeyInvalid creates a variant key invalid error.
func ErrVariantKeyInvalid(operation string, key string, reason string) error {
	return errors.Validation(operation, reason).
		WithCode(ErrCodeVariantKeyInvalid).
		WithMeta("variant_key", key)
}

// ErrRuleInvalid creates a rule invalid error.
func ErrRuleInvalid(operation string, reason string) error {
	return errors.Validation(operation, reason).
		WithCode(ErrCodeRuleInvalid)
}

// ErrRuleNotFound creates a rule not found error.
func ErrRuleNotFound(operation string, flagKey string, ruleID string) error {
	return errors.NotFound(operation, "rule").
		WithCode(ErrCodeRuleNotFound).
		WithMeta("flag_key", flagKey).
		WithMeta("rule_id", ruleID)
}

// ErrOverrideNotFound creates an override not found error.
func ErrOverrideNotFound(operation string, flagKey string, targetType string, targetID string) error {
	return errors.NotFound(operation, "override").
		WithCode(ErrCodeOverrideNotFound).
		WithMeta("flag_key", flagKey).
		WithMeta("target_type", targetType).
		WithMeta("target_id", targetID)
}

// ErrEvaluationFailed creates an evaluation failed error.
func ErrEvaluationFailed(operation string, flagKey string, reason string) error {
	return errors.Domain(operation, reason).
		WithCode(ErrCodeEvaluationFailed).
		WithMeta("flag_key", flagKey)
}
