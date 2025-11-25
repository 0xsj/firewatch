package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/0xsj/hexagonal-go/internal/identity/domain/oauth"
	"github.com/0xsj/hexagonal-go/pkg/database"
	pkgerrors "github.com/0xsj/hexagonal-go/pkg/errors"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// PostgresOAuthRepository is a PostgreSQL implementation of oauth.Repository.
type PostgresOAuthRepository struct {
	db database.DB
}

// NewPostgresOAuthRepository creates a new PostgreSQL OAuth repository.
func NewPostgresOAuthRepository(db database.DB) *PostgresOAuthRepository {
	return &PostgresOAuthRepository{
		db: db,
	}
}

// Save persists an OAuth account.
func (r *PostgresOAuthRepository) Save(ctx context.Context, account *oauth.Account) error {
	const op = "PostgresOAuthRepository.Save"

	// Serialize profile data to JSON
	profileJSON, err := json.Marshal(account.ProfileData())
	if err != nil {
		return fmt.Errorf("%s: failed to marshal profile data: %w", op, err)
	}

	query := `
		INSERT INTO identity.oauth_accounts (
			id, user_id, tenant_id, provider, provider_user_id, email,
			access_token, refresh_token, token_expires_at, profile_data,
			created_at, updated_at, last_used_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
		ON CONFLICT (user_id, provider) DO UPDATE SET
			provider_user_id = EXCLUDED.provider_user_id,
			email = EXCLUDED.email,
			access_token = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			token_expires_at = EXCLUDED.token_expires_at,
			profile_data = EXCLUDED.profile_data,
			updated_at = EXCLUDED.updated_at,
			last_used_at = EXCLUDED.last_used_at
	`

	_, err = r.db.Exec(ctx, query,
		account.ID().String(),
		account.UserID().String(),
		account.TenantID(),
		account.Provider().String(),
		account.ProviderUserID(),
		account.Email(),
		account.AccessToken(),
		stringToNullString(account.RefreshToken()),
		nullTime(account.TokenExpiresAt()),
		profileJSON,
		account.CreatedAt().Time(),
		account.UpdatedAt().Time(),
		nullTimestamp(account.LastUsedAt()),
	)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// FindByID retrieves an OAuth account by its ID.
func (r *PostgresOAuthRepository) FindByID(ctx context.Context, id types.ID) (*oauth.Account, error) {
	const op = "PostgresOAuthRepository.FindByID"

	query := `
		SELECT
			id, user_id, tenant_id, provider, provider_user_id, email,
			access_token, refresh_token, token_expires_at, profile_data,
			created_at, updated_at, last_used_at
		FROM identity.oauth_accounts
		WHERE id = $1
	`

	account, err := r.scanAccount(ctx, query, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.NotFound(op, "oauth account not found")
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return account, nil
}

// FindByUserAndProvider finds an OAuth account by user and provider.
func (r *PostgresOAuthRepository) FindByUserAndProvider(ctx context.Context, userID types.ID, provider oauth.Provider) (*oauth.Account, error) {
	const op = "PostgresOAuthRepository.FindByUserAndProvider"

	query := `
		SELECT
			id, user_id, tenant_id, provider, provider_user_id, email,
			access_token, refresh_token, token_expires_at, profile_data,
			created_at, updated_at, last_used_at
		FROM identity.oauth_accounts
		WHERE user_id = $1 AND provider = $2
	`

	account, err := r.scanAccount(ctx, query, userID.String(), provider.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.NotFound(op, "oauth account not found")
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return account, nil
}

// FindByProviderUserID finds an OAuth account by provider and provider's user ID.
func (r *PostgresOAuthRepository) FindByProviderUserID(ctx context.Context, provider oauth.Provider, providerUserID string) (*oauth.Account, error) {
	const op = "PostgresOAuthRepository.FindByProviderUserID"

	query := `
		SELECT
			id, user_id, tenant_id, provider, provider_user_id, email,
			access_token, refresh_token, token_expires_at, profile_data,
			created_at, updated_at, last_used_at
		FROM identity.oauth_accounts
		WHERE provider = $1 AND provider_user_id = $2
	`

	account, err := r.scanAccount(ctx, query, provider.String(), providerUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.NotFound(op, "oauth account not found")
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return account, nil
}

// FindByUser retrieves all OAuth accounts for a user.
func (r *PostgresOAuthRepository) FindByUser(ctx context.Context, userID types.ID) ([]*oauth.Account, error) {
	const op = "PostgresOAuthRepository.FindByUser"

	query := `
		SELECT
			id, user_id, tenant_id, provider, provider_user_id, email,
			access_token, refresh_token, token_expires_at, profile_data,
			created_at, updated_at, last_used_at
		FROM identity.oauth_accounts
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID.String())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var accounts []*oauth.Account
	for rows.Next() {
		account, err := r.scanAccountFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return accounts, nil
}

// Delete removes an OAuth account.
func (r *PostgresOAuthRepository) Delete(ctx context.Context, id types.ID) error {
	const op = "PostgresOAuthRepository.Delete"

	query := `DELETE FROM identity.oauth_accounts WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return pkgerrors.NotFound(op, "oauth account not found")
	}

	return nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// scanAccount scans a single OAuth account from a query.
func (r *PostgresOAuthRepository) scanAccount(ctx context.Context, query string, args ...interface{}) (*oauth.Account, error) {
	row := r.db.QueryRow(ctx, query, args...)
	return r.scanAccountFromRow(row)
}

// scanAccountFromRow scans an OAuth account from a database row.
func (r *PostgresOAuthRepository) scanAccountFromRow(row interface {
	Scan(dest ...interface{}) error
}) (*oauth.Account, error) {
	var (
		id             string
		userID         string
		tenantID       string
		provider       string
		providerUserID string
		email          string
		accessToken    string
		refreshToken   sql.NullString
		tokenExpiresAt sql.NullTime
		profileJSON    []byte
		createdAt      time.Time
		updatedAt      time.Time
		lastUsedAt     sql.NullTime
	)

	err := row.Scan(
		&id,
		&userID,
		&tenantID,
		&provider,
		&providerUserID,
		&email,
		&accessToken,
		&refreshToken,
		&tokenExpiresAt,
		&profileJSON,
		&createdAt,
		&updatedAt,
		&lastUsedAt,
	)
	if err != nil {
		return nil, err
	}

	// Parse IDs
	accountID, err := types.ParseID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid account id: %w", err)
	}

	accountUserID, err := types.ParseID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	// Parse provider
	oauthProvider, err := oauth.ParseProvider(provider)
	if err != nil {
		return nil, fmt.Errorf("invalid provider: %w", err)
	}

	// Deserialize profile data
	var profileData map[string]interface{}
	if len(profileJSON) > 0 {
		if err := json.Unmarshal(profileJSON, &profileData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal profile data: %w", err)
		}
	}

	// Convert timestamps
	var tokenExpiry *time.Time
	if tokenExpiresAt.Valid {
		tokenExpiry = &tokenExpiresAt.Time
	}

	var lastUsed *types.Timestamp
	if lastUsedAt.Valid {
		ts := types.NewTimestamp(lastUsedAt.Time)
		lastUsed = &ts
	}

	// Reconstitute account
	account := oauth.Reconstitute(
		accountID,
		accountUserID,
		tenantID,
		oauthProvider,
		providerUserID,
		email,
		accessToken,
		valueOrEmpty(refreshToken),
		tokenExpiry,
		profileData,
		types.NewTimestamp(createdAt),
		types.NewTimestamp(updatedAt),
		lastUsed,
		1, // version
	)

	return account, nil
}

// valueOrEmpty returns the string value or empty string if null.
func valueOrEmpty(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// nullTimestamp converts a *types.Timestamp to sql.NullTime.
func nullTimestamp(ts *types.Timestamp) sql.NullTime {
	if ts == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: ts.Time(), Valid: true}
}
