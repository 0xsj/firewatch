package auth

import (
	"time"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Claims represents JWT claims for access tokens.
type Claims struct {
	// Standard claims
	Subject   string    `json:"sub"`           // User ID
	IssuedAt  time.Time `json:"iat"`           // Issued at
	ExpiresAt time.Time `json:"exp"`           // Expiration time
	NotBefore time.Time `json:"nbf,omitempty"` // Not valid before
	Issuer    string    `json:"iss,omitempty"` // Issuer
	Audience  []string  `json:"aud,omitempty"` // Audience

	// Custom claims
	TenantID  string   `json:"tenant_id"`        // Tenant ID
	Email     string   `json:"email"`            // User email
	Role      string   `json:"role"`             // User role
	Provider  string   `json:"provider"`         // Auth provider used
	SessionID string   `json:"session_id"`       // Session ID
	Scopes    []string `json:"scopes,omitempty"` // Granted scopes
}

// NewClaims creates new JWT claims.
func NewClaims(
	userID types.ID,
	tenantID string,
	email string,
	role string,
	provider Provider,
	sessionID types.ID,
	ttl time.Duration,
) *Claims {
	now := time.Now().UTC()
	return &Claims{
		Subject:   userID.String(),
		IssuedAt:  now,
		ExpiresAt: now.Add(ttl),
		NotBefore: now,
		TenantID:  tenantID,
		Email:     email,
		Role:      role,
		Provider:  provider.String(),
		SessionID: sessionID.String(),
	}
}

// WithIssuer sets the issuer.
func (c *Claims) WithIssuer(issuer string) *Claims {
	c.Issuer = issuer
	return c
}

// WithAudience sets the audience.
func (c *Claims) WithAudience(audience ...string) *Claims {
	c.Audience = audience
	return c
}

// WithScopes sets the scopes.
func (c *Claims) WithScopes(scopes ...string) *Claims {
	c.Scopes = scopes
	return c
}

// IsExpired returns true if the claims have expired.
func (c *Claims) IsExpired() bool {
	return time.Now().UTC().After(c.ExpiresAt)
}

// IsNotYetValid returns true if the claims are not yet valid.
func (c *Claims) IsNotYetValid() bool {
	return time.Now().UTC().Before(c.NotBefore)
}

// IsValid returns true if the claims are currently valid.
func (c *Claims) IsValid() bool {
	return !c.IsExpired() && !c.IsNotYetValid()
}

// UserID returns the user ID from the subject.
func (c *Claims) UserID() (types.ID, error) {
	return types.ParseID(c.Subject)
}

// GetSessionID returns the session ID.
func (c *Claims) GetSessionID() (types.ID, error) {
	return types.ParseID(c.SessionID)
}

// TimeToExpiry returns the duration until expiration.
func (c *Claims) TimeToExpiry() time.Duration {
	return time.Until(c.ExpiresAt)
}

// HasScope returns true if the claims include the given scope.
func (c *Claims) HasScope(scope string) bool {
	for _, s := range c.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// HasAnyScope returns true if the claims include any of the given scopes.
func (c *Claims) HasAnyScope(scopes ...string) bool {
	for _, scope := range scopes {
		if c.HasScope(scope) {
			return true
		}
	}
	return false
}

// HasAllScopes returns true if the claims include all of the given scopes.
func (c *Claims) HasAllScopes(scopes ...string) bool {
	for _, scope := range scopes {
		if !c.HasScope(scope) {
			return false
		}
	}
	return true
}
