package dto

// RegisterUserRequest represents a user registration request with password.
type RegisterUserRequest struct {
	TenantID string `json:"tenant_id" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// RegisterPasswordlessRequest represents a passwordless user registration.
// Used for OAuth, magic link, or SSO authentication.
type RegisterPasswordlessRequest struct {
	TenantID string `json:"tenant_id" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Source   string `json:"source" validate:"required"` // "magic_link", "google", "github", etc.
}

// LoginRequest represents a user login request.
type LoginRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required"`
	IPAddress string `json:"-"` // Set from request context, not JSON body
	UserAgent string `json:"-"` // Set from request headers, not JSON body
}

// VerifyEmailRequest represents an email verification request.
type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

// ChangePasswordRequest represents a password change request.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
	IPAddress   string `json:"-"` // Set from request context, not JSON body
}

// RequestPasswordResetRequest is the request for requesting a password reset.
type RequestPasswordResetRequest struct {
	Email     string `json:"email" validate:"required,email"`
	IPAddress string `json:"-"` // Set from request context, not JSON body
	UserAgent string `json:"-"` // Set from request headers, not JSON body
}

// ReactivateUserRequest represents a user reactivation request (admin operation).
type ReactivateUserRequest struct {
	// No fields needed - UserID comes from URL path parameter
	// Admin identity comes from JWT claims
}

// ResetPasswordRequest represents a password reset request.
// Used after user clicks reset link in email.
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
	IPAddress   string `json:"-"` // Set from request context, not JSON body
}

// DeleteUserRequest represents a user deletion request (admin operation).
type DeleteUserRequest struct {
	Reason string `json:"reason" validate:"required"`
}

// ChangeRoleRequest represents a role change request (admin operation).
type ChangeRoleRequest struct {
	Role   string `json:"role" validate:"required,oneof=guest user moderator admin super_admin"`
	Reason string `json:"reason"` // Optional reason for audit trail
}

// SuspendUserRequest represents a user suspension request (admin operation).
type SuspendUserRequest struct {
	Reason string `json:"reason" validate:"required"`
}

// ListUsersRequest represents a query to list users with filters.
type ListUsersRequest struct {
	// Filters
	Status        *string `json:"status,omitempty"`
	Role          *string `json:"role,omitempty"`
	EmailVerified *bool   `json:"email_verified,omitempty"`
	EmailContains string  `json:"email_contains,omitempty"`

	// Pagination
	Limit  int `json:"limit" validate:"min=1,max=100"`
	Offset int `json:"offset" validate:"min=0"`

	// Sorting
	SortBy    string `json:"sort_by"`    // "created_at", "updated_at", "email", "last_login_at"
	SortOrder string `json:"sort_order"` // "asc", "desc"
}

// RefreshTokenRequest represents a token refresh request.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
