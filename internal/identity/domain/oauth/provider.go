package oauth

import "fmt"

// Provider represents an OAuth provider type.
type Provider string

const (
	ProviderGoogle Provider = "google"
	ProviderGitHub Provider = "github"
)

// String returns the string representation of the provider.
func (p Provider) String() string {
	return string(p)
}

// Validate validates the provider value.
func (p Provider) Validate() error {
	switch p {
	case ProviderGoogle, ProviderGitHub:
		return nil
	default:
		return fmt.Errorf("invalid oauth provider: %s", p)
	}
}

// IsValid returns true if the provider is valid.
func (p Provider) IsValid() bool {
	return p.Validate() == nil
}

// ParseProvider parses a string into a Provider.
func ParseProvider(s string) (Provider, error) {
	p := Provider(s)
	if err := p.Validate(); err != nil {
		return "", err
	}
	return p, nil
}
