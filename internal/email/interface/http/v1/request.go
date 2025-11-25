package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/hexagonal-go/internal/email/application/dto"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// ParseCreateTemplateRequest parses and validates a create template request.
func ParseCreateTemplateRequest(r *http.Request) (dto.CreateTemplateRequest, error) {
	var req dto.CreateTemplateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, fmt.Errorf("invalid JSON body: %w", err)
	}

	if err := validateCreateTemplateRequest(req); err != nil {
		return req, err
	}

	return req, nil
}

// ParseUpdateTemplateRequest parses and validates an update template request.
func ParseUpdateTemplateRequest(r *http.Request) (dto.UpdateTemplateRequest, error) {
	var req dto.UpdateTemplateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, fmt.Errorf("invalid JSON body: %w", err)
	}

	return req, nil
}

// ParseTemplateID extracts and validates template ID from URL path parameter.
func ParseTemplateID(r *http.Request) (types.ID, error) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		return types.ID{}, fmt.Errorf("template ID is required")
	}

	id, err := types.ParseID(idStr)
	if err != nil {
		return types.ID{}, fmt.Errorf("invalid template ID format: %w", err)
	}

	return id, nil
}

// ParseListTemplatesRequest parses query parameters for listing templates.
func ParseListTemplatesRequest(r *http.Request) dto.ListTemplatesRequest {
	req := dto.ListTemplatesRequest{
		IncludeSystemTemplates: true,
		Limit:                  20,
		Offset:                 0,
		SortBy:                 "created_at",
		SortOrder:              "desc",
	}

	if tenantID := r.URL.Query().Get("tenant_id"); tenantID != "" {
		req.TenantID = &tenantID
	}

	if status := r.URL.Query().Get("status"); status != "" {
		req.Status = &status
	}

	if locale := r.URL.Query().Get("locale"); locale != "" {
		req.Locale = &locale
	}

	if slugContains := r.URL.Query().Get("slug_contains"); slugContains != "" {
		req.SlugContains = slugContains
	}

	if nameContains := r.URL.Query().Get("name_contains"); nameContains != "" {
		req.NameContains = nameContains
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			req.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			req.Offset = offset
		}
	}

	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		req.SortBy = sortBy
	}

	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder != "" {
		req.SortOrder = sortOrder
	}

	if includeSystem := r.URL.Query().Get("include_system"); includeSystem == "false" {
		req.IncludeSystemTemplates = false
	}

	return req
}

// ParseGetTemplateBySlugRequest parses query parameters for getting template by slug.
func ParseGetTemplateBySlugRequest(r *http.Request) (dto.GetTemplateBySlugRequest, error) {
	req := dto.GetTemplateBySlugRequest{
		Slug:   r.URL.Query().Get("slug"),
		Locale: r.URL.Query().Get("locale"),
	}

	if tenantID := r.URL.Query().Get("tenant_id"); tenantID != "" {
		req.TenantID = &tenantID
	}

	if req.Slug == "" {
		return req, fmt.Errorf("slug is required")
	}

	if req.Locale == "" {
		req.Locale = "en"
	}

	return req, nil
}

// ParsePreviewTemplateRequest parses a preview template request.
func ParsePreviewTemplateRequest(r *http.Request) (dto.PreviewTemplateRequest, error) {
	templateID, err := ParseTemplateID(r)
	if err != nil {
		return dto.PreviewTemplateRequest{}, err
	}

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return dto.PreviewTemplateRequest{}, fmt.Errorf("invalid JSON body: %w", err)
	}

	return dto.PreviewTemplateRequest{
		ID:   templateID.String(),
		Data: data,
	}, nil
}

// ParsePreviewTemplateBySlugRequest parses a preview by slug request.
func ParsePreviewTemplateBySlugRequest(r *http.Request) (dto.PreviewTemplateBySlugRequest, error) {
	var req dto.PreviewTemplateBySlugRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, fmt.Errorf("invalid JSON body: %w", err)
	}

	if req.Slug == "" {
		return req, fmt.Errorf("slug is required")
	}

	if req.Locale == "" {
		req.Locale = "en"
	}

	return req, nil
}

// ============================================================================
// Validation Helpers
// ============================================================================

func validateCreateTemplateRequest(req dto.CreateTemplateRequest) error {
	if req.Slug == "" {
		return fmt.Errorf("slug is required")
	}
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Subject == "" {
		return fmt.Errorf("subject is required")
	}
	if req.BodyHTML == "" {
		return fmt.Errorf("body_html is required")
	}
	if req.Locale == "" {
		return fmt.Errorf("locale is required")
	}
	return nil
}
