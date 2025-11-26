package dto

import (
	tenant "github.com/0xsj/hexagonal-go/internal/tenant/domain"
)

// ToTenantDTO maps a domain Tenant to a TenantDTO.
func ToTenantDTO(t *tenant.Tenant) *TenantDTO {
	if t == nil {
		return nil
	}

	return &TenantDTO{
		ID:          t.ID().String(),
		Slug:        t.Slug().String(),
		Name:        t.Name(),
		Plan:        t.Plan().String(),
		Status:      t.Status().String(),
		Settings:    t.Settings().ToMap(),
		OwnerID:     t.OwnerID().String(),
		BillingID:   t.BillingID(),
		TrialEndsAt: formatTimestampPtr(t.TrialEndsAt()),
		CreatedAt:   formatTimestamp(t.CreatedAt()),
		UpdatedAt:   formatTimestamp(t.UpdatedAt()),
	}
}

// ToTenantSummaryDTO maps a domain Tenant to a TenantSummaryDTO.
func ToTenantSummaryDTO(t *tenant.Tenant) TenantSummaryDTO {
	return TenantSummaryDTO{
		ID:        t.ID().String(),
		Slug:      t.Slug().String(),
		Name:      t.Name(),
		Plan:      t.Plan().String(),
		Status:    t.Status().String(),
		CreatedAt: formatTimestamp(t.CreatedAt()),
	}
}

// ToTenantSummaryDTOs maps a slice of domain Tenants to TenantSummaryDTOs.
func ToTenantSummaryDTOs(tenants []*tenant.Tenant) []TenantSummaryDTO {
	result := make([]TenantSummaryDTO, len(tenants))
	for i, t := range tenants {
		result[i] = ToTenantSummaryDTO(t)
	}
	return result
}

// ToTenantListDTO maps a slice of tenants with pagination info to TenantListDTO.
func ToTenantListDTO(tenants []*tenant.Tenant, total int64, offset, limit int) *TenantListDTO {
	summaries := ToTenantSummaryDTOs(tenants)
	hasMore := int64(offset+len(tenants)) < total

	return &TenantListDTO{
		Tenants: summaries,
		Total:   total,
		Offset:  offset,
		Limit:   limit,
		HasMore: hasMore,
	}
}

// ToSettingsDTO maps tenant settings to TenantSettingsDTO.
func ToSettingsDTO(t *tenant.Tenant) *TenantSettingsDTO {
	if t == nil {
		return nil
	}

	return &TenantSettingsDTO{
		TenantID: t.ID().String(),
		Settings: t.Settings().ToMap(),
	}
}
