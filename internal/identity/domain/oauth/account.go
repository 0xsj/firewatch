package oauth

import (
	"fmt"
	"time"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Account represents an OAuth account linked to a user.
// This is an aggregate that manages the lifecycle of OAuth provider connections.
type Account struct {
	// Identity
	id             types.ID
	userID         types.ID
	tenantID       string
	provider       Provider
	providerUserID string
	email          string

	// Tokens (should be encrypted in production)
	accessToken    string
	refreshToken   string
	tokenExpiresAt *time.Time

	// Profile data from provider (stored as JSON)
	profileData map[string]interface{}

	// Timestamps
	createdAt  types.Timestamp
	updatedAt  types.Timestamp
	lastUsedAt *types.Timestamp

	// Aggregate version
	version int
}

// ============================================================================
// Aggregate Getters
// ============================================================================

func (a *Account) ID() types.ID                        { return a.id }
func (a *Account) UserID() types.ID                    { return a.userID }
func (a *Account) TenantID() string                    { return a.tenantID }
func (a *Account) Provider() Provider                  { return a.provider }
func (a *Account) ProviderUserID() string              { return a.providerUserID }
func (a *Account) Email() string                       { return a.email }
func (a *Account) AccessToken() string                 { return a.accessToken }
func (a *Account) RefreshToken() string                { return a.refreshToken }
func (a *Account) TokenExpiresAt() *time.Time          { return a.tokenExpiresAt }
func (a *Account) ProfileData() map[string]interface{} { return a.profileData }
func (a *Account) CreatedAt() types.Timestamp          { return a.createdAt }
func (a *Account) UpdatedAt() types.Timestamp          { return a.updatedAt }
func (a *Account) LastUsedAt() *types.Timestamp        { return a.lastUsedAt }
func (a *Account) Version() int                        { return a.version }

// ============================================================================
// Factory Methods
// ============================================================================

// NewAccount creates a new OAuth account linked to a user.
func NewAccount(
	id types.ID,
	userID types.ID,
	tenantID string,
	provider Provider,
	providerUserID string,
	email string,
	accessToken string,
	refreshToken string,
	tokenExpiresAt *time.Time,
	profileData map[string]interface{},
) (*Account, error) {
	const op = "oauth.NewAccount"

	// Validate inputs
	if id.IsEmpty() {
		return nil, fmt.Errorf("%s: account id is required", op)
	}
	if userID.IsEmpty() {
		return nil, fmt.Errorf("%s: user id is required", op)
	}
	if tenantID == "" {
		return nil, fmt.Errorf("%s: tenant id is required", op)
	}
	if err := provider.Validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if providerUserID == "" {
		return nil, fmt.Errorf("%s: provider user id is required", op)
	}
	if email == "" {
		return nil, fmt.Errorf("%s: email is required", op)
	}
	if accessToken == "" {
		return nil, fmt.Errorf("%s: access token is required", op)
	}

	now := types.Now()

	account := &Account{
		id:             id,
		userID:         userID,
		tenantID:       tenantID,
		provider:       provider,
		providerUserID: providerUserID,
		email:          email,
		accessToken:    accessToken,
		refreshToken:   refreshToken,
		tokenExpiresAt: tokenExpiresAt,
		profileData:    profileData,
		createdAt:      now,
		updatedAt:      now,
		lastUsedAt:     nil,
		version:        1,
	}

	return account, nil
}

// Reconstitute recreates an OAuth account from stored state.
func Reconstitute(
	id types.ID,
	userID types.ID,
	tenantID string,
	provider Provider,
	providerUserID string,
	email string,
	accessToken string,
	refreshToken string,
	tokenExpiresAt *time.Time,
	profileData map[string]interface{},
	createdAt types.Timestamp,
	updatedAt types.Timestamp,
	lastUsedAt *types.Timestamp,
	version int,
) *Account {
	return &Account{
		id:             id,
		userID:         userID,
		tenantID:       tenantID,
		provider:       provider,
		providerUserID: providerUserID,
		email:          email,
		accessToken:    accessToken,
		refreshToken:   refreshToken,
		tokenExpiresAt: tokenExpiresAt,
		profileData:    profileData,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
		lastUsedAt:     lastUsedAt,
		version:        version,
	}
}

// ============================================================================
// Behavior Methods
// ============================================================================

// UpdateTokens updates the OAuth tokens.
func (a *Account) UpdateTokens(accessToken, refreshToken string, expiresAt *time.Time) {
	a.accessToken = accessToken
	if refreshToken != "" {
		a.refreshToken = refreshToken
	}
	a.tokenExpiresAt = expiresAt
	a.updatedAt = types.Now()
}

// UpdateProfile updates the profile data from the provider.
func (a *Account) UpdateProfile(profileData map[string]interface{}) {
	a.profileData = profileData
	a.updatedAt = types.Now()
}

// RecordUsage records that this OAuth account was used.
func (a *Account) RecordUsage() {
	now := types.Now()
	a.lastUsedAt = &now
	a.updatedAt = now
}

// IsTokenExpired returns true if the access token is expired.
func (a *Account) IsTokenExpired() bool {
	if a.tokenExpiresAt == nil {
		return false // No expiry set (like GitHub)
	}
	return time.Now().After(*a.tokenExpiresAt)
}

// IncrementVersion increments the aggregate version.
func (a *Account) IncrementVersion() {
	a.version++
	a.updatedAt = types.Now()
}
