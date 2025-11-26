package tenant

import "fmt"

// Plan represents the subscription tier of a tenant.
type Plan string

const (
	// PlanFree is the free tier with limited features.
	PlanFree Plan = "free"

	// PlanStarter is the entry-level paid tier.
	PlanStarter Plan = "starter"

	// PlanPro is the professional tier with advanced features.
	PlanPro Plan = "pro"

	// PlanEnterprise is the enterprise tier with full features.
	PlanEnterprise Plan = "enterprise"
)

// String returns the string representation of the plan.
func (p Plan) String() string {
	return string(p)
}

// Validate checks if the plan is a valid value.
func (p Plan) Validate() error {
	switch p {
	case PlanFree, PlanStarter, PlanPro, PlanEnterprise:
		return nil
	default:
		return fmt.Errorf("invalid tenant plan: %s", p)
	}
}

// Tier returns the numeric tier level for comparison.
// Higher values indicate higher-tier plans.
func (p Plan) Tier() int {
	switch p {
	case PlanFree:
		return 0
	case PlanStarter:
		return 1
	case PlanPro:
		return 2
	case PlanEnterprise:
		return 3
	default:
		return -1
	}
}

// IsHigherThan returns true if this plan is a higher tier than the other.
func (p Plan) IsHigherThan(other Plan) bool {
	return p.Tier() > other.Tier()
}

// IsLowerThan returns true if this plan is a lower tier than the other.
func (p Plan) IsLowerThan(other Plan) bool {
	return p.Tier() < other.Tier()
}

// IsPaid returns true if the plan is a paid tier.
func (p Plan) IsPaid() bool {
	return p != PlanFree
}

// AllPlans returns all valid tenant plans.
func AllPlans() []Plan {
	return []Plan{
		PlanFree,
		PlanStarter,
		PlanPro,
		PlanEnterprise,
	}
}

// ParsePlan parses a string into a Plan.
func ParsePlan(s string) (Plan, error) {
	plan := Plan(s)
	if err := plan.Validate(); err != nil {
		return "", err
	}
	return plan, nil
}
