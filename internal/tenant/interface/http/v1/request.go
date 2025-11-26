package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/0xsj/hexagonal-go/internal/tenant/application/dto"
	"github.com/0xsj/hexagonal-go/pkg/types"
	"github.com/go-chi/chi/v5"
)

// ParseCreateTenantRequest parses and validates a create tenant request.
func ParseCreateTenantRequest(r *http.Request) (dto.CreateTenantRequest, error) {
	var req dto.CreateTenantRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, fmt.Errorf("invalid JSON body: %w", err)
	}

	if err := validateCreateTenantRequest(req); err != nil {
		return req, err
	}

	return req, nil
}

// ParseUpdateTenantRequest parses and validates an update tenant request.
func ParseUpdateTenantRequest(r *http.Request) (dto.UpdateTenantRequest, error) {
	var req dto.UpdateTenantRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, fmt.Errorf("invalid JSON body: %w", err)
	}

	return req, nil
}

// ParseUpdateSettingsRequest parses and validates an update settings request.
func ParseUpdateSettingsRequest(r *http.Request) (dto.UpdateSettingsRequest, error) {
	var req dto.UpdateSettingsRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, fmt.Errorf("invalid JSON body: %w", err)
	}

	if req.Settings == nil {
		req.Settings = make(map[string]any)
	}

	return req, nil
}

// ParseChangePlanRequest parses and validates a change plan request.
func ParseChangePlanRequest(r *http.Request) (dto.ChangePlanRequest, error) {
	var req dto.ChangePlanRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, fmt.Errorf("invalid JSON body: %w", err)
	}

	if req.Plan == "" {
		return req, fmt.Errorf("plan is required")
	}

	return req, nil
}

// ParseSuspendTenantRequest parses and validates a suspend tenant request.
func ParseSuspendTenantRequest(r *http.Request) (dto.SuspendTenantRequest, error) {
	var req dto.SuspendTenantRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, fmt.Errorf("invalid JSON body: %w", err)
	}

	if req.Reason == "" {
		return req, fmt.Errorf("reason is required")
	}

	return req, nil
}

// ParseDeleteTenantRequest parses a delete tenant request.
func ParseDeleteTenantRequest(r *http.Request) (dto.DeleteTenantRequest, error) {
	var req dto.DeleteTenantRequest

	// Body is optional for delete
	_ = json.NewDecoder(r.Body).Decode(&req)

	return req, nil
}

// ParseListTenantsRequest parses query parameters for listing tenants.
func ParseListTenantsRequest(r *http.Request) (dto.ListTenantsRequest, error) {
	req := dto.ListTenantsRequest{
		Limit:  20,
		Offset: 0,
	}

	// Parse query parameters
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return req, fmt.Errorf("invalid limit parameter")
		}
		req.Limit = limit
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return req, fmt.Errorf("invalid offset parameter")
		}
		req.Offset = offset
	}

	// Optional filters
	if status := r.URL.Query().Get("status"); status != "" {
		req.Status = &status
	}

	if plan := r.URL.Query().Get("plan"); plan != "" {
		req.Plan = &plan
	}

	if ownerID := r.URL.Query().Get("owner_id"); ownerID != "" {
		req.OwnerID = &ownerID
	}

	if search := r.URL.Query().Get("search"); search != "" {
		req.Search = &search
	}

	// Validate
	if err := validateListTenantsRequest(req); err != nil {
		return req, err
	}

	return req, nil
}

// ParseTenantID extracts and validates tenant ID from URL path parameter.
func ParseTenantID(r *http.Request) (types.ID, error) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		return types.ID{}, fmt.Errorf("tenant ID is required")
	}

	id, err := types.ParseID(idStr)
	if err != nil {
		return types.ID{}, fmt.Errorf("invalid tenant ID format: %w", err)
	}

	return id, nil
}

// ParseTenantSlug extracts tenant slug from URL path parameter.
func ParseTenantSlug(r *http.Request) (string, error) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		return "", fmt.Errorf("tenant slug is required")
	}

	return slug, nil
}

// ============================================================================
// Validation Helpers
// ============================================================================

func validateCreateTenantRequest(req dto.CreateTenantRequest) error {
	if req.Slug == "" {
		return fmt.Errorf("slug is required")
	}
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.OwnerID == "" {
		return fmt.Errorf("owner_id is required")
	}
	return nil
}

func validateListTenantsRequest(req dto.ListTenantsRequest) error {
	if req.Limit < 1 {
		return fmt.Errorf("limit must be at least 1")
	}
	if req.Limit > 100 {
		return fmt.Errorf("limit cannot exceed 100")
	}
	if req.Offset < 0 {
		return fmt.Errorf("offset cannot be negative")
	}
	return nil
}
