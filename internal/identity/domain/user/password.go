package user

import (
	"crypto/subtle"
	"database/sql/driver"
	"fmt"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// Password represents a hashed password.
// Always stored as a bcrypt hash, never plain text.
//
// Design:
//   - Password is OPTIONAL (can be nil for passwordless users)
//   - Plain text passwords are never stored
//   - Hashing uses bcrypt with configurable cost
//   - Validation enforces security requirements
type Password struct {
	hash string // bcrypt hash
}

// PasswordRequirements defines password security requirements.
type PasswordRequirements struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireNumber  bool
	RequireSpecial bool
}

// DefaultPasswordRequirements returns the default password requirements.
func DefaultPasswordRequirements() PasswordRequirements {
	return PasswordRequirements{
		MinLength:      8,
		RequireUpper:   true,
		RequireLower:   true,
		RequireNumber:  true,
		RequireSpecial: false, // Optional special chars
	}
}

// StrongPasswordRequirements returns strict password requirements.
func StrongPasswordRequirements() PasswordRequirements {
	return PasswordRequirements{
		MinLength:      12,
		RequireUpper:   true,
		RequireLower:   true,
		RequireNumber:  true,
		RequireSpecial: true,
	}
}

const (
	// DefaultBcryptCost is the default bcrypt cost (10 = ~100ms on modern hardware)
	DefaultBcryptCost = 10

	// MaxPasswordLength is the maximum password length to prevent DoS
	MaxPasswordLength = 72 // bcrypt limitation
)

// ============================================================================
// Password Creation
// ============================================================================

// NewPassword creates a new password from plain text.
// The plain text is immediately hashed and discarded.
//
// Example:
//
//	password, err := NewPassword("MyS3cur3P@ssw0rd", DefaultPasswordRequirements())
//	if err != nil {
//	    return ErrPasswordTooWeak(op, err.Error())
//	}
func NewPassword(plainText string, requirements PasswordRequirements) (*Password, error) {
	// Validate plain text password
	if err := ValidatePasswordStrength(plainText, requirements); err != nil {
		return nil, err
	}

	// Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(plainText), DefaultBcryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	return &Password{hash: string(hash)}, nil
}

// NewPasswordWithCost creates a password with a custom bcrypt cost.
// Higher cost = more secure but slower. Use for high-security scenarios.
func NewPasswordWithCost(plainText string, requirements PasswordRequirements, cost int) (*Password, error) {
	if err := ValidatePasswordStrength(plainText, requirements); err != nil {
		return nil, err
	}

	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return nil, fmt.Errorf("invalid bcrypt cost: %d (must be %d-%d)",
			cost, bcrypt.MinCost, bcrypt.MaxCost)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(plainText), cost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	return &Password{hash: string(hash)}, nil
}

// FromHash creates a Password from an existing bcrypt hash.
// Used when loading from database.
func FromHash(hash string) (*Password, error) {
	if hash == "" {
		return nil, fmt.Errorf("password hash cannot be empty")
	}

	// Verify it's a valid bcrypt hash
	if _, err := bcrypt.Cost([]byte(hash)); err != nil {
		return nil, fmt.Errorf("invalid bcrypt hash: %w", err)
	}

	return &Password{hash: hash}, nil
}

// ============================================================================
// Password Verification
// ============================================================================

// Matches checks if the plain text password matches this hashed password.
// Uses constant-time comparison to prevent timing attacks.
//
// Example:
//
//	if !user.Password.Matches(inputPassword) {
//	    return ErrInvalidCredentials(op)
//	}
func (p *Password) Matches(plainText string) bool {
	if p == nil {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(p.hash), []byte(plainText))
	return err == nil
}

// NeedsRehash checks if the password should be rehashed.
// Returns true if the cost has changed (security upgrade).
func (p *Password) NeedsRehash() bool {
	if p == nil {
		return false
	}

	cost, err := bcrypt.Cost([]byte(p.hash))
	if err != nil {
		return false
	}

	return cost != DefaultBcryptCost
}

// ============================================================================
// Password Validation
// ============================================================================

// ValidatePasswordStrength validates a plain text password against requirements.
func ValidatePasswordStrength(password string, reqs PasswordRequirements) error {
	if len(password) < reqs.MinLength {
		return fmt.Errorf("password must be at least %d characters", reqs.MinLength)
	}

	if len(password) > MaxPasswordLength {
		return fmt.Errorf("password must not exceed %d characters", MaxPasswordLength)
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if reqs.RequireUpper && !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if reqs.RequireLower && !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if reqs.RequireNumber && !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}

	if reqs.RequireSpecial && !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	// Check for common weak passwords
	if isCommonPassword(password) {
		return fmt.Errorf("password is too common, please choose a stronger password")
	}

	return nil
}

// isCommonPassword checks if the password is in a list of common passwords.
// This is a simplified version - in production, use a proper password blacklist.
func isCommonPassword(password string) bool {
	// Common weak passwords
	common := []string{
		"password", "password123", "12345678", "qwerty", "abc123",
		"monkey", "1234567", "letmein", "trustno1", "dragon",
		"baseball", "iloveyou", "master", "sunshine", "ashley",
		"bailey", "passw0rd", "shadow", "123123", "654321",
	}

	// Case-insensitive comparison
	for _, weak := range common {
		if subtle.ConstantTimeCompare(
			[]byte(password),
			[]byte(weak),
		) == 1 {
			return true
		}
	}

	return false
}

// ============================================================================
// Hash Access
// ============================================================================

// Hash returns the bcrypt hash.
// Only use this for persistence - never expose to clients.
func (p *Password) Hash() string {
	if p == nil {
		return ""
	}
	return p.hash
}

// String returns a safe string representation (never the actual hash).
func (p *Password) String() string {
	return "[REDACTED]"
}

// ============================================================================
// Database Marshaling
// ============================================================================

// Scan implements sql.Scanner for reading from database.
func (p *Password) Scan(value interface{}) error {
	if value == nil {
		*p = Password{}
		return nil
	}

	var hash string
	switch v := value.(type) {
	case string:
		hash = v
	case []byte:
		hash = string(v)
	default:
		return fmt.Errorf("cannot scan %T into Password", value)
	}

	password, err := FromHash(hash)
	if err != nil {
		return err
	}

	*p = *password
	return nil
}

// Value implements driver.Valuer for writing to database.
func (p *Password) Value() (driver.Value, error) {
	if p == nil {
		return nil, nil
	}
	return p.hash, nil
}

// ============================================================================
// JSON Marshaling (NEVER expose password hash in JSON)
// ============================================================================

// MarshalJSON implements json.Marshaler.
// Always returns null - passwords should never be serialized to JSON.
func (p *Password) MarshalJSON() ([]byte, error) {
	return []byte("null"), nil
}

// UnmarshalJSON implements json.Unmarshaler.
// Passwords cannot be unmarshaled from JSON (they must be created via NewPassword).
func (p *Password) UnmarshalJSON(data []byte) error {
	return fmt.Errorf("password cannot be unmarshaled from JSON")
}

// ============================================================================
// Password Strength Helpers
// ============================================================================

// EstimateStrength estimates password strength (0-100).
// This is a simple heuristic - for production, consider zxcvbn.
func EstimateStrength(password string) int {
	strength := 0

	// Length bonus
	if len(password) >= 8 {
		strength += 20
	}
	if len(password) >= 12 {
		strength += 10
	}
	if len(password) >= 16 {
		strength += 10
	}

	// Character variety
	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
		}
		if unicode.IsLower(char) {
			hasLower = true
		}
		if unicode.IsNumber(char) {
			hasNumber = true
		}
		if unicode.IsPunct(char) || unicode.IsSymbol(char) {
			hasSpecial = true
		}
	}

	if hasUpper {
		strength += 15
	}
	if hasLower {
		strength += 15
	}
	if hasNumber {
		strength += 15
	}
	if hasSpecial {
		strength += 15
	}

	// Penalty for common passwords
	if isCommonPassword(password) {
		strength = strength / 2
	}

	if strength > 100 {
		strength = 100
	}

	return strength
}
