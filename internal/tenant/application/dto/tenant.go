package dto

import (
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// TenantDTO represents a tenant in API responses.
type TenantDTO struct {
	ID          string         `json:"id"`
	Slug        string         `json:"slug"`
	Name        string         `json:"name"`
	Plan        string         `json:"plan"`
	Status      string         `json:"status"`
	Settings    map[string]any `json:"settings,omitempty"`
	OwnerID     string         `json:"owner_id"`
	OwnerEmail  string         `json:"owner_email,omitempty"`
	BillingID   *string        `json:"billing_id,omitempty"`
	TrialEndsAt *string        `json:"trial_ends_at,omitempty"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
}

// TenantSummaryDTO represents a tenant summary for list responses.
type TenantSummaryDTO struct {
	ID        string `json:"id"`
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	Plan      string `json:"plan"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// TenantListDTO represents a paginated list of tenants.
type TenantListDTO struct {
	Tenants []TenantSummaryDTO `json:"tenants"`
	Total   int64              `json:"total"`
	Offset  int                `json:"offset"`
	Limit   int                `json:"limit"`
	HasMore bool               `json:"has_more"`
}

// TenantSettingsDTO represents tenant settings in API responses.
type TenantSettingsDTO struct {
	TenantID string         `json:"tenant_id"`
	Settings map[string]any `json:"settings"`
}

// formatTimestamp formats a timestamp for API responses.
func formatTimestamp(ts types.Timestamp) string {
	return ts.Time().Format("2006-01-02T15:04:05Z07:00")
}

// formatTimestampPtr formats an optional timestamp for API responses.
func formatTimestampPtr(ts *types.Timestamp) *string {
	if ts == nil {
		return nil
	}
	formatted := ts.Time().Format("2006-01-02T15:04:05Z07:00")
	return &formatted
}
