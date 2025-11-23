package user

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"

	pkgtypes "github.com/0xsj/hexagonal-go/pkg/types"
)

// Email represents a user's email address in the identity domain.
// Wraps pkg/types.Email with domain-specific behavior.
//
// Design:
//   - Immutable once created
//   - Always normalized (lowercase, trimmed)
//   - Validates format on creation
//   - Tracks verification state
type Email struct {
	address pkgtypes.Email
}

// NewEmail creates a new email from a string.
// Validates format and normalizes the address.
//
// Example:
//
//	email, err := NewEmail("user@example.com")
//	if err != nil {
//	    return ErrEmailInvalid(op, raw)
//	}
func NewEmail(address string) (Email, error) {
	// Normalize: trim whitespace and lowercase
	address = strings.TrimSpace(strings.ToLower(address))

	// Validate using pkg/types
	pkgEmail, err := pkgtypes.NewEmail(address)
	if err != nil {
		return Email{}, fmt.Errorf("invalid email address: %w", err)
	}

	return Email{address: pkgEmail}, nil
}

// MustNewEmail creates a new email and panics if invalid.
// Only use for constants where you're certain the value is valid.
func MustNewEmail(address string) Email {
	email, err := NewEmail(address)
	if err != nil {
		panic(fmt.Sprintf("invalid email: %v", err))
	}
	return email
}

// FromEmail creates a domain email from pkg/types.Email.
// Useful when you already have a validated email from pkg/types.
func FromEmail(email pkgtypes.Email) Email {
	return Email{address: email}
}

// ============================================================================
// Email Methods
// ============================================================================

// String returns the email address as a string.
func (e Email) String() string {
	return e.address.String()
}

// Address returns the underlying pkg/types.Email.
func (e Email) Address() pkgtypes.Email {
	return e.address
}

// Equals checks if two emails are equal.
// Comparison is case-insensitive and normalized.
func (e Email) Equals(other Email) bool {
	return e.String() == other.String()
}

// Domain returns the domain part of the email (everything after @).
//
// Example:
//
//	email := MustNewEmail("user@example.com")
//	domain := email.Domain() // "example.com"
func (e Email) Domain() string {
	parts := strings.Split(e.String(), "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

// LocalPart returns the local part of the email (everything before @).
//
// Example:
//
//	email := MustNewEmail("user@example.com")
//	local := email.LocalPart() // "user"
func (e Email) LocalPart() string {
	parts := strings.Split(e.String(), "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}

// IsEmpty returns true if the email is empty.
func (e Email) IsEmpty() bool {
	return e.String() == ""
}

// ============================================================================
// Domain-Specific Validation
// ============================================================================

// IsDisposable checks if the email is from a disposable email provider.
// Returns true for known disposable domains.
func (e Email) IsDisposable() bool {
	domain := e.Domain()

	// List of common disposable email domains
	// In production, use a comprehensive list or external service
	disposableDomains := []string{
		"tempmail.com",
		"throwaway.email",
		"guerrillamail.com",
		"10minutemail.com",
		"mailinator.com",
		"maildrop.cc",
		"sharklasers.com",
		"yopmail.com",
		"fakeinbox.com",
		"trashmail.com",
	}

	for _, disposable := range disposableDomains {
		if domain == disposable {
			return true
		}
	}

	return false
}

// IsCorporate checks if the email is from a corporate domain.
// Returns false for free email providers (gmail, yahoo, etc.).
func (e Email) IsCorporate() bool {
	domain := e.Domain()

	// List of common free email providers
	freeProviders := []string{
		"gmail.com",
		"yahoo.com",
		"hotmail.com",
		"outlook.com",
		"live.com",
		"icloud.com",
		"aol.com",
		"protonmail.com",
		"mail.com",
		"zoho.com",
	}

	for _, free := range freeProviders {
		if domain == free {
			return false
		}
	}

	// If not a free provider and not disposable, assume corporate
	return !e.IsDisposable()
}

// MatchesDomain checks if the email belongs to a specific domain.
//
// Example:
//
//	if !email.MatchesDomain("acme-corp.com") {
//	    return ErrEmailDomainMismatch(op, email.Domain())
//	}
func (e Email) MatchesDomain(domain string) bool {
	return strings.EqualFold(e.Domain(), domain)
}

// ============================================================================
// Obfuscation (for Privacy)
// ============================================================================

// Obfuscate returns a partially hidden email for display purposes.
// Example: "j***n@example.com" or "user@ex*****.com"
//
// Example:
//
//	email := MustNewEmail("john.doe@example.com")
//	obfuscated := email.Obfuscate() // "j******e@example.com"
func (e Email) Obfuscate() string {
	parts := strings.Split(e.String(), "@")
	if len(parts) != 2 {
		return "***@***"
	}

	local := parts[0]
	domain := parts[1]

	// Obfuscate local part
	var obfuscatedLocal string
	if len(local) <= 2 {
		obfuscatedLocal = "***"
	} else {
		obfuscatedLocal = string(local[0]) + strings.Repeat("*", len(local)-2) + string(local[len(local)-1])
	}

	return obfuscatedLocal + "@" + domain
}

// ObfuscateFull returns a fully obfuscated email for high privacy.
// Example: "j***@ex*****.com"
func (e Email) ObfuscateFull() string {
	parts := strings.Split(e.String(), "@")
	if len(parts) != 2 {
		return "***@***"
	}

	local := parts[0]
	domain := parts[1]

	// Obfuscate local
	obfuscatedLocal := string(local[0]) + "***"

	// Obfuscate domain
	domainParts := strings.Split(domain, ".")
	if len(domainParts) >= 2 {
		domainName := domainParts[0]
		extension := domainParts[len(domainParts)-1]
		obfuscatedDomain := string(domainName[0]) + string(domainName[1]) + "****." + extension
		return obfuscatedLocal + "@" + obfuscatedDomain
	}

	return obfuscatedLocal + "@***"
}

// ============================================================================
// Database Marshaling
// ============================================================================

// Scan implements sql.Scanner for reading from database.
func (e *Email) Scan(value interface{}) error {
	if value == nil {
		*e = Email{}
		return nil
	}

	var str string
	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("cannot scan %T into Email", value)
	}

	if str == "" {
		*e = Email{}
		return nil
	}

	email, err := NewEmail(str)
	if err != nil {
		return err
	}

	*e = email
	return nil
}

// Value implements driver.Valuer for writing to database.
func (e Email) Value() (driver.Value, error) {
	if e.IsEmpty() {
		return nil, nil
	}
	return e.String(), nil
}

// ============================================================================
// JSON Marshaling
// ============================================================================

// MarshalJSON implements json.Marshaler.
func (e Email) MarshalJSON() ([]byte, error) {
	if e.IsEmpty() {
		return []byte("null"), nil
	}
	return json.Marshal(e.String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (e *Email) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	if str == "" {
		*e = Email{}
		return nil
	}

	email, err := NewEmail(str)
	if err != nil {
		return err
	}

	*e = email
	return nil
}

// ============================================================================
// Validation Helpers
// ============================================================================

// ValidateForRegistration checks if email is valid for user registration.
// Rejects disposable emails and enforces additional rules.
func (e Email) ValidateForRegistration() error {
	if e.IsEmpty() {
		return fmt.Errorf("email address is required")
	}

	if e.IsDisposable() {
		return fmt.Errorf("disposable email addresses are not allowed")
	}

	// Add more rules as needed
	// - Check domain MX records
	// - Check against blacklist
	// - Enforce domain whitelist for enterprise

	return nil
}

// ValidateForTenant checks if email is valid for a specific tenant.
// Can enforce tenant-specific email domain requirements.
//
// Example:
//
//	// Enterprise tenant requires company email domain
//	if err := email.ValidateForTenant(tenant); err != nil {
//	    return ErrEmailDomainNotAllowed(op, email.Domain())
//	}
func (e Email) ValidateForTenant(tenantDomain string) error {
	if tenantDomain == "" {
		// No domain restriction
		return nil
	}

	if !e.MatchesDomain(tenantDomain) {
		return fmt.Errorf("email must be from domain: %s", tenantDomain)
	}

	return nil
}

// ============================================================================
// Comparison & Sorting
// ============================================================================

// Less returns true if this email is lexicographically less than other.
// Useful for sorting.
func (e Email) Less(other Email) bool {
	return e.String() < other.String()
}
