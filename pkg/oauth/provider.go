package oauth

import (
	"context"
)

// Provider defines the interface for OAuth providers (Google, GitHub, etc.)
type Provider interface {
	// Name returns the provider name (e.g., "google", "github")
	Name() string

	// AuthCodeURL generates the OAuth authorization URL
	// State token is used for CSRF protection
	AuthCodeURL(state string) string

	// Exchange exchanges an authorization code for tokens and user info
	Exchange(ctx context.Context, code string) (*UserInfo, *Tokens, error)

	// RefreshToken refreshes an access token using a refresh token
	RefreshToken(ctx context.Context, refreshToken string) (*Tokens, error)
}

// UserInfo represents user information from an OAuth provider
type UserInfo struct {
	ID            string                 // Provider's user ID
	Email         string                 // User's email
	EmailVerified bool                   // Whether email is verified
	Name          string                 // Full name
	GivenName     string                 // First name
	FamilyName    string                 // Last name
	Picture       string                 // Profile picture URL
	Locale        string                 // Locale/language
	Raw           map[string]interface{} // Raw profile data
}

// Tokens represents OAuth tokens
type Tokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64 // Unix timestamp
	TokenType    string
}
