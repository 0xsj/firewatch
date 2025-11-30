package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/0xsj/hexagonal-go/pkg/http/response"
	"github.com/0xsj/hexagonal-go/pkg/types"
	"github.com/go-chi/chi/v5"
)

// ============================================================================
// Request DTOs
// ============================================================================

// CreateRoleRequest represents the request to create a role.
type CreateRoleRequest struct {
	TenantID    string   `json:"tenant_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// UpdateRoleRequest represents the request to update a role.
type UpdateRoleRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// AssignRoleRequest represents the request to assign a role.
type AssignRoleRequest struct {
	UserID     string     `json:"user_id"`
	RoleID     string     `json:"role_id"`
	TenantID   string     `json:"tenant_id,omitempty"`
	ResourceID string     `json:"resource_id,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

// RevokeRoleRequest represents the request to revoke a role.
type RevokeRoleRequest struct {
	UserID     string `json:"user_id"`
	RoleID     string `json:"role_id"`
	TenantID   string `json:"tenant_id,omitempty"`
	ResourceID string `json:"resource_id,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

// ListRolesRequest represents the request to list roles.
type ListRolesRequest struct {
	TenantID      string
	IncludeSystem bool
	Search        string
	Limit         int
	Offset        int
}

// ============================================================================
// Request Parsers
// ============================================================================

// ParseRoleID extracts and validates role ID from URL path parameter.
func ParseRoleID(r *http.Request) (types.ID, error) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		return types.ID{}, fmt.Errorf("role ID is required")
	}

	id, err := types.ParseID(idStr)
	if err != nil {
		return types.ID{}, fmt.Errorf("invalid role ID format: %w", err)
	}

	return id, nil
}

// ParseCreateRoleRequest parses and validates a create role request.
func ParseCreateRoleRequest(r *http.Request) (CreateRoleRequest, error) {
	var req CreateRoleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, fmt.Errorf("invalid JSON body: %w", err)
	}

	if err := validateCreateRoleRequest(req); err != nil {
		return req, err
	}

	return req, nil
}

// ParseUpdateRoleRequest parses and validates an update role request.
func ParseUpdateRoleRequest(r *http.Request) (UpdateRoleRequest, error) {
	var req UpdateRoleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, fmt.Errorf("invalid JSON body: %w", err)
	}

	return req, nil
}

// ParseListRolesRequest parses query parameters for listing roles.
func ParseListRolesRequest(r *http.Request) (ListRolesRequest, error) {
	req := ListRolesRequest{
		Limit:  50,
		Offset: 0,
	}

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

	if includeSystemStr := r.URL.Query().Get("include_system"); includeSystemStr != "" {
		includeSystem, err := strconv.ParseBool(includeSystemStr)
		if err != nil {
			return req, fmt.Errorf("invalid include_system parameter")
		}
		req.IncludeSystem = includeSystem
	}

	if search := r.URL.Query().Get("search"); search != "" {
		req.Search = search
	}

	if err := validateListRolesRequest(req); err != nil {
		return req, err
	}

	return req, nil
}

// ParseAssignRoleRequest parses and validates an assign role request.
func ParseAssignRoleRequest(r *http.Request) (AssignRoleRequest, error) {
	var req AssignRoleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, fmt.Errorf("invalid JSON body: %w", err)
	}

	if err := validateAssignRoleRequest(req); err != nil {
		return req, err
	}

	return req, nil
}

// ParseRevokeRoleRequest parses and validates a revoke role request.
func ParseRevokeRoleRequest(r *http.Request) (RevokeRoleRequest, error) {
	var req RevokeRoleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, fmt.Errorf("invalid JSON body: %w", err)
	}

	if err := validateRevokeRoleRequest(req); err != nil {
		return req, err
	}

	return req, nil
}

// ============================================================================
// Validation Helpers
// ============================================================================

func validateCreateRoleRequest(req CreateRoleRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(req.Name) < 2 {
		return fmt.Errorf("name must be at least 2 characters")
	}

	if len(req.Name) > 64 {
		return fmt.Errorf("name must be at most 64 characters")
	}

	return nil
}

func validateListRolesRequest(req ListRolesRequest) error {
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

func validateAssignRoleRequest(req AssignRoleRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("user_id is required")
	}

	if req.RoleID == "" {
		return fmt.Errorf("role_id is required")
	}

	return nil
}

func validateRevokeRoleRequest(req RevokeRoleRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("user_id is required")
	}

	if req.RoleID == "" {
		return fmt.Errorf("role_id is required")
	}

	return nil
}

// ============================================================================
// Error Response Helper
// ============================================================================

// RespondValidationError writes a validation error response.
func RespondValidationError(w http.ResponseWriter, err error) {
	response.ValidationError(w, map[string]string{
		"error": err.Error(),
	})
}
