package jwt

import "time"

// Config holds JWT configuration.
type Config struct {
	// Secret is the key used to sign JWTs (for HMAC algorithms).
	Secret string `env:"SECRET"`

	// Issuer is the JWT issuer claim.
	Issuer string `env:"ISSUER"`

	// Audience is the JWT audience claim.
	Audience []string `env:"AUDIENCE"`

	// AccessTokenTTL is the lifetime of access tokens.
	AccessTokenTTL time.Duration `env:"ACCESS_TOKEN_TTL"`

	// RefreshTokenTTL is the lifetime of refresh tokens.
	RefreshTokenTTL time.Duration `env:"REFRESH_TOKEN_TTL"`

	// Algorithm is the signing algorithm (HS256, HS384, HS512).
	Algorithm string `env:"ALGORITHM"`
}

// DefaultConfig returns default JWT configuration.
func DefaultConfig() Config {
	return Config{
		Secret:          "change-me-in-production-use-at-least-32-chars",
		Issuer:          "hexagonal-go",
		Audience:        []string{"hexagonal-go"},
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour, // 7 days
		Algorithm:       "HS256",
	}
}
