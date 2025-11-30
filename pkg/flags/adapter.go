package flags

import (
	"context"

	"github.com/0xsj/hexagonal-go/internal/flags/domain"
	"github.com/0xsj/hexagonal-go/internal/flags/infrastructure/repository"
	"github.com/0xsj/hexagonal-go/pkg/database"
)

// ============================================================================
// Repository Adapter
// ============================================================================

// RepositoryAdapter adapts the domain repository to the pkg/flags.Repository interface.
type RepositoryAdapter struct {
	repo *repository.PostgresRepository
}

// NewRepositoryAdapter creates a new RepositoryAdapter.
func NewRepositoryAdapter(db database.DB) *RepositoryAdapter {
	return &RepositoryAdapter{
		repo: repository.NewPostgresRepository(db),
	}
}

// FindByKey retrieves a flag by key and returns it as a pkg/flags.Flag.
func (a *RepositoryAdapter) FindByKey(ctx context.Context, tenantID, key string) (Flag, error) {
	domainFlag, err := a.repo.FindByKey(ctx, tenantID, key)
	if err != nil {
		return nil, err
	}
	return &FlagAdapter{flag: domainFlag}, nil
}

// ============================================================================
// Flag Adapter
// ============================================================================

// FlagAdapter adapts domain.Flag to the pkg/flags.Flag interface.
type FlagAdapter struct {
	flag *domain.Flag
}

// Key returns the flag key.
func (a *FlagAdapter) Key() string {
	return a.flag.Key()
}

// Enabled returns whether the flag is enabled.
func (a *FlagAdapter) Enabled() bool {
	return a.flag.Enabled()
}

// Rules returns the targeting rules as pkg/flags.Rule interfaces.
func (a *FlagAdapter) Rules() []Rule {
	domainRules := a.flag.Rules()
	rules := make([]Rule, len(domainRules))
	for i, r := range domainRules {
		rules[i] = &RuleAdapter{rule: r}
	}
	return rules
}

// GetOverride returns an override for a specific target.
func (a *FlagAdapter) GetOverride(targetType, targetID string) (Override, bool) {
	domainOverride, exists := a.flag.GetOverride(targetType, targetID)
	if !exists {
		return Override{}, false
	}
	return Override{
		TargetType: domainOverride.TargetType,
		TargetID:   domainOverride.TargetID,
		VariantKey: domainOverride.VariantKey,
	}, true
}

// GetVariant returns a variant by key.
func (a *FlagAdapter) GetVariant(key string) (Variant, bool) {
	domainVariant, exists := a.flag.GetVariant(key)
	if !exists {
		return nil, false
	}
	return &VariantAdapter{variant: domainVariant}, true
}

// GetDefaultVariantValue returns the default variant.
func (a *FlagAdapter) GetDefaultVariantValue() Variant {
	domainVariant := a.flag.GetDefaultVariantValue()
	return &VariantAdapter{variant: domainVariant}
}

// ============================================================================
// Rule Adapter
// ============================================================================

// RuleAdapter adapts domain.Rule to the pkg/flags.Rule interface.
type RuleAdapter struct {
	rule domain.Rule
}

// ID returns the rule ID.
func (a *RuleAdapter) ID() string {
	return a.rule.ID().String()
}

// Type returns the rule type.
func (a *RuleAdapter) Type() string {
	return a.rule.Type().String()
}

// Attribute returns the attribute name.
func (a *RuleAdapter) Attribute() string {
	return a.rule.Attribute()
}

// Operator returns the comparison operator.
func (a *RuleAdapter) Operator() string {
	return a.rule.Operator().String()
}

// Values returns the target values.
func (a *RuleAdapter) Values() []string {
	return a.rule.Values()
}

// Percentage returns the percentage for percent rules.
func (a *RuleAdapter) Percentage() int {
	return a.rule.Percentage()
}

// VariantKey returns the variant to serve when rule matches.
func (a *RuleAdapter) VariantKey() string {
	return a.rule.VariantKey()
}

// Priority returns the rule priority.
func (a *RuleAdapter) Priority() int {
	return a.rule.Priority()
}

// ============================================================================
// Variant Adapter
// ============================================================================

// VariantAdapter adapts domain.Variant to the pkg/flags.Variant interface.
type VariantAdapter struct {
	variant domain.Variant
}

// Key returns the variant key.
func (a *VariantAdapter) Key() string {
	return a.variant.Key()
}

// Value returns the variant value.
func (a *VariantAdapter) Value() string {
	return a.variant.Value()
}
