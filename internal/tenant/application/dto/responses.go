package dto

// TenantResponse wraps a single tenant response.
type TenantResponse struct {
	Tenant *TenantDTO `json:"tenant"`
}

// TenantsResponse wraps a list of tenants response.
type TenantsResponse struct {
	Tenants []TenantSummaryDTO `json:"tenants"`
	Total   int64              `json:"total"`
	Offset  int                `json:"offset"`
	Limit   int                `json:"limit"`
	HasMore bool               `json:"has_more"`
}

// SettingsResponse wraps a tenant settings response.
type SettingsResponse struct {
	TenantID string         `json:"tenant_id"`
	Settings map[string]any `json:"settings"`
}

// MessageResponse represents a simple message response.
type MessageResponse struct {
	Message string `json:"message"`
}

// NewTenantResponse creates a new TenantResponse.
func NewTenantResponse(tenant *TenantDTO) *TenantResponse {
	return &TenantResponse{
		Tenant: tenant,
	}
}

// NewTenantsResponse creates a new TenantsResponse.
func NewTenantsResponse(tenants []TenantSummaryDTO, total int64, offset, limit int) *TenantsResponse {
	hasMore := int64(offset+len(tenants)) < total
	return &TenantsResponse{
		Tenants: tenants,
		Total:   total,
		Offset:  offset,
		Limit:   limit,
		HasMore: hasMore,
	}
}

// NewSettingsResponse creates a new SettingsResponse.
func NewSettingsResponse(tenantID string, settings map[string]any) *SettingsResponse {
	return &SettingsResponse{
		TenantID: tenantID,
		Settings: settings,
	}
}

// NewMessageResponse creates a new MessageResponse.
func NewMessageResponse(message string) *MessageResponse {
	return &MessageResponse{
		Message: message,
	}
}
