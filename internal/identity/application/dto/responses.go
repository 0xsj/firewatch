package dto

// LoginResponse represents the response after successful login.
type LoginResponse struct {
	User         *UserDTO `json:"user"`
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token,omitempty"`
	ExpiresIn    int      `json:"expires_in"` // Token expiration in seconds
	TokenType    string   `json:"token_type"` // "Bearer"
}

// ListUsersResponse represents a paginated list of users.
type ListUsersResponse struct {
	Users      []*UserSummaryDTO `json:"users"`
	TotalCount int               `json:"total_count"`
	Limit      int               `json:"limit"`
	Offset     int               `json:"offset"`
	HasMore    bool              `json:"has_more"` // True if more results available
}

// MessageResponse represents a simple success message.
// Used for operations that don't return data (e.g., email verification).
type MessageResponse struct {
	Message string `json:"message"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Code    string                 `json:"code,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ValidationErrorResponse represents validation errors.
type ValidationErrorResponse struct {
	Error  string            `json:"error"`
	Code   string            `json:"code"`
	Fields map[string]string `json:"fields"` // field -> error message
}
