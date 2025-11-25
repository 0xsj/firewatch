package auth

import "fmt"

// Provider represents an authentication provider.
type Provider string

const (
	// ProviderPassword is traditional email/password authentication.
	ProviderPassword Provider = "password"

	// ProviderGoogle is Google OAuth authentication.
	ProviderGoogle Provider = "google"

	// ProviderGitHub is GitHub OAuth authentication.
	ProviderGitHub Provider = "github"

	// ProviderMagicLink is passwordless email authentication.
	ProviderMagicLink Provider = "magic_link"
)

// String returns the string representation.
func (p Provider) String() string {
	return string(p)
}

// Validate validates the provider value.
func (p Provider) Validate() error {
	if !p.IsValid() {
		return fmt.Errorf("invalid auth provider: %s", p)
	}
	return nil
}

// IsValid checks if the provider is valid.
func (p Provider) IsValid() bool {
	switch p {
	case ProviderPassword, ProviderGoogle, ProviderGitHub, ProviderMagicLink:
		return true
	default:
		return false
	}
}

// IsOAuth returns true if this is an OAuth provider.
func (p Provider) IsOAuth() bool {
	switch p {
	case ProviderGoogle, ProviderGitHub:
		return true
	default:
		return false
	}
}

// IsPasswordless returns true if this provider doesn't require a password.
func (p Provider) IsPasswordless() bool {
	switch p {
	case ProviderGoogle, ProviderGitHub, ProviderMagicLink:
		return true
	default:
		return false
	}
}

// OAuthProviders returns all OAuth providers.
func OAuthProviders() []Provider {
	return []Provider{ProviderGoogle, ProviderGitHub}
}

// AllProviders returns all authentication providers.
func AllProviders() []Provider {
	return []Provider{ProviderPassword, ProviderGoogle, ProviderGitHub, ProviderMagicLink}
}

// ParseProvider parses a string into a Provider.
func ParseProvider(s string) (Provider, error) {
	p := Provider(s)
	if err := p.Validate(); err != nil {
		return "", err
	}
	return p, nil
}
