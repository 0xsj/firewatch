package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/0xsj/hexagonal-go/internal/identity/domain/auth"
	"github.com/0xsj/hexagonal-go/pkg/database"
	pkgerrors "github.com/0xsj/hexagonal-go/pkg/errors"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// PostgresTokenRepository is a PostgreSQL implementation for password reset tokens.
type PostgresTokenRepository struct {
	db database.DB
}

// NewPostgresTokenRepository creates a new PostgreSQL token repository.
func NewPostgresTokenRepository(db database.DB) *PostgresTokenRepository {
	return &PostgresTokenRepository{
		db: db,
	}
}

// SavePasswordResetToken stores a password reset token.
func (r *PostgresTokenRepository) SavePasswordResetToken(ctx context.Context, token *auth.Token, ipAddress, userAgent string) error {
	const op = "PostgresTokenRepository.SavePasswordResetToken"

	query := `
		INSERT INTO identity.password_reset_tokens (
			id, user_id, tenant_id, token,
			ip_address, user_agent,
			expires_at, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`

	_, err := r.db.Exec(ctx, query,
		token.ID().String(),
		token.UserID().String(),
		token.TenantID(),
		token.Value(),
		stringToNullString(ipAddress),
		stringToNullString(userAgent),
		token.ExpiresAt(),
		token.CreatedAt(),
	)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// FindPasswordResetToken retrieves a valid password reset token by its value.
func (r *PostgresTokenRepository) FindPasswordResetToken(ctx context.Context, tokenValue string) (*auth.Token, error) {
	const op = "PostgresTokenRepository.FindPasswordResetToken"

	query := `
		SELECT
			id, user_id, tenant_id, token,
			expires_at, used_at, created_at
		FROM identity.password_reset_tokens
		WHERE token = $1 AND used_at IS NULL AND expires_at > NOW()
	`

	var row struct {
		ID        string
		UserID    string
		TenantID  string
		Token     string
		ExpiresAt time.Time
		UsedAt    *time.Time
		CreatedAt time.Time
	}

	err := r.db.QueryRow(ctx, query, tokenValue).Scan(
		&row.ID,
		&row.UserID,
		&row.TenantID,
		&row.Token,
		&row.ExpiresAt,
		&row.UsedAt,
		&row.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, pkgerrors.NotFound(op, "password reset token not found or expired")
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Parse IDs
	id, err := types.ParseID(row.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid token id: %w", op, err)
	}

	userID, err := types.ParseID(row.UserID)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid user id: %w", op, err)
	}

	// Reconstruct token
	token := auth.NewTokenWithValue(
		id,
		auth.TokenTypeReset,
		row.Token,
		userID,
		row.TenantID,
		row.ExpiresAt,
		row.CreatedAt,
		row.UsedAt,
		nil, // revokedAt
	)

	return token, nil
}

// MarkPasswordResetTokenUsed marks a token as used.
func (r *PostgresTokenRepository) MarkPasswordResetTokenUsed(ctx context.Context, tokenValue string) error {
	const op = "PostgresTokenRepository.MarkPasswordResetTokenUsed"

	query := `
		UPDATE identity.password_reset_tokens
		SET used_at = NOW()
		WHERE token = $1 AND used_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, tokenValue)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return pkgerrors.NotFound(op, "token not found or already used")
	}

	return nil
}

// DeleteExpiredPasswordResetTokens removes expired tokens.
func (r *PostgresTokenRepository) DeleteExpiredPasswordResetTokens(ctx context.Context) (int, error) {
	const op = "PostgresTokenRepository.DeleteExpiredPasswordResetTokens"

	query := `DELETE FROM identity.password_reset_tokens WHERE expires_at < NOW()`

	result, err := r.db.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	count, _ := result.RowsAffected()
	return int(count), nil
}

// stringToNullString converts a string to sql.NullString.
func stringToNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// SaveEmailVerificationToken stores an email verification token.
func (r *PostgresTokenRepository) SaveEmailVerificationToken(ctx context.Context, token *auth.Token, ipAddress, userAgent string) error {
	const op = "PostgresTokenRepository.SaveEmailVerificationToken"

	query := `
		INSERT INTO identity.email_verification_tokens (
			id, user_id, tenant_id, token,
			ip_address, user_agent,
			expires_at, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`

	_, err := r.db.Exec(ctx, query,
		token.ID().String(),
		token.UserID().String(),
		token.TenantID(),
		token.Value(),
		stringToNullString(ipAddress),
		stringToNullString(userAgent),
		token.ExpiresAt(),
		token.CreatedAt(),
	)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// FindEmailVerificationToken retrieves a valid email verification token by its value.
func (r *PostgresTokenRepository) FindEmailVerificationToken(ctx context.Context, tokenValue string) (*auth.Token, error) {
	const op = "PostgresTokenRepository.FindEmailVerificationToken"

	query := `
		SELECT
			id, user_id, tenant_id, token,
			expires_at, used_at, created_at
		FROM identity.email_verification_tokens
		WHERE token = $1 AND used_at IS NULL AND expires_at > NOW()
	`

	var row struct {
		ID        string
		UserID    string
		TenantID  string
		Token     string
		ExpiresAt time.Time
		UsedAt    *time.Time
		CreatedAt time.Time
	}

	err := r.db.QueryRow(ctx, query, tokenValue).Scan(
		&row.ID,
		&row.UserID,
		&row.TenantID,
		&row.Token,
		&row.ExpiresAt,
		&row.UsedAt,
		&row.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, pkgerrors.NotFound(op, "email verification token not found or expired")
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Parse IDs
	id, err := types.ParseID(row.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid token id: %w", op, err)
	}

	userID, err := types.ParseID(row.UserID)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid user id: %w", op, err)
	}

	// Reconstruct token
	token := auth.NewTokenWithValue(
		id,
		auth.TokenTypeVerification,
		row.Token,
		userID,
		row.TenantID,
		row.ExpiresAt,
		row.CreatedAt,
		row.UsedAt,
		nil, // revokedAt
	)

	return token, nil
}

// MarkEmailVerificationTokenUsed marks a token as used.
func (r *PostgresTokenRepository) MarkEmailVerificationTokenUsed(ctx context.Context, tokenValue string) error {
	const op = "PostgresTokenRepository.MarkEmailVerificationTokenUsed"

	query := `
		UPDATE identity.email_verification_tokens
		SET used_at = NOW()
		WHERE token = $1 AND used_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, tokenValue)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return pkgerrors.NotFound(op, "token not found or already used")
	}

	return nil
}

// DeleteExpiredEmailVerificationTokens removes expired tokens.
func (r *PostgresTokenRepository) DeleteExpiredEmailVerificationTokens(ctx context.Context) (int, error) {
	const op = "PostgresTokenRepository.DeleteExpiredEmailVerificationTokens"

	query := `DELETE FROM identity.email_verification_tokens WHERE expires_at < NOW()`

	result, err := r.db.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	count, _ := result.RowsAffected()
	return int(count), nil
}
