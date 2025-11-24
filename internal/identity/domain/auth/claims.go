package auth

import (
	"github.com/0xsj/hexagonal-go/pkg/security/jwt"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Claim keys for identity domain.
const (
	ClaimUserID    = "user_id"
	ClaimTenantID  = "tenant_id"
	ClaimEmail     = "email"
	ClaimRole      = "role"
	ClaimProvider  = "provider"
	ClaimSessionID = "session_id"
	ClaimScopes    = "scopes"
)

// Claims represents identity-specific JWT claims.
type Claims struct {
	UserID    string
	TenantID  string
	Email     string
	Role      string
	Provider  string
	SessionID string
	Scopes    []string
}

// NewClaims creates new identity claims.
func NewClaims(
	userID types.ID,
	tenantID string,
	email string,
	role string,
	provider Provider,
	sessionID types.ID,
) *Claims {
	return &Claims{
		UserID:    userID.String(),
		TenantID:  tenantID,
		Email:     email,
		Role:      role,
		Provider:  provider.String(),
		SessionID: sessionID.String(),
	}
}

// WithScopes sets the scopes.
func (c *Claims) WithScopes(scopes ...string) *Claims {
	c.Scopes = scopes
	return c
}

// ToJWT converts identity claims to generic JWT claims.
func (c *Claims) ToJWT() jwt.Claims {
	claims := jwt.Claims{
		jwt.ClaimSubject: c.UserID,
		ClaimUserID:      c.UserID,
		ClaimTenantID:    c.TenantID,
		ClaimEmail:       c.Email,
		ClaimRole:        c.Role,
		ClaimProvider:    c.Provider,
		ClaimSessionID:   c.SessionID,
	}

	if len(c.Scopes) > 0 {
		claims[ClaimScopes] = c.Scopes
	}

	return claims
}

// ClaimsFromJWT converts generic JWT claims to identity claims.
func ClaimsFromJWT(jwtClaims jwt.Claims) *Claims {
	claims := &Claims{
		UserID:    jwtClaims.GetString(ClaimUserID),
		TenantID:  jwtClaims.GetString(ClaimTenantID),
		Email:     jwtClaims.GetString(ClaimEmail),
		Role:      jwtClaims.GetString(ClaimRole),
		Provider:  jwtClaims.GetString(ClaimProvider),
		SessionID: jwtClaims.GetString(ClaimSessionID),
	}

	// Handle scopes
	if scopes, ok := jwtClaims[ClaimScopes].([]any); ok {
		for _, s := range scopes {
			if str, ok := s.(string); ok {
				claims.Scopes = append(claims.Scopes, str)
			}
		}
	}

	// Fallback to subject if user_id not set
	if claims.UserID == "" {
		claims.UserID = jwtClaims.Subject()
	}

	return claims
}

// GetUserID returns the user ID.
func (c *Claims) GetUserID() (types.ID, error) {
	return types.ParseID(c.UserID)
}

// GetSessionID returns the session ID.
func (c *Claims) GetSessionID() (types.ID, error) {
	return types.ParseID(c.SessionID)
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
