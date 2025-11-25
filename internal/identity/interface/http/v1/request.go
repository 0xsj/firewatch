package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/0xsj/hexagonal-go/internal/identity/application/dto"
	"github.com/0xsj/hexagonal-go/pkg/http/response"
	"github.com/0xsj/hexagonal-go/pkg/types"
	"github.com/go-chi/chi/v5"
)

// ParseRegisterRequest parses and validates a user registration request.
func ParseRegisterRequest(r *http.Request) (dto.RegisterUserRequest, error) {
	var req dto.RegisterUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, fmt.Errorf("invalid JSON body: %w", err)
	}

	// Basic validation
	if err := validateRegisterRequest(req); err != nil {
		return req, err
	}

	return req, nil
}

// ParseLoginRequest parses and validates a login request.
// Extracts IP address and user agent from request context.
func ParseLoginRequest(r *http.Request) (dto.LoginRequest, error) {
	var req dto.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, fmt.Errorf("invalid JSON body: %w", err)
	}

	// Extract metadata from request (not from body for security)
	req.IPAddress = extractIPAddress(r)
	req.UserAgent = r.Header.Get("User-Agent")

	// Basic validation
	if err := validateLoginRequest(req); err != nil {
		return req, err
	}

	return req, nil
}

// Add after ParseLoginRequest

// ParseRefreshTokenRequest parses and validates a refresh token request.
func ParseRefreshTokenRequest(r *http.Request) (dto.RefreshTokenRequest, error) {
	var req dto.RefreshTokenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, fmt.Errorf("invalid JSON body: %w", err)
	}

	// Basic validation
	if req.RefreshToken == "" {
		return req, fmt.Errorf("refresh_token is required")
	}

	return req, nil
}

// ParseVerifyEmailRequest parses an email verification request.
// Token can come from query parameter or JSON body.
func ParseVerifyEmailRequest(r *http.Request) (dto.VerifyEmailRequest, error) {
	// Try query parameter first (common for email links)
	token := r.URL.Query().Get("token")

	// If not in query, try JSON body
	if token == "" {
		var req dto.VerifyEmailRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return req, fmt.Errorf("invalid JSON body: %w", err)
		}
		token = req.Token
	}

	if token == "" {
		return dto.VerifyEmailRequest{}, fmt.Errorf("token is required")
	}

	return dto.VerifyEmailRequest{Token: token}, nil
}

// ParseListUsersRequest parses query parameters for listing users.
func ParseListUsersRequest(r *http.Request) (dto.ListUsersRequest, error) {
	req := dto.ListUsersRequest{
		Limit:     50, // Default limit
		Offset:    0,  // Default offset
		SortBy:    "created_at",
		SortOrder: "desc",
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

	if role := r.URL.Query().Get("role"); role != "" {
		req.Role = &role
	}

	if emailVerifiedStr := r.URL.Query().Get("email_verified"); emailVerifiedStr != "" {
		emailVerified, err := strconv.ParseBool(emailVerifiedStr)
		if err != nil {
			return req, fmt.Errorf("invalid email_verified parameter")
		}
		req.EmailVerified = &emailVerified
	}

	if emailContains := r.URL.Query().Get("email_contains"); emailContains != "" {
		req.EmailContains = emailContains
	}

	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		req.SortBy = sortBy
	}

	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder != "" {
		req.SortOrder = sortOrder
	}

	// Validate
	if err := validateListUsersRequest(req); err != nil {
		return req, err
	}

	return req, nil
}

// ParseUserID extracts and validates user ID from URL path parameter.
func ParseUserID(r *http.Request) (types.ID, error) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		return types.ID{}, fmt.Errorf("user ID is required")
	}

	id, err := types.ParseID(idStr)
	if err != nil {
		return types.ID{}, fmt.Errorf("invalid user ID format: %w", err)
	}

	return id, nil
}

// ============================================================================
// Validation Helpers
// ============================================================================

// validateRegisterRequest validates registration request fields.
func validateRegisterRequest(req dto.RegisterUserRequest) error {
	if req.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	return nil
}

// validateLoginRequest validates login request fields.
func validateLoginRequest(req dto.LoginRequest) error {
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

// validateListUsersRequest validates list users request parameters.
func validateListUsersRequest(req dto.ListUsersRequest) error {
	if req.Limit < 1 {
		return fmt.Errorf("limit must be at least 1")
	}
	if req.Limit > 100 {
		return fmt.Errorf("limit cannot exceed 100")
	}
	if req.Offset < 0 {
		return fmt.Errorf("offset cannot be negative")
	}

	// Validate sort_by
	validSortFields := map[string]bool{
		"created_at":    true,
		"updated_at":    true,
		"email":         true,
		"last_login_at": true,
	}
	if !validSortFields[req.SortBy] {
		return fmt.Errorf("invalid sort_by field: %s", req.SortBy)
	}

	// Validate sort_order
	if req.SortOrder != "asc" && req.SortOrder != "desc" {
		return fmt.Errorf("sort_order must be 'asc' or 'desc'")
	}

	return nil
}

// ============================================================================
// Request Metadata Extraction
// ============================================================================

// extractIPAddress extracts the client's IP address from the request.
// Checks X-Forwarded-For, X-Real-IP headers (from proxies) before falling back to RemoteAddr.
func extractIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header (from load balancers/proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs: "client, proxy1, proxy2"
		// Take the first one (the original client)
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header (from proxies)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	// RemoteAddr format: "192.168.1.1:54321" - strip port
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}

	return r.RemoteAddr
}

// ============================================================================
// Error Response Helpers
// ============================================================================

// RespondValidationError writes a validation error response.
func RespondValidationError(w http.ResponseWriter, err error) {
	response.ValidationError(w, map[string]string{
		"error": err.Error(),
	})
}

// ParseRequestPasswordResetRequest parses a password reset request.
func ParseRequestPasswordResetRequest(r *http.Request) (dto.RequestPasswordResetRequest, error) {
	var req dto.RequestPasswordResetRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, fmt.Errorf("invalid JSON body: %w", err)
	}

	// Extract metadata
	req.IPAddress = extractIPAddress(r)
	req.UserAgent = r.Header.Get("User-Agent")

	// Validate
	if req.Email == "" {
		return req, fmt.Errorf("email is required")
	}

	return req, nil
}

// ParseResetPasswordRequest parses a reset password request.
func ParseResetPasswordRequest(r *http.Request) (dto.ResetPasswordRequest, error) {
	var req dto.ResetPasswordRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, fmt.Errorf("invalid JSON body: %w", err)
	}

	// Extract metadata (for audit trail)
	req.IPAddress = extractIPAddress(r)

	// Validate
	if req.Token == "" {
		return req, fmt.Errorf("token is required")
	}
	if req.NewPassword == "" {
		return req, fmt.Errorf("new_password is required")
	}
	if len(req.NewPassword) < 8 {
		return req, fmt.Errorf("password must be at least 8 characters")
	}

	return req, nil
}

// ParseChangePasswordRequest parses a change password request.
func ParseChangePasswordRequest(r *http.Request) (dto.ChangePasswordRequest, error) {
	var req dto.ChangePasswordRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, fmt.Errorf("invalid JSON body: %w", err)
	}

	// Extract metadata (for audit trail)
	req.IPAddress = extractIPAddress(r)

	// Validate
	if req.OldPassword == "" {
		return req, fmt.Errorf("old_password is required")
	}
	if req.NewPassword == "" {
		return req, fmt.Errorf("new_password is required")
	}
	if len(req.NewPassword) < 8 {
		return req, fmt.Errorf("new password must be at least 8 characters")
	}

	return req, nil
}

// ParseSuspendUserRequest parses a suspend user request.
func ParseSuspendUserRequest(r *http.Request) (dto.SuspendUserRequest, error) {
	var req dto.SuspendUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, fmt.Errorf("invalid JSON body: %w", err)
	}

	// Validate
	if req.Reason == "" {
		return req, fmt.Errorf("reason is required")
	}

	return req, nil
}

// ParseReactivateUserRequest parses a reactivate user request.
func ParseReactivateUserRequest(r *http.Request) (dto.ReactivateUserRequest, error) {
	// No body needed, just return empty struct
	return dto.ReactivateUserRequest{}, nil
}

// ParseChangeRoleRequest parses a change role request.
func ParseChangeRoleRequest(r *http.Request) (dto.ChangeRoleRequest, error) {
	var req dto.ChangeRoleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, fmt.Errorf("invalid JSON body: %w", err)
	}

	// Validate
	if req.Role == "" {
		return req, fmt.Errorf("role is required")
	}

	// Validate role is one of the allowed values
	validRoles := map[string]bool{
		"guest":       true,
		"user":        true,
		"moderator":   true,
		"admin":       true,
		"super_admin": true,
	}
	if !validRoles[req.Role] {
		return req, fmt.Errorf("invalid role: must be one of guest, user, moderator, admin, super_admin")
	}

	return req, nil
}

// ParseDeleteUserRequest parses a delete user request.
func ParseDeleteUserRequest(r *http.Request) (dto.DeleteUserRequest, error) {
	var req dto.DeleteUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, fmt.Errorf("invalid JSON body: %w", err)
	}

	// Validate
	if req.Reason == "" {
		return req, fmt.Errorf("reason is required")
	}

	return req, nil
}
