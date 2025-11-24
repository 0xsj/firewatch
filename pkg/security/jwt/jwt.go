package jwt

import (
	"context"
	"time"
)

// Claims represents JWT claims as a generic map.
// The JWT package doesn't know about domain-specific claim structures.
type Claims map[string]any

// Standard claim keys.
const (
	ClaimSubject   = "sub"
	ClaimIssuer    = "iss"
	ClaimAudience  = "aud"
	ClaimExpiresAt = "exp"
	ClaimIssuedAt  = "iat"
	ClaimNotBefore = "nbf"
	ClaimJTI       = "jti"
)

// Get retrieves a claim value.
func (c Claims) Get(key string) (any, bool) {
	v, ok := c[key]
	return v, ok
}

// GetString retrieves a string claim.
func (c Claims) GetString(key string) string {
	if v, ok := c[key].(string); ok {
		return v
	}
	return ""
}

// GetTime retrieves a time claim.
func (c Claims) GetTime(key string) time.Time {
	switch v := c[key].(type) {
	case time.Time:
		return v
	case float64:
		return time.Unix(int64(v), 0)
	case int64:
		return time.Unix(v, 0)
	}
	return time.Time{}
}

// Subject returns the subject claim.
func (c Claims) Subject() string {
	return c.GetString(ClaimSubject)
}

// ExpiresAt returns the expiration time.
func (c Claims) ExpiresAt() time.Time {
	return c.GetTime(ClaimExpiresAt)
}

// IsExpired returns true if the token has expired.
func (c Claims) IsExpired() bool {
	exp := c.ExpiresAt()
	if exp.IsZero() {
		return false
	}
	return time.Now().UTC().After(exp)
}

// Service is the port for JWT operations.
type Service interface {
	// Generate creates a new JWT from claims.
	Generate(ctx context.Context, claims Claims) (string, error)

	// Validate parses and validates a JWT, returning the claims.
	Validate(ctx context.Context, token string) (Claims, error)

	// Invalidate adds a token to the blacklist (for logout).
	Invalidate(ctx context.Context, token string) error

	// IsInvalidated checks if a token has been invalidated.
	IsInvalidated(ctx context.Context, token string) (bool, error)
}

// Errors
var (
	ErrInvalidToken     = &Error{Code: "invalid_token", Message: "token is invalid"}
	ErrExpiredToken     = &Error{Code: "expired_token", Message: "token has expired"}
	ErrInvalidSignature = &Error{Code: "invalid_signature", Message: "token signature is invalid"}
	ErrInvalidClaims    = &Error{Code: "invalid_claims", Message: "token claims are invalid"}
	ErrTokenNotYetValid = &Error{Code: "token_not_yet_valid", Message: "token is not yet valid"}
	ErrTokenBlacklisted = &Error{Code: "token_blacklisted", Message: "token has been revoked"}
)

// Error represents a JWT-specific error.
type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string {
	return e.Message
}
