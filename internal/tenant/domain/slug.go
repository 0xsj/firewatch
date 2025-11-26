package tenant

import (
	"regexp"
	"strings"
)

// Slug is a URL-friendly identifier for a tenant.
// Format: lowercase alphanumeric with hyphens, 3-63 characters.
// Examples: "acme-corp", "johns-startup", "dev-team-alpha"
type Slug struct {
	value string
}

const (
	slugMinLength = 3
	slugMaxLength = 63
)

// slugPattern validates slug format: lowercase alphanumeric with hyphens,
// must start and end with alphanumeric.
var slugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$|^[a-z0-9]$`)

// Default reserved slugs that cannot be used by tenants.
// Can be extended via configuration.
var defaultReservedSlugs = map[string]bool{
	"admin":     true,
	"api":       true,
	"app":       true,
	"auth":      true,
	"billing":   true,
	"blog":      true,
	"cdn":       true,
	"dashboard": true,
	"docs":      true,
	"help":      true,
	"login":     true,
	"logout":    true,
	"mail":      true,
	"manage":    true,
	"portal":    true,
	"register":  true,
	"settings":  true,
	"signup":    true,
	"static":    true,
	"status":    true,
	"support":   true,
	"system":    true,
	"web":       true,
	"www":       true,
}

// NewSlug creates a new Slug from a string.
// Returns an error if the slug is invalid.
func NewSlug(value string) (Slug, error) {
	return NewSlugWithReserved(value, nil)
}

// NewSlugWithReserved creates a new Slug with a custom reserved list.
// If reservedSlugs is nil, uses the default reserved list.
func NewSlugWithReserved(value string, reservedSlugs map[string]bool) (Slug, error) {
	const op = "Slug.New"

	// Normalize: lowercase and trim
	normalized := strings.ToLower(strings.TrimSpace(value))

	// Check length
	if len(normalized) < slugMinLength {
		return Slug{}, ErrSlugInvalid(op, value, "must be at least 3 characters")
	}
	if len(normalized) > slugMaxLength {
		return Slug{}, ErrSlugInvalid(op, value, "must be at most 63 characters")
	}

	// Check format
	if !slugPattern.MatchString(normalized) {
		return Slug{}, ErrSlugInvalid(op, value, "must contain only lowercase letters, numbers, and hyphens, and must start and end with a letter or number")
	}

	// Check for consecutive hyphens
	if strings.Contains(normalized, "--") {
		return Slug{}, ErrSlugInvalid(op, value, "must not contain consecutive hyphens")
	}

	// Check reserved list
	reserved := reservedSlugs
	if reserved == nil {
		reserved = defaultReservedSlugs
	}
	if reserved[normalized] {
		return Slug{}, ErrSlugReserved(op, normalized)
	}

	return Slug{value: normalized}, nil
}

// String returns the string representation of the slug.
func (s Slug) String() string {
	return s.value
}

// IsEmpty returns true if the slug is empty.
func (s Slug) IsEmpty() bool {
	return s.value == ""
}

// Equals checks if two slugs are equal.
func (s Slug) Equals(other Slug) bool {
	return s.value == other.value
}

// IsReserved checks if a slug is in the default reserved list.
func IsReserved(slug string) bool {
	return defaultReservedSlugs[strings.ToLower(slug)]
}

// ReservedSlugs returns a copy of the default reserved slugs.
func ReservedSlugs() []string {
	slugs := make([]string, 0, len(defaultReservedSlugs))
	for slug := range defaultReservedSlugs {
		slugs = append(slugs, slug)
	}
	return slugs
}
