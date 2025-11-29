package dto

import "time"

// FlagDTO represents a feature flag for API responses.
type FlagDTO struct {
	ID             string        `json:"id"`
	TenantID       string        `json:"tenant_id"`
	Key            string        `json:"key"`
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	Enabled        bool          `json:"enabled"`
	Variants       []VariantDTO  `json:"variants"`
	DefaultVariant string        `json:"default_variant"`
	Rules          []RuleDTO     `json:"rules,omitempty"`
	Overrides      []OverrideDTO `json:"overrides,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	Version        int           `json:"version"`
}

// VariantDTO represents a flag variant for API responses.
type VariantDTO struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Weight      int    `json:"weight"`
	Description string `json:"description,omitempty"`
}

// RuleDTO represents a targeting rule for API responses.
type RuleDTO struct {
	ID         string   `json:"id"`
	Type       string   `json:"type"`
	Attribute  string   `json:"attribute,omitempty"`
	Operator   string   `json:"operator,omitempty"`
	Values     []string `json:"values,omitempty"`
	Percentage int      `json:"percentage,omitempty"`
	VariantKey string   `json:"variant_key"`
	Priority   int      `json:"priority"`
}

// OverrideDTO represents an override for API responses.
type OverrideDTO struct {
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	VariantKey string `json:"variant_key"`
}

// EvaluationResultDTO represents the result of flag evaluation.
type EvaluationResultDTO struct {
	FlagKey      string `json:"flag_key"`
	Enabled      bool   `json:"enabled"`
	VariantKey   string `json:"variant_key"`
	VariantValue string `json:"variant_value"`
	Reason       string `json:"reason"`
}

// FlagSummaryDTO represents a lightweight flag summary for listings.
type FlagSummaryDTO struct {
	ID          string    `json:"id"`
	Key         string    `json:"key"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled"`
	RuleCount   int       `json:"rule_count"`
	UpdatedAt   time.Time `json:"updated_at"`
}
