package jwt

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// service is the default JWT service implementation.
type service struct {
	config    Config
	blacklist map[string]time.Time
	mu        sync.RWMutex
}

// NewService creates a new JWT service.
func NewService(config Config) Service {
	return &service{
		config:    config,
		blacklist: make(map[string]time.Time),
	}
}

// Generate creates a new JWT from claims.
func (s *service) Generate(ctx context.Context, claims Claims) (string, error) {
	now := time.Now().UTC()

	// Build JWT claims
	jwtClaims := jwt.MapClaims{}

	// Copy all custom claims
	for k, v := range claims {
		jwtClaims[k] = v
	}

	// Set standard claims if not already set
	if _, ok := jwtClaims[ClaimIssuer]; !ok {
		jwtClaims[ClaimIssuer] = s.config.Issuer
	}
	if _, ok := jwtClaims[ClaimAudience]; !ok {
		jwtClaims[ClaimAudience] = s.config.Audience
	}
	if _, ok := jwtClaims[ClaimIssuedAt]; !ok {
		jwtClaims[ClaimIssuedAt] = now.Unix()
	}
	if _, ok := jwtClaims[ClaimNotBefore]; !ok {
		jwtClaims[ClaimNotBefore] = now.Unix()
	}
	if _, ok := jwtClaims[ClaimExpiresAt]; !ok {
		jwtClaims[ClaimExpiresAt] = now.Add(s.config.AccessTokenTTL).Unix()
	}

	// Create and sign token
	token := jwt.NewWithClaims(s.getSigningMethod(), jwtClaims)
	tokenString, err := token.SignedString([]byte(s.config.Secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// Validate parses and validates a JWT, returning the claims.
func (s *service) Validate(ctx context.Context, tokenString string) (Claims, error) {
	// Check blacklist first
	if blacklisted, _ := s.IsInvalidated(ctx, tokenString); blacklisted {
		return nil, ErrTokenBlacklisted
	}

	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if token.Method.Alg() != s.config.Algorithm {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.Secret), nil
	})

	if err != nil {
		return nil, s.mapError(err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// Extract claims
	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidClaims
	}

	// Convert to our Claims type
	claims := make(Claims)
	for k, v := range mapClaims {
		claims[k] = v
	}

	return claims, nil
}

// Invalidate adds a token to the blacklist.
func (s *service) Invalidate(ctx context.Context, tokenString string) error {
	// Parse to get expiry
	token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.Secret), nil
	})

	var expiry time.Time
	if token != nil {
		if mapClaims, ok := token.Claims.(jwt.MapClaims); ok {
			if exp, ok := mapClaims[ClaimExpiresAt].(float64); ok {
				expiry = time.Unix(int64(exp), 0)
			}
		}
	}

	// Default expiry if we couldn't parse
	if expiry.IsZero() {
		expiry = time.Now().Add(s.config.AccessTokenTTL)
	}

	s.mu.Lock()
	s.blacklist[tokenString] = expiry
	s.mu.Unlock()

	// Clean up expired entries
	go s.cleanupBlacklist()

	return nil
}

// IsInvalidated checks if a token has been invalidated.
func (s *service) IsInvalidated(ctx context.Context, tokenString string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.blacklist[tokenString]
	return exists, nil
}

// getSigningMethod returns the JWT signing method based on config.
func (s *service) getSigningMethod() jwt.SigningMethod {
	switch s.config.Algorithm {
	case "HS384":
		return jwt.SigningMethodHS384
	case "HS512":
		return jwt.SigningMethodHS512
	default:
		return jwt.SigningMethodHS256
	}
}

// mapError converts jwt-go errors to our error types.
func (s *service) mapError(err error) error {
	switch err {
	case jwt.ErrTokenExpired:
		return ErrExpiredToken
	case jwt.ErrTokenNotValidYet:
		return ErrTokenNotYetValid
	case jwt.ErrTokenSignatureInvalid:
		return ErrInvalidSignature
	default:
		return ErrInvalidToken
	}
}

// cleanupBlacklist removes expired tokens from the blacklist.
func (s *service) cleanupBlacklist() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for token, expiry := range s.blacklist {
		if now.After(expiry) {
			delete(s.blacklist, token)
		}
	}
}
