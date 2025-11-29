package query

import (
	"context"
	"fmt"
	"sort"

	"github.com/0xsj/hexagonal-go/internal/flags/application/dto"
	"github.com/0xsj/hexagonal-go/internal/flags/domain"
)

// EvaluateFlagQuery handles evaluating a feature flag for a given context.
type EvaluateFlagQuery struct {
	repo domain.Repository
}

// NewEvaluateFlagQuery creates a new EvaluateFlagQuery.
func NewEvaluateFlagQuery(repo domain.Repository) *EvaluateFlagQuery {
	return &EvaluateFlagQuery{
		repo: repo,
	}
}

// EvaluateFlagRequest is the input for evaluating a flag.
type EvaluateFlagRequest struct {
	TenantID   string
	FlagKey    string
	UserID     string
	Attributes map[string]string
}

// Handle executes the query to evaluate a flag.
func (q *EvaluateFlagQuery) Handle(ctx context.Context, req EvaluateFlagRequest) (*dto.EvaluateFlagResponse, error) {
	const op = "EvaluateFlagQuery.Handle"

	// Find flag by key
	flag, err := q.repo.FindByKey(ctx, req.TenantID, req.FlagKey)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Build evaluation context
	evalCtx := domain.NewEvaluationContext(req.TenantID, req.UserID)
	if req.Attributes != nil {
		evalCtx.WithAttributes(req.Attributes)
	}

	// Evaluate
	result := q.evaluate(flag, evalCtx)

	return &dto.EvaluateFlagResponse{
		Result: dto.MapEvaluationResultToDTO(result),
	}, nil
}

// evaluate performs the actual flag evaluation logic.
func (q *EvaluateFlagQuery) evaluate(flag *domain.Flag, evalCtx *domain.EvaluationContext) *domain.EvaluationResult {
	// If flag is disabled, return disabled result
	if !flag.Enabled() {
		return domain.NewDisabledResult(flag.Key())
	}

	// Check for user override
	if evalCtx.UserID() != "" {
		if override, exists := flag.GetOverride("user", evalCtx.UserID()); exists {
			variant, _ := flag.GetVariant(override.VariantKey)
			return domain.NewEvaluationResult(flag.Key(), true, variant, domain.ReasonOverride)
		}
	}

	// Check for tenant override
	if evalCtx.TenantID() != "" {
		if override, exists := flag.GetOverride("tenant", evalCtx.TenantID()); exists {
			variant, _ := flag.GetVariant(override.VariantKey)
			return domain.NewEvaluationResult(flag.Key(), true, variant, domain.ReasonOverride)
		}
	}

	// Evaluate rules by priority
	rules := flag.Rules()
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority() < rules[j].Priority()
	})

	for _, rule := range rules {
		if q.matchesRule(rule, evalCtx, flag.Key()) {
			variant, _ := flag.GetVariant(rule.VariantKey())
			result := domain.NewEvaluationResult(flag.Key(), true, variant, domain.ReasonRule)
			return result.WithMatchedRule(&rule)
		}
	}

	// Return default variant
	defaultVariant := flag.GetDefaultVariantValue()
	return domain.NewEvaluationResult(flag.Key(), true, defaultVariant, domain.ReasonDefault)
}

// matchesRule checks if an evaluation context matches a rule.
func (q *EvaluateFlagQuery) matchesRule(rule domain.Rule, evalCtx *domain.EvaluationContext, flagKey string) bool {
	switch rule.Type() {
	case domain.RuleTypeTenant:
		return q.matchesTenantRule(rule, evalCtx)
	case domain.RuleTypeUser:
		return q.matchesUserRule(rule, evalCtx)
	case domain.RuleTypePercent:
		return q.matchesPercentRule(rule, evalCtx, flagKey)
	case domain.RuleTypeAttribute:
		return q.matchesAttributeRule(rule, evalCtx)
	default:
		return false
	}
}

// matchesTenantRule checks if tenant ID is in the rule's values.
func (q *EvaluateFlagQuery) matchesTenantRule(rule domain.Rule, evalCtx *domain.EvaluationContext) bool {
	tenantID := evalCtx.TenantID()
	if tenantID == "" {
		return false
	}

	for _, v := range rule.Values() {
		if v == tenantID {
			return true
		}
	}
	return false
}

// matchesUserRule checks if user ID is in the rule's values.
func (q *EvaluateFlagQuery) matchesUserRule(rule domain.Rule, evalCtx *domain.EvaluationContext) bool {
	userID := evalCtx.UserID()
	if userID == "" {
		return false
	}

	for _, v := range rule.Values() {
		if v == userID {
			return true
		}
	}
	return false
}

// matchesPercentRule checks if the context falls within the percentage.
func (q *EvaluateFlagQuery) matchesPercentRule(rule domain.Rule, evalCtx *domain.EvaluationContext, flagKey string) bool {
	bucket := evalCtx.HashBucket(flagKey)
	return bucket < rule.Percentage()
}

// matchesAttributeRule checks if an attribute matches the rule.
func (q *EvaluateFlagQuery) matchesAttributeRule(rule domain.Rule, evalCtx *domain.EvaluationContext) bool {
	attrValue, exists := evalCtx.GetAttribute(rule.Attribute())

	switch rule.Operator() {
	case domain.OperatorExists:
		return exists
	case domain.OperatorNotExists:
		return !exists
	case domain.OperatorEquals:
		return exists && len(rule.Values()) > 0 && attrValue == rule.Values()[0]
	case domain.OperatorNotEquals:
		return exists && len(rule.Values()) > 0 && attrValue != rule.Values()[0]
	case domain.OperatorIn:
		if !exists {
			return false
		}
		for _, v := range rule.Values() {
			if v == attrValue {
				return true
			}
		}
		return false
	case domain.OperatorNotIn:
		if !exists {
			return true
		}
		for _, v := range rule.Values() {
			if v == attrValue {
				return false
			}
		}
		return true
	case domain.OperatorContains:
		return exists && len(rule.Values()) > 0 && contains(attrValue, rule.Values()[0])
	case domain.OperatorNotContains:
		return exists && len(rule.Values()) > 0 && !contains(attrValue, rule.Values()[0])
	default:
		return false
	}
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && searchString(s, substr)
}

// searchString searches for substr in s.
func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
