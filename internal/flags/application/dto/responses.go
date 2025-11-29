package dto

// CreateFlagResponse represents the response after creating a flag.
type CreateFlagResponse struct {
	Flag *FlagDTO `json:"flag"`
}

// UpdateFlagResponse represents the response after updating a flag.
type UpdateFlagResponse struct {
	Flag *FlagDTO `json:"flag"`
}

// GetFlagResponse represents the response for getting a flag.
type GetFlagResponse struct {
	Flag *FlagDTO `json:"flag"`
}

// ListFlagsResponse represents the response for listing flags.
type ListFlagsResponse struct {
	Flags   []*FlagSummaryDTO `json:"flags"`
	Total   int               `json:"total"`
	Limit   int               `json:"limit"`
	Offset  int               `json:"offset"`
	HasMore bool              `json:"has_more"`
}

// DeleteFlagResponse represents the response after deleting a flag.
type DeleteFlagResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id"`
}

// EnableFlagResponse represents the response after enabling a flag.
type EnableFlagResponse struct {
	Flag *FlagDTO `json:"flag"`
}

// DisableFlagResponse represents the response after disabling a flag.
type DisableFlagResponse struct {
	Flag *FlagDTO `json:"flag"`
}

// AddVariantResponse represents the response after adding a variant.
type AddVariantResponse struct {
	Flag *FlagDTO `json:"flag"`
}

// RemoveVariantResponse represents the response after removing a variant.
type RemoveVariantResponse struct {
	Flag *FlagDTO `json:"flag"`
}

// AddRuleResponse represents the response after adding a rule.
type AddRuleResponse struct {
	Flag   *FlagDTO `json:"flag"`
	RuleID string   `json:"rule_id"`
}

// RemoveRuleResponse represents the response after removing a rule.
type RemoveRuleResponse struct {
	Flag *FlagDTO `json:"flag"`
}

// SetOverrideResponse represents the response after setting an override.
type SetOverrideResponse struct {
	Flag *FlagDTO `json:"flag"`
}

// RemoveOverrideResponse represents the response after removing an override.
type RemoveOverrideResponse struct {
	Flag *FlagDTO `json:"flag"`
}

// EvaluateFlagResponse represents the response for flag evaluation.
type EvaluateFlagResponse struct {
	Result *EvaluationResultDTO `json:"result"`
}

// EvaluateFlagsResponse represents the response for bulk flag evaluation.
type EvaluateFlagsResponse struct {
	Results []*EvaluationResultDTO `json:"results"`
}
