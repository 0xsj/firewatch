package dto

import "github.com/0xsj/hexagonal-go/internal/flags/domain"

// MapFlagToDTO converts a domain Flag to a FlagDTO.
func MapFlagToDTO(flag *domain.Flag) *FlagDTO {
	if flag == nil {
		return nil
	}

	return &FlagDTO{
		ID:             flag.ID().String(),
		TenantID:       flag.TenantID(),
		Key:            flag.Key(),
		Name:           flag.Name(),
		Description:    flag.Description(),
		Enabled:        flag.Enabled(),
		Variants:       MapVariantsToDTO(flag.Variants()),
		DefaultVariant: flag.DefaultVariant(),
		Rules:          MapRulesToDTO(flag.Rules()),
		Overrides:      MapOverridesToDTO(flag.Overrides()),
		CreatedAt:      flag.CreatedAt().Time(),
		UpdatedAt:      flag.UpdatedAt().Time(),
		Version:        flag.Version(),
	}
}

// MapFlagToSummaryDTO converts a domain Flag to a FlagSummaryDTO.
func MapFlagToSummaryDTO(flag *domain.Flag) *FlagSummaryDTO {
	if flag == nil {
		return nil
	}

	return &FlagSummaryDTO{
		ID:          flag.ID().String(),
		Key:         flag.Key(),
		Name:        flag.Name(),
		Description: flag.Description(),
		Enabled:     flag.Enabled(),
		RuleCount:   len(flag.Rules()),
		UpdatedAt:   flag.UpdatedAt().Time(),
	}
}

// MapFlagsToSummaryDTO converts a slice of domain Flags to FlagSummaryDTOs.
func MapFlagsToSummaryDTO(flags []*domain.Flag) []*FlagSummaryDTO {
	result := make([]*FlagSummaryDTO, len(flags))
	for i, flag := range flags {
		result[i] = MapFlagToSummaryDTO(flag)
	}
	return result
}

// MapVariantToDTO converts a domain Variant to a VariantDTO.
func MapVariantToDTO(variant domain.Variant) VariantDTO {
	return VariantDTO{
		Key:         variant.Key(),
		Value:       variant.Value(),
		Weight:      variant.Weight(),
		Description: variant.Description(),
	}
}

// MapVariantsToDTO converts a slice of domain Variants to VariantDTOs.
func MapVariantsToDTO(variants []domain.Variant) []VariantDTO {
	result := make([]VariantDTO, len(variants))
	for i, variant := range variants {
		result[i] = MapVariantToDTO(variant)
	}
	return result
}

// MapRuleToDTO converts a domain Rule to a RuleDTO.
func MapRuleToDTO(rule domain.Rule) RuleDTO {
	return RuleDTO{
		ID:         rule.ID().String(),
		Type:       rule.Type().String(),
		Attribute:  rule.Attribute(),
		Operator:   rule.Operator().String(),
		Values:     rule.Values(),
		Percentage: rule.Percentage(),
		VariantKey: rule.VariantKey(),
		Priority:   rule.Priority(),
	}
}

// MapRulesToDTO converts a slice of domain Rules to RuleDTOs.
func MapRulesToDTO(rules []domain.Rule) []RuleDTO {
	result := make([]RuleDTO, len(rules))
	for i, rule := range rules {
		result[i] = MapRuleToDTO(rule)
	}
	return result
}

// MapOverrideToDTO converts a domain Override to an OverrideDTO.
func MapOverrideToDTO(override domain.Override) OverrideDTO {
	return OverrideDTO{
		TargetType: override.TargetType,
		TargetID:   override.TargetID,
		VariantKey: override.VariantKey,
	}
}

// MapOverridesToDTO converts a map of domain Overrides to OverrideDTOs.
func MapOverridesToDTO(overrides map[string]domain.Override) []OverrideDTO {
	result := make([]OverrideDTO, 0, len(overrides))
	for _, override := range overrides {
		result = append(result, MapOverrideToDTO(override))
	}
	return result
}

// MapEvaluationResultToDTO converts a domain EvaluationResult to an EvaluationResultDTO.
func MapEvaluationResultToDTO(result *domain.EvaluationResult) *EvaluationResultDTO {
	if result == nil {
		return nil
	}

	return &EvaluationResultDTO{
		FlagKey:      result.FlagKey(),
		Enabled:      result.Enabled(),
		VariantKey:   result.VariantKey(),
		VariantValue: result.VariantValue(),
		Reason:       string(result.Reason()),
	}
}

// MapEvaluationResultsToDTO converts a slice of domain EvaluationResults to DTOs.
func MapEvaluationResultsToDTO(results []*domain.EvaluationResult) []*EvaluationResultDTO {
	dtos := make([]*EvaluationResultDTO, len(results))
	for i, result := range results {
		dtos[i] = MapEvaluationResultToDTO(result)
	}
	return dtos
}
