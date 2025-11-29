package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/hexagonal-go/internal/flags/application/dto"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// ParseFlagID extracts and validates flag ID from URL path parameter.
func ParseFlagID(r *http.Request) (types.ID, error) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		return types.ID{}, fmt.Errorf("flag ID is required")
	}

	id, err := types.ParseID(idStr)
	if err != nil {
		return types.ID{}, fmt.Errorf("invalid flag ID format: %w", err)
	}

	return id, nil
}

// ParseRuleID extracts and validates rule ID from URL path parameter.
func ParseRuleID(r *http.Request) (types.ID, error) {
	idStr := chi.URLParam(r, "ruleId")
	if idStr == "" {
		return types.ID{}, fmt.Errorf("rule ID is required")
	}

	id, err := types.ParseID(idStr)
	if err != nil {
		return types.ID{}, fmt.Errorf("invalid rule ID format: %w", err)
	}

	return id, nil
}

// ParseCreateFlagRequest parses and validates a create flag request.
func ParseCreateFlagRequest(r *http.Request) (dto.CreateFlagRequest, error) {
	var req dto.CreateFlagRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, fmt.Errorf("invalid JSON body: %w", err)
	}

	if err := validateCreateFlagRequest(req); err != nil {
		return req, err
	}

	return req, nil
}

// ParseUpdateFlagRequest parses and validates an update flag request.
func ParseUpdateFlagRequest(r *http.Request) (dto.UpdateFlagRequest, error) {
	var req dto.UpdateFlagRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, fmt.Errorf("invalid JSON body: %w", err)
	}

	return req, nil
}

// ParseAddVariantRequest parses and validates an add variant request.
func ParseAddVariantRequest(r *http.Request) (dto.AddVariantRequest, error) {
	var req dto.AddVariantRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, fmt.Errorf("invalid JSON body: %w", err)
	}

	if err := validateAddVariantRequest(req); err != nil {
		return req, err
	}

	return req, nil
}

// ParseAddRuleRequest parses and validates an add rule request.
func ParseAddRuleRequest(r *http.Request) (dto.AddRuleRequest, error) {
	var req dto.AddRuleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, fmt.Errorf("invalid JSON body: %w", err)
	}

	if err := validateAddRuleRequest(req); err != nil {
		return req, err
	}

	return req, nil
}

// ParseSetOverrideRequest parses and validates a set override request.
func ParseSetOverrideRequest(r *http.Request) (dto.SetOverrideRequest, error) {
	var req dto.SetOverrideRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, fmt.Errorf("invalid JSON body: %w", err)
	}

	if err := validateSetOverrideRequest(req); err != nil {
		return req, err
	}

	return req, nil
}

// ParseRemoveOverrideRequest parses and validates a remove override request.
func ParseRemoveOverrideRequest(r *http.Request) (dto.RemoveOverrideRequest, error) {
	var req dto.RemoveOverrideRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, fmt.Errorf("invalid JSON body: %w", err)
	}

	if err := validateRemoveOverrideRequest(req); err != nil {
		return req, err
	}

	return req, nil
}

// ParseEvaluateFlagRequest parses and validates an evaluate flag request.
func ParseEvaluateFlagRequest(r *http.Request) (dto.EvaluateFlagRequest, error) {
	var req dto.EvaluateFlagRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, fmt.Errorf("invalid JSON body: %w", err)
	}

	return req, nil
}

// ParseListFlagsRequest parses query parameters for listing flags.
func ParseListFlagsRequest(r *http.Request) dto.ListFlagsRequest {
	req := dto.ListFlagsRequest{
		Limit:  50,
		Offset: 0,
	}

	if enabledStr := r.URL.Query().Get("enabled"); enabledStr != "" {
		enabled := enabledStr == "true"
		req.Enabled = &enabled
	}

	if search := r.URL.Query().Get("search"); search != "" {
		req.Search = search
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

	return req
}

// ============================================================================
// Validation Helpers
// ============================================================================

func validateCreateFlagRequest(req dto.CreateFlagRequest) error {
	if req.Key == "" {
		return fmt.Errorf("key is required")
	}
	return nil
}

func validateAddVariantRequest(req dto.AddVariantRequest) error {
	if req.Key == "" {
		return fmt.Errorf("key is required")
	}
	if req.Value == "" {
		return fmt.Errorf("value is required")
	}
	return nil
}

func validateAddRuleRequest(req dto.AddRuleRequest) error {
	if req.Type == "" {
		return fmt.Errorf("type is required")
	}
	if req.VariantKey == "" {
		return fmt.Errorf("variant_key is required")
	}
	return nil
}

func validateSetOverrideRequest(req dto.SetOverrideRequest) error {
	if req.TargetType == "" {
		return fmt.Errorf("target_type is required")
	}
	if req.TargetID == "" {
		return fmt.Errorf("target_id is required")
	}
	if req.VariantKey == "" {
		return fmt.Errorf("variant_key is required")
	}
	return nil
}

func validateRemoveOverrideRequest(req dto.RemoveOverrideRequest) error {
	if req.TargetType == "" {
		return fmt.Errorf("target_type is required")
	}
	if req.TargetID == "" {
		return fmt.Errorf("target_id is required")
	}
	return nil
}
