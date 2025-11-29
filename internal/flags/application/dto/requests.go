package dto

// CreateFlagRequest represents a request to create a feature flag.
type CreateFlagRequest struct {
	Key         string `json:"key" validate:"required,min=1,max=128"`
	Name        string `json:"name" validate:"max=255"`
	Description string `json:"description" validate:"max=1024"`
	Enabled     bool   `json:"enabled"`
}

// UpdateFlagRequest represents a request to update a feature flag.
type UpdateFlagRequest struct {
	Name        *string `json:"name" validate:"omitempty,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1024"`
}

// AddVariantRequest represents a request to add a variant to a flag.
type AddVariantRequest struct {
	Key         string `json:"key" validate:"required,min=1,max=64"`
	Value       string `json:"value" validate:"required"`
	Weight      int    `json:"weight" validate:"min=0,max=100"`
	Description string `json:"description" validate:"max=255"`
}

// AddRuleRequest represents a request to add a targeting rule.
type AddRuleRequest struct {
	Type       string   `json:"type" validate:"required,oneof=tenant user percent attribute"`
	Attribute  string   `json:"attribute" validate:"required_if=Type attribute"`
	Operator   string   `json:"operator" validate:"required_if=Type attribute"`
	Values     []string `json:"values" validate:"required_if=Type tenant,required_if=Type user"`
	Percentage int      `json:"percentage" validate:"required_if=Type percent,min=0,max=100"`
	VariantKey string   `json:"variant_key" validate:"required"`
	Priority   int      `json:"priority" validate:"min=0"`
}

// SetOverrideRequest represents a request to set an override.
type SetOverrideRequest struct {
	TargetType string `json:"target_type" validate:"required,oneof=tenant user"`
	TargetID   string `json:"target_id" validate:"required"`
	VariantKey string `json:"variant_key" validate:"required"`
}

// RemoveOverrideRequest represents a request to remove an override.
type RemoveOverrideRequest struct {
	TargetType string `json:"target_type" validate:"required,oneof=tenant user"`
	TargetID   string `json:"target_id" validate:"required"`
}

// EvaluateFlagRequest represents a request to evaluate a flag.
type EvaluateFlagRequest struct {
	TenantID   string            `json:"tenant_id"`
	UserID     string            `json:"user_id"`
	Attributes map[string]string `json:"attributes"`
}

// ListFlagsRequest represents a request to list flags.
type ListFlagsRequest struct {
	Enabled *bool  `json:"enabled"`
	Search  string `json:"search"`
	Limit   int    `json:"limit"`
	Offset  int    `json:"offset"`
}
