package flags

import (
	"context"
	"sync"
	"time"

	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// Repository defines the interface for flag data access.
// This is implemented by internal/flags/infrastructure/repository.
type Repository interface {
	FindByKey(ctx context.Context, tenantID, key string) (Flag, error)
}

// Flag defines the interface for a feature flag.
// This is implemented by internal/flags/domain.Flag.
type Flag interface {
	Key() string
	Enabled() bool
	Rules() []Rule
	GetOverride(targetType, targetID string) (Override, bool)
	GetVariant(key string) (Variant, bool)
	GetDefaultVariantValue() Variant
}

// Rule defines the interface for a targeting rule.
type Rule interface {
	ID() string
	Type() string
	Attribute() string
	Operator() string
	Values() []string
	Percentage() int
	VariantKey() string
	Priority() int
}

// Variant defines the interface for a flag variant.
type Variant interface {
	Key() string
	Value() string
}

// Override represents a flag override.
type Override struct {
	TargetType string
	TargetID   string
	VariantKey string
}

// Client provides a simple interface for evaluating feature flags.
type Client interface {
	// IsEnabled checks if a flag is enabled for the current context.
	IsEnabled(ctx context.Context, key string) bool

	// IsEnabledFor checks if a flag is enabled for a specific evaluation context.
	IsEnabledFor(ctx context.Context, key string, evalCtx *EvaluationContext) bool

	// GetVariant returns the variant value for a flag.
	GetVariant(ctx context.Context, key string) string

	// GetVariantFor returns the variant value for a specific evaluation context.
	GetVariantFor(ctx context.Context, key string, evalCtx *EvaluationContext) string

	// Evaluate returns the full evaluation result.
	Evaluate(ctx context.Context, key string) *Result

	// EvaluateFor returns the full evaluation result for a specific context.
	EvaluateFor(ctx context.Context, key string, evalCtx *EvaluationContext) *Result
}

// cachedFlag holds a cached flag with expiration.
type cachedFlag struct {
	flag      Flag
	expiresAt time.Time
}

// DefaultClient is the default implementation of the Client interface.
type DefaultClient struct {
	repo   Repository
	logger logger.Logger

	// Optional caching
	cache        map[string]*cachedFlag
	cacheMu      sync.RWMutex
	cacheTTL     time.Duration
	cacheEnabled bool

	// Default tenant
	defaultTenantID string
}

// NewClient creates a new feature flag client.
func NewClient(repo Repository, log logger.Logger, opts ...ClientOption) *DefaultClient {
	c := &DefaultClient{
		repo:   repo,
		logger: log,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// IsEnabled checks if a flag is enabled for the current context.
func (c *DefaultClient) IsEnabled(ctx context.Context, key string) bool {
	evalCtx := FromContext(ctx)
	return c.IsEnabledFor(ctx, key, evalCtx)
}

// IsEnabledFor checks if a flag is enabled for a specific evaluation context.
func (c *DefaultClient) IsEnabledFor(ctx context.Context, key string, evalCtx *EvaluationContext) bool {
	result := c.EvaluateFor(ctx, key, evalCtx)
	return result.Enabled
}

// GetVariant returns the variant value for a flag.
func (c *DefaultClient) GetVariant(ctx context.Context, key string) string {
	evalCtx := FromContext(ctx)
	return c.GetVariantFor(ctx, key, evalCtx)
}

// GetVariantFor returns the variant value for a specific evaluation context.
func (c *DefaultClient) GetVariantFor(ctx context.Context, key string, evalCtx *EvaluationContext) string {
	result := c.EvaluateFor(ctx, key, evalCtx)
	return result.Variant
}

// Evaluate returns the full evaluation result.
func (c *DefaultClient) Evaluate(ctx context.Context, key string) *Result {
	evalCtx := FromContext(ctx)
	return c.EvaluateFor(ctx, key, evalCtx)
}

// EvaluateFor returns the full evaluation result for a specific context.
func (c *DefaultClient) EvaluateFor(ctx context.Context, key string, evalCtx *EvaluationContext) *Result {
	// Use default tenant if not set
	tenantID := evalCtx.TenantID()
	if tenantID == "" {
		tenantID = c.defaultTenantID
	}

	if tenantID == "" {
		c.logger.Warn("no tenant ID for flag evaluation",
			logger.String("key", key),
		)
		return &Result{
			Key:     key,
			Enabled: false,
			Reason:  ReasonDisabled,
		}
	}

	// Get flag
	flag, err := c.getFlag(ctx, tenantID, key)
	if err != nil {
		c.logger.Warn("failed to get flag, returning disabled",
			logger.String("key", key),
			logger.Err(err),
		)
		return &Result{
			Key:     key,
			Enabled: false,
			Reason:  ReasonDisabled,
		}
	}

	// Evaluate
	return c.evaluate(flag, evalCtx)
}

// getFlag retrieves a flag, using cache if enabled.
func (c *DefaultClient) getFlag(ctx context.Context, tenantID, key string) (Flag, error) {
	cacheKey := tenantID + ":" + key

	// Check cache first
	if c.cacheEnabled {
		c.cacheMu.RLock()
		if cached, ok := c.cache[cacheKey]; ok && time.Now().Before(cached.expiresAt) {
			c.cacheMu.RUnlock()
			return cached.flag, nil
		}
		c.cacheMu.RUnlock()
	}

	// Fetch from repository
	flag, err := c.repo.FindByKey(ctx, tenantID, key)
	if err != nil {
		return nil, err
	}

	// Update cache
	if c.cacheEnabled {
		c.cacheMu.Lock()
		c.cache[cacheKey] = &cachedFlag{
			flag:      flag,
			expiresAt: time.Now().Add(c.cacheTTL),
		}
		c.cacheMu.Unlock()
	}

	return flag, nil
}

// evaluate performs the flag evaluation logic.
func (c *DefaultClient) evaluate(flag Flag, evalCtx *EvaluationContext) *Result {
	result := &Result{
		Key: flag.Key(),
	}

	// If flag is disabled, return disabled result
	if !flag.Enabled() {
		result.Enabled = false
		result.Reason = ReasonDisabled
		return result
	}

	result.Enabled = true

	// Check for user override
	if evalCtx.UserID() != "" {
		if override, exists := flag.GetOverride("user", evalCtx.UserID()); exists {
			if variant, ok := flag.GetVariant(override.VariantKey); ok {
				result.Variant = variant.Value()
			}
			result.Reason = ReasonOverride
			return result
		}
	}

	// Check for tenant override
	if evalCtx.TenantID() != "" {
		if override, exists := flag.GetOverride("tenant", evalCtx.TenantID()); exists {
			if variant, ok := flag.GetVariant(override.VariantKey); ok {
				result.Variant = variant.Value()
			}
			result.Reason = ReasonOverride
			return result
		}
	}

	// Evaluate rules by priority
	for _, rule := range flag.Rules() {
		if c.matchesRule(rule, evalCtx, flag.Key()) {
			if variant, ok := flag.GetVariant(rule.VariantKey()); ok {
				result.Variant = variant.Value()
			}
			result.Reason = ReasonRule
			result.MatchedRule = rule.ID()
			return result
		}
	}

	// Return default variant
	defaultVariant := flag.GetDefaultVariantValue()
	if defaultVariant != nil {
		result.Variant = defaultVariant.Value()
	}
	result.Reason = ReasonDefault
	return result
}

// matchesRule checks if an evaluation context matches a rule.
func (c *DefaultClient) matchesRule(rule Rule, evalCtx *EvaluationContext, flagKey string) bool {
	switch rule.Type() {
	case "tenant":
		return c.matchesTenantRule(rule, evalCtx)
	case "user":
		return c.matchesUserRule(rule, evalCtx)
	case "percent":
		return c.matchesPercentRule(rule, evalCtx, flagKey)
	case "attribute":
		return c.matchesAttributeRule(rule, evalCtx)
	default:
		return false
	}
}

// matchesTenantRule checks if tenant ID is in the rule's values.
func (c *DefaultClient) matchesTenantRule(rule Rule, evalCtx *EvaluationContext) bool {
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
func (c *DefaultClient) matchesUserRule(rule Rule, evalCtx *EvaluationContext) bool {
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
func (c *DefaultClient) matchesPercentRule(rule Rule, evalCtx *EvaluationContext, flagKey string) bool {
	bucket := evalCtx.HashBucket(flagKey)
	return bucket < rule.Percentage()
}

// matchesAttributeRule checks if an attribute matches the rule.
func (c *DefaultClient) matchesAttributeRule(rule Rule, evalCtx *EvaluationContext) bool {
	attrValue, exists := evalCtx.GetAttribute(rule.Attribute())

	switch rule.Operator() {
	case "exists":
		return exists
	case "not_exists":
		return !exists
	case "equals":
		return exists && len(rule.Values()) > 0 && attrValue == rule.Values()[0]
	case "not_equals":
		return exists && len(rule.Values()) > 0 && attrValue != rule.Values()[0]
	case "in":
		if !exists {
			return false
		}
		for _, v := range rule.Values() {
			if v == attrValue {
				return true
			}
		}
		return false
	case "not_in":
		if !exists {
			return true
		}
		for _, v := range rule.Values() {
			if v == attrValue {
				return false
			}
		}
		return true
	case "contains":
		return exists && len(rule.Values()) > 0 && contains(attrValue, rule.Values()[0])
	case "not_contains":
		return exists && len(rule.Values()) > 0 && !contains(attrValue, rule.Values()[0])
	default:
		return false
	}
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	if len(substr) == 0 || len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
