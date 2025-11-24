package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// TokenType represents the type of token.
type TokenType string

const (
	TokenTypeAccess    TokenType = "access"
	TokenTypeRefresh   TokenType = "refresh"
	TokenTypeMagicLink TokenType = "magic_link"
	TokenTypeVerify    TokenType = "verify_email"
	TokenTypeReset     TokenType = "password_reset"
)

// String returns the string representation.
func (t TokenType) String() string {
	return string(t)
}

// Token represents an authentication token.
type Token struct {
	id        types.ID
	tokenType TokenType
	value     string
	userID    types.ID
	tenantID  string
	expiresAt time.Time
	createdAt time.Time
	usedAt    *time.Time
	revokedAt *time.Time
}

// NewToken creates a new token.
func NewToken(tokenType TokenType, userID types.ID, tenantID string, ttl time.Duration) (*Token, error) {
	value, err := generateSecureToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	now := time.Now().UTC()
	return &Token{
		id:        types.NewID(),
		tokenType: tokenType,
		value:     value,
		userID:    userID,
		tenantID:  tenantID,
		expiresAt: now.Add(ttl),
		createdAt: now,
	}, nil
}

// NewTokenWithValue creates a token with a specific value (for reconstitution).
func NewTokenWithValue(
	id types.ID,
	tokenType TokenType,
	value string,
	userID types.ID,
	tenantID string,
	expiresAt time.Time,
	createdAt time.Time,
	usedAt *time.Time,
	revokedAt *time.Time,
) *Token {
	return &Token{
		id:        id,
		tokenType: tokenType,
		value:     value,
		userID:    userID,
		tenantID:  tenantID,
		expiresAt: expiresAt,
		createdAt: createdAt,
		usedAt:    usedAt,
		revokedAt: revokedAt,
	}
}

// Getters

func (t *Token) ID() types.ID          { return t.id }
func (t *Token) Type() TokenType       { return t.tokenType }
func (t *Token) Value() string         { return t.value }
func (t *Token) UserID() types.ID      { return t.userID }
func (t *Token) TenantID() string      { return t.tenantID }
func (t *Token) ExpiresAt() time.Time  { return t.expiresAt }
func (t *Token) CreatedAt() time.Time  { return t.createdAt }
func (t *Token) UsedAt() *time.Time    { return t.usedAt }
func (t *Token) RevokedAt() *time.Time { return t.revokedAt }

// IsExpired returns true if the token has expired.
func (t *Token) IsExpired() bool {
	return time.Now().UTC().After(t.expiresAt)
}

// IsUsed returns true if the token has been used.
func (t *Token) IsUsed() bool {
	return t.usedAt != nil
}

// IsRevoked returns true if the token has been revoked.
func (t *Token) IsRevoked() bool {
	return t.revokedAt != nil
}

// IsValid returns true if the token is valid (not expired, used, or revoked).
func (t *Token) IsValid() bool {
	return !t.IsExpired() && !t.IsUsed() && !t.IsRevoked()
}

// MarkUsed marks the token as used.
func (t *Token) MarkUsed() error {
	if t.IsUsed() {
		return fmt.Errorf("token already used")
	}
	if t.IsRevoked() {
		return fmt.Errorf("token is revoked")
	}
	if t.IsExpired() {
		return fmt.Errorf("token is expired")
	}

	now := time.Now().UTC()
	t.usedAt = &now
	return nil
}

// Revoke revokes the token.
func (t *Token) Revoke() error {
	if t.IsRevoked() {
		return fmt.Errorf("token already revoked")
	}

	now := time.Now().UTC()
	t.revokedAt = &now
	return nil
}

// TimeToExpiry returns the duration until the token expires.
func (t *Token) TimeToExpiry() time.Duration {
	return time.Until(t.expiresAt)
}

// Snapshot for persistence.
type TokenSnapshot struct {
	ID        string
	TokenType string
	Value     string
	UserID    string
	TenantID  string
	ExpiresAt time.Time
	CreatedAt time.Time
	UsedAt    *time.Time
	RevokedAt *time.Time
}

// ToSnapshot converts to a snapshot for persistence.
func (t *Token) ToSnapshot() TokenSnapshot {
	return TokenSnapshot{
		ID:        t.id.String(),
		TokenType: t.tokenType.String(),
		Value:     t.value,
		UserID:    t.userID.String(),
		TenantID:  t.tenantID,
		ExpiresAt: t.expiresAt,
		CreatedAt: t.createdAt,
		UsedAt:    t.usedAt,
		RevokedAt: t.revokedAt,
	}
}

// TokenFromSnapshot reconstitutes from a snapshot.
func TokenFromSnapshot(s TokenSnapshot) (*Token, error) {
	id, err := types.ParseID(s.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid token id: %w", err)
	}

	userID, err := types.ParseID(s.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	return &Token{
		id:        id,
		tokenType: TokenType(s.TokenType),
		value:     s.Value,
		userID:    userID,
		tenantID:  s.TenantID,
		expiresAt: s.ExpiresAt,
		createdAt: s.CreatedAt,
		usedAt:    s.UsedAt,
		revokedAt: s.RevokedAt,
	}, nil
}

// generateSecureToken generates a cryptographically secure random token.
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// Default TTLs for different token types.
const (
	AccessTokenTTL   = 15 * time.Minute
	RefreshTokenTTL  = 7 * 24 * time.Hour // 7 days
	MagicLinkTTL     = 15 * time.Minute
	VerifyEmailTTL   = 24 * time.Hour // 24 hours
	PasswordResetTTL = 1 * time.Hour
)
