package domain

import (
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// RuleType defines the type of targeting rule.
type RuleType string

const (
	// RuleTypeTenant targets specific tenants.
	RuleTypeTenant RuleType = "tenant"

	// RuleTypeUser targets specific users.
	RuleTypeUser RuleType = "user"

	// RuleTypePercent targets a percentage of traffic.
	RuleTypePercent RuleType = "percent"

	// RuleTypeAttribute targets based on custom attributes.
	RuleTypeAttribute RuleType = "attribute"
)

// IsValid checks if the rule type is valid.
func (rt RuleType) IsValid() bool {
	switch rt {
	case RuleTypeTenant, RuleTypeUser, RuleTypePercent, RuleTypeAttribute:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (rt RuleType) String() string {
	return string(rt)
}

// Operator defines comparison operators for rules.
type Operator string

const (
	OperatorEquals      Operator = "eq"
	OperatorNotEquals   Operator = "neq"
	OperatorContains    Operator = "contains"
	OperatorNotContains Operator = "not_contains"
	OperatorIn          Operator = "in"
	OperatorNotIn       Operator = "not_in"
	OperatorGreaterThan Operator = "gt"
	OperatorLessThan    Operator = "lt"
	OperatorExists      Operator = "exists"
	OperatorNotExists   Operator = "not_exists"
)

// IsValid checks if the operator is valid.
func (o Operator) IsValid() bool {
	switch o {
	case OperatorEquals, OperatorNotEquals, OperatorContains, OperatorNotContains,
		OperatorIn, OperatorNotIn, OperatorGreaterThan, OperatorLessThan,
		OperatorExists, OperatorNotExists:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (o Operator) String() string {
	return string(o)
}

// Rule represents a targeting rule for a feature flag.
// Rules determine which variant a user receives based on conditions.
type Rule struct {
	id         types.ID
	ruleType   RuleType
	attribute  string
	operator   Operator
	values     []string
	percentage int
	variantKey string
	priority   int
}

// NewRule creates a new Rule with validation.
func NewRule(
	id types.ID,
	ruleType RuleType,
	attribute string,
	operator Operator,
	values []string,
	percentage int,
	variantKey string,
	priority int,
) (Rule, error) {
	const op = "Rule.New"

	if !ruleType.IsValid() {
		return Rule{}, ErrRuleInvalid(op, "invalid rule type")
	}

	// Validate based on rule type
	switch ruleType {
	case RuleTypePercent:
		if percentage < 0 || percentage > 100 {
			return Rule{}, ErrRuleInvalid(op, "percentage must be between 0 and 100")
		}
	case RuleTypeAttribute:
		if attribute == "" {
			return Rule{}, ErrRuleInvalid(op, "attribute is required for attribute rules")
		}
		if !operator.IsValid() {
			return Rule{}, ErrRuleInvalid(op, "invalid operator")
		}
	case RuleTypeTenant, RuleTypeUser:
		if len(values) == 0 {
			return Rule{}, ErrRuleInvalid(op, "at least one value is required")
		}
	}

	if variantKey == "" {
		return Rule{}, ErrRuleInvalid(op, "variant key is required")
	}

	return Rule{
		id:         id,
		ruleType:   ruleType,
		attribute:  attribute,
		operator:   operator,
		values:     values,
		percentage: percentage,
		variantKey: variantKey,
		priority:   priority,
	}, nil
}

// ID returns the rule ID.
func (r Rule) ID() types.ID {
	return r.id
}

// Type returns the rule type.
func (r Rule) Type() RuleType {
	return r.ruleType
}

// Attribute returns the attribute name (for attribute rules).
func (r Rule) Attribute() string {
	return r.attribute
}

// Operator returns the comparison operator.
func (r Rule) Operator() Operator {
	return r.operator
}

// Values returns the target values.
func (r Rule) Values() []string {
	return r.values
}

// Percentage returns the percentage (for percent rules).
func (r Rule) Percentage() int {
	return r.percentage
}

// VariantKey returns the variant to serve when rule matches.
func (r Rule) VariantKey() string {
	return r.variantKey
}

// Priority returns the rule priority (lower = higher priority).
func (r Rule) Priority() int {
	return r.priority
}

// IsEmpty returns true if the rule is empty.
func (r Rule) IsEmpty() bool {
	return r.id.IsEmpty()
}

// ============================================================================
// Rule Factories
// ============================================================================

// NewTenantRule creates a rule targeting specific tenants.
func NewTenantRule(id types.ID, tenantIDs []string, variantKey string, priority int) (Rule, error) {
	return NewRule(id, RuleTypeTenant, "", OperatorIn, tenantIDs, 0, variantKey, priority)
}

// NewUserRule creates a rule targeting specific users.
func NewUserRule(id types.ID, userIDs []string, variantKey string, priority int) (Rule, error) {
	return NewRule(id, RuleTypeUser, "", OperatorIn, userIDs, 0, variantKey, priority)
}

// NewPercentRule creates a rule targeting a percentage of traffic.
func NewPercentRule(id types.ID, percentage int, variantKey string, priority int) (Rule, error) {
	return NewRule(id, RuleTypePercent, "", "", nil, percentage, variantKey, priority)
}

// NewAttributeRule creates a rule targeting based on custom attributes.
func NewAttributeRule(id types.ID, attribute string, operator Operator, values []string, variantKey string, priority int) (Rule, error) {
	return NewRule(id, RuleTypeAttribute, attribute, operator, values, 0, variantKey, priority)
}
