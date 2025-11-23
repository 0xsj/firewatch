package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/mail"
	"strings"
)

// Email represents a validated email address value object.
// It is immutable and always valid after construction.
//
// Zero value is invalid - use NewEmail() or ParseEmail() to create.
//
// Example:
//
//	email, err := types.ParseEmail("user@example.com")
//	if err != nil {
//	    return err
//	}
//	fmt.Println(email.String()) // "user@example.com"
//	fmt.Println(email.Domain()) // "example.com"
type Email struct {
	value string
}

// NewEmail creates an Email from a string with validation.
// Returns an error if the email format is invalid.
//
// The email is normalized:
//   - Converted to lowercase
//   - Trimmed of whitespace
//
// Example:
//
//	email, err := types.NewEmail("User@Example.COM")
//	// Result: "user@example.com"
func NewEmail(address string) (Email, error) {
	return ParseEmail(address)
}

// ParseEmail parses and validates an email address.
// Performs RFC 5322 validation and normalization.
//
// Returns an error if:
//   - Email is empty
//   - Email format is invalid
//   - Email is too long (>254 characters per RFC)
func ParseEmail(address string) (Email, error) {
	// Trim and normalize
	address = strings.TrimSpace(address)
	address = strings.ToLower(address)

	// Check empty
	if address == "" {
		return Email{}, fmt.Errorf("email cannot be empty")
	}

	// Check length (RFC 5321: max 254 characters)
	if len(address) > 254 {
		return Email{}, fmt.Errorf("email too long: max 254 characters")
	}

	// Validate format using net/mail (RFC 5322)
	parsed, err := mail.ParseAddress(address)
	if err != nil {
		return Email{}, fmt.Errorf("invalid email format: %w", err)
	}

	// Use the normalized address from mail.ParseAddress
	normalized := strings.ToLower(parsed.Address)

	// Additional validation: must contain @
	if !strings.Contains(normalized, "@") {
		return Email{}, fmt.Errorf("invalid email: missing @")
	}

	// Split and validate parts
	parts := strings.Split(normalized, "@")
	if len(parts) != 2 {
		return Email{}, fmt.Errorf("invalid email: multiple @ symbols")
	}

	local, domain := parts[0], parts[1]

	// Validate local part (before @)
	if local == "" {
		return Email{}, fmt.Errorf("invalid email: empty local part")
	}
	if len(local) > 64 {
		return Email{}, fmt.Errorf("invalid email: local part too long (max 64)")
	}

	// Validate domain part (after @)
	if domain == "" {
		return Email{}, fmt.Errorf("invalid email: empty domain")
	}
	if len(domain) > 253 {
		return Email{}, fmt.Errorf("invalid email: domain too long (max 253)")
	}
	if !strings.Contains(domain, ".") {
		return Email{}, fmt.Errorf("invalid email: domain must contain at least one dot")
	}

	return Email{value: normalized}, nil
}

// MustParseEmail parses an email and panics if invalid.
// Only use this for constants where you're certain the value is valid.
func MustParseEmail(address string) Email {
	email, err := ParseEmail(address)
	if err != nil {
		panic(fmt.Sprintf("invalid email: %v", err))
	}
	return email
}

// String returns the normalized email address.
func (e Email) String() string {
	return e.value
}

// IsZero returns true if the email is the zero value (invalid/uninitialized).
func (e Email) IsZero() bool {
	return e.value == ""
}

// IsValid returns true if the email is valid (non-zero).
func (e Email) IsValid() bool {
	return e.value != ""
}

// Equals checks if two emails are equal (case-insensitive).
func (e Email) Equals(other Email) bool {
	return e.value == other.value
}

// Local returns the local part of the email (before @).
//
// Example:
//
//	email := types.MustParseEmail("user@example.com")
//	email.Local() // "user"
func (e Email) Local() string {
	if e.IsZero() {
		return ""
	}
	parts := strings.Split(e.value, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}

// Domain returns the domain part of the email (after @).
//
// Example:
//
//	email := types.MustParseEmail("user@example.com")
//	email.Domain() // "example.com"
func (e Email) Domain() string {
	if e.IsZero() {
		return ""
	}
	parts := strings.Split(e.value, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

// Masked returns a privacy-safe masked version of the email.
// Shows first character of local part and full domain.
//
// Example:
//
//	email := types.MustParseEmail("john.doe@example.com")
//	email.Masked() // "j***@example.com"
func (e Email) Masked() string {
	if e.IsZero() {
		return ""
	}

	local := e.Local()
	domain := e.Domain()

	if len(local) == 0 {
		return "***@" + domain
	}

	// Show first character + ***
	masked := string(local[0]) + "***"
	return masked + "@" + domain
}

// MaskedFull returns a fully masked version showing only domain.
//
// Example:
//
//	email := types.MustParseEmail("user@example.com")
//	email.MaskedFull() // "***@example.com"
func (e Email) MaskedFull() string {
	if e.IsZero() {
		return ""
	}
	return "***@" + e.Domain()
}

// ============================================================================
// JSON Marshaling
// ============================================================================

// MarshalJSON implements json.Marshaler.
// Encodes email as a JSON string.
func (e Email) MarshalJSON() ([]byte, error) {
	if e.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(e.value)
}

// UnmarshalJSON implements json.Unmarshaler.
// Decodes and validates email from JSON string.
func (e *Email) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	if s == "" || s == "null" {
		*e = Email{}
		return nil
	}

	parsed, err := ParseEmail(s)
	if err != nil {
		return err
	}

	*e = parsed
	return nil
}

// ============================================================================
// SQL Scanning
// ============================================================================

// Scan implements sql.Scanner.
// Allows reading email from database.
func (e *Email) Scan(value interface{}) error {
	if value == nil {
		*e = Email{}
		return nil
	}

	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("cannot scan %T into Email", value)
	}

	if s == "" {
		*e = Email{}
		return nil
	}

	parsed, err := ParseEmail(s)
	if err != nil {
		return err
	}

	*e = parsed
	return nil
}

// Value implements driver.Valuer.
// Allows writing email to database.
func (e Email) Value() (driver.Value, error) {
	if e.IsZero() {
		return nil, nil
	}
	return e.value, nil
}

// ============================================================================
// Text Marshaling (YAML, TOML, etc.)
// ============================================================================

// MarshalText implements encoding.TextMarshaler.
func (e Email) MarshalText() ([]byte, error) {
	if e.IsZero() {
		return []byte{}, nil
	}
	return []byte(e.value), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (e *Email) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*e = Email{}
		return nil
	}

	parsed, err := ParseEmail(string(text))
	if err != nil {
		return err
	}

	*e = parsed
	return nil
}

// ============================================================================
// Helpers
// ============================================================================

// GoString implements fmt.GoStringer for better debugging.
func (e Email) GoString() string {
	if e.IsZero() {
		return "Email{zero}"
	}
	return fmt.Sprintf("Email{%s}", e.value)
}

// Compare compares two emails lexicographically.
// Returns:
//   - -1 if e < other
//   - 0 if e == other
//   - +1 if e > other
func (e Email) Compare(other Email) int {
	return strings.Compare(e.value, other.value)
}

// IsDomain checks if the email is from a specific domain.
//
// Example:
//
//	email.IsDomain("example.com") // true for user@example.com
//	email.IsDomain("gmail.com")   // false
func (e Email) IsDomain(domain string) bool {
	return strings.EqualFold(e.Domain(), strings.ToLower(domain))
}

// IsSubdomain checks if the email is from a subdomain of the given domain.
//
// Example:
//
//	email := types.MustParseEmail("user@mail.example.com")
//	email.IsSubdomain("example.com") // true
func (e Email) IsSubdomain(parentDomain string) bool {
	domain := e.Domain()
	parentDomain = strings.ToLower(parentDomain)
	
	// Exact match
	if domain == parentDomain {
		return true
	}
	
	// Subdomain match (ends with .parentDomain)
	return strings.HasSuffix(domain, "."+parentDomain)
}