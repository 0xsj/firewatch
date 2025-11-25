package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/0xsj/hexagonal-go/pkg/oauth"
)

const (
	authURL      = "https://github.com/login/oauth/authorize"
	tokenURL     = "https://github.com/login/oauth/access_token"
	userURL      = "https://api.github.com/user"
	userEmailURL = "https://api.github.com/user/emails"
)

// Provider implements OAuth provider for GitHub
type Provider struct {
	config oauth.Config
}

// NewProvider creates a new GitHub OAuth provider
func NewProvider(config oauth.Config) (*Provider, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Set default scopes if not provided
	if len(config.Scopes) == 0 {
		config.Scopes = []string{
			"user:email", // Access user email addresses
		}
	}

	return &Provider{
		config: config,
	}, nil
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "github"
}

// AuthCodeURL generates the OAuth authorization URL
func (p *Provider) AuthCodeURL(state string) string {
	params := url.Values{
		"client_id":    {p.config.ClientID},
		"redirect_uri": {p.config.RedirectURL},
		"scope":        {strings.Join(p.config.Scopes, " ")},
		"state":        {state},
	}

	return authURL + "?" + params.Encode()
}

// Exchange exchanges an authorization code for tokens and user info
func (p *Provider) Exchange(ctx context.Context, code string) (*oauth.UserInfo, *oauth.Tokens, error) {
	// Exchange code for tokens
	tokens, err := p.exchangeCode(ctx, code)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Fetch user info
	userInfo, err := p.fetchUserInfo(ctx, tokens.AccessToken)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch user info: %w", err)
	}

	return userInfo, tokens, nil
}

// RefreshToken - GitHub doesn't support refresh tokens for OAuth apps
// GitHub Apps would need to be used for refresh tokens
func (p *Provider) RefreshToken(ctx context.Context, refreshToken string) (*oauth.Tokens, error) {
	return nil, fmt.Errorf("github oauth apps do not support refresh tokens")
}

// exchangeCode exchanges authorization code for tokens
func (p *Provider) exchangeCode(ctx context.Context, code string) (*oauth.Tokens, error) {
	data := url.Values{
		"client_id":     {p.config.ClientID},
		"client_secret": {p.config.ClientSecret},
		"code":          {code},
		"redirect_uri":  {p.config.RedirectURL},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: %s", oauth.ErrTokenExchange, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	return &oauth.Tokens{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: "", // GitHub doesn't provide refresh tokens for OAuth apps
		ExpiresAt:    0,  // GitHub tokens don't expire
		TokenType:    tokenResp.TokenType,
	}, nil
}

// fetchUserInfo fetches user information from GitHub
func (p *Provider) fetchUserInfo(ctx context.Context, accessToken string) (*oauth.UserInfo, error) {
	// Fetch user profile
	user, err := p.fetchUser(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	// Fetch primary email (GitHub may not return email in user profile)
	if user.Email == "" {
		email, verified, err := p.fetchPrimaryEmail(ctx, accessToken)
		if err != nil {
			return nil, err
		}
		user.Email = email
		user.EmailVerified = verified
	}

	return user, nil
}

// fetchUser fetches the user profile
func (p *Provider) fetchUser(ctx context.Context, accessToken string) (*oauth.UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", userURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: %s", oauth.ErrUserInfoFetch, string(body))
	}

	var githubUser struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
		Location  string `json:"location"`
		Company   string `json:"company"`
		Bio       string `json:"bio"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		return nil, err
	}

	// Convert to generic UserInfo
	userInfo := &oauth.UserInfo{
		ID:            fmt.Sprintf("%d", githubUser.ID),
		Email:         githubUser.Email,
		EmailVerified: false, // Will be updated if we fetch emails
		Name:          githubUser.Name,
		GivenName:     "", // GitHub doesn't provide separate first/last names
		FamilyName:    "",
		Picture:       githubUser.AvatarURL,
		Locale:        "",
		Raw: map[string]interface{}{
			"id":         githubUser.ID,
			"login":      githubUser.Login,
			"email":      githubUser.Email,
			"name":       githubUser.Name,
			"avatar_url": githubUser.AvatarURL,
			"location":   githubUser.Location,
			"company":    githubUser.Company,
			"bio":        githubUser.Bio,
		},
	}

	return userInfo, nil
}

// fetchPrimaryEmail fetches the user's primary verified email
func (p *Provider) fetchPrimaryEmail(ctx context.Context, accessToken string) (string, bool, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", userEmailURL, nil)
	if err != nil {
		return "", false, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", false, fmt.Errorf("failed to fetch emails: %s", string(body))
	}

	var emails []struct {
		Email      string `json:"email"`
		Primary    bool   `json:"primary"`
		Verified   bool   `json:"verified"`
		Visibility string `json:"visibility"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", false, err
	}

	// Find primary verified email
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, true, nil
		}
	}

	// Fallback to first verified email
	for _, e := range emails {
		if e.Verified {
			return e.Email, true, nil
		}
	}

	// No verified email found
	if len(emails) > 0 {
		return emails[0].Email, false, nil
	}

	return "", false, fmt.Errorf("no email found")
}
