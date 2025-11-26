package dto

// CreateTenantRequest represents a request to create a new tenant.
type CreateTenantRequest struct {
	Slug    string `json:"slug"`
	Name    string `json:"name"`
	Plan    string `json:"plan,omitempty"`
	OwnerID string `json:"owner_id"`
}

// CreateTenantWithTrialRequest represents a request to create a tenant with trial.
type CreateTenantWithTrialRequest struct {
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	Plan      string `json:"plan,omitempty"`
	OwnerID   string `json:"owner_id"`
	TrialDays int    `json:"trial_days,omitempty"`
}

// UpdateTenantRequest represents a request to update tenant details.
type UpdateTenantRequest struct {
	Name *string `json:"name,omitempty"`
}

// UpdateSettingsRequest represents a request to update tenant settings.
type UpdateSettingsRequest struct {
	Settings map[string]any `json:"settings"`
}

// ChangePlanRequest represents a request to change the tenant's plan.
type ChangePlanRequest struct {
	Plan   string `json:"plan"`
	Reason string `json:"reason,omitempty"`
}

// SuspendTenantRequest represents a request to suspend a tenant.
type SuspendTenantRequest struct {
	Reason string `json:"reason"`
}

// DeleteTenantRequest represents a request to delete a tenant.
type DeleteTenantRequest struct {
	Reason string `json:"reason,omitempty"`
}

// ListTenantsRequest represents a request to list tenants.
type ListTenantsRequest struct {
	Status  *string `json:"status,omitempty"`
	Plan    *string `json:"plan,omitempty"`
	OwnerID *string `json:"owner_id,omitempty"`
	Search  *string `json:"search,omitempty"`
	Offset  int     `json:"offset,omitempty"`
	Limit   int     `json:"limit,omitempty"`
}
