package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/0xsj/hexagonal-go/internal/identity/domain/session"
	"github.com/0xsj/hexagonal-go/pkg/database"
	pkgerrors "github.com/0xsj/hexagonal-go/pkg/errors"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// PostgresSessionRepository is a PostgreSQL implementation of session.Repository.
type PostgresSessionRepository struct {
	db database.DB
}

// NewPostgresSessionRepository creates a new PostgreSQL session repository.
func NewPostgresSessionRepository(db database.DB) *PostgresSessionRepository {
	return &PostgresSessionRepository{
		db: db,
	}
}

// Save persists a session (insert or update).
func (r *PostgresSessionRepository) Save(ctx context.Context, s *session.Session) error {
	const op = "PostgresSessionRepository.Save"

	query := `
		INSERT INTO identity.sessions (
			id, user_id, tenant_id, provider, status, refresh_token,
			ip_address, user_agent, device_id,
			expires_at, created_at, updated_at, last_active_at, revoked_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9,
			$10, $11, $12, $13, $14
		)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			refresh_token = EXCLUDED.refresh_token,
			ip_address = EXCLUDED.ip_address,
			user_agent = EXCLUDED.user_agent,
			device_id = EXCLUDED.device_id,
			expires_at = EXCLUDED.expires_at,
			updated_at = EXCLUDED.updated_at,
			last_active_at = EXCLUDED.last_active_at,
			revoked_at = EXCLUDED.revoked_at
	`

	snap := s.ToSnapshot()

	_, err := r.db.Exec(ctx, query,
		snap.ID,
		snap.UserID,
		snap.TenantID,
		snap.Provider,
		snap.Status,
		snap.RefreshToken,
		toNullString(snap.IPAddress),
		toNullString(snap.UserAgent),
		toNullString(snap.DeviceID),
		snap.ExpiresAt,
		snap.CreatedAt,
		snap.UpdatedAt,
		snap.LastActiveAt,
		snap.RevokedAt,
	)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// FindByID retrieves a session by ID.
func (r *PostgresSessionRepository) FindByID(ctx context.Context, id types.ID) (*session.Session, error) {
	const op = "PostgresSessionRepository.FindByID"

	query := `
		SELECT
			id, user_id, tenant_id, provider, status, refresh_token,
			ip_address, user_agent, device_id,
			expires_at, created_at, updated_at, last_active_at, revoked_at
		FROM identity.sessions
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id.String())
	s, err := r.scanSession(row)
	if err == sql.ErrNoRows {
		return nil, pkgerrors.NotFound(op, "session")
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return s, nil
}

// FindByRefreshToken retrieves a session by refresh token.
func (r *PostgresSessionRepository) FindByRefreshToken(ctx context.Context, token string) (*session.Session, error) {
	const op = "PostgresSessionRepository.FindByRefreshToken"

	query := `
		SELECT
			id, user_id, tenant_id, provider, status, refresh_token,
			ip_address, user_agent, device_id,
			expires_at, created_at, updated_at, last_active_at, revoked_at
		FROM identity.sessions
		WHERE refresh_token = $1
	`

	row := r.db.QueryRow(ctx, query, token)
	s, err := r.scanSession(row)
	if err == sql.ErrNoRows {
		return nil, pkgerrors.NotFound(op, "session")
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return s, nil
}

// FindActiveByUserID retrieves all active sessions for a user.
func (r *PostgresSessionRepository) FindActiveByUserID(ctx context.Context, userID types.ID) ([]*session.Session, error) {
	const op = "PostgresSessionRepository.FindActiveByUserID"

	query := `
		SELECT
			id, user_id, tenant_id, provider, status, refresh_token,
			ip_address, user_agent, device_id,
			expires_at, created_at, updated_at, last_active_at, revoked_at
		FROM identity.sessions
		WHERE user_id = $1 AND status = 'active' AND expires_at > NOW()
		ORDER BY last_active_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID.String())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	return r.scanSessions(rows, op)
}

// FindActiveByUserIDAndTenant retrieves active sessions for a user in a tenant.
func (r *PostgresSessionRepository) FindActiveByUserIDAndTenant(ctx context.Context, userID types.ID, tenantID string) ([]*session.Session, error) {
	const op = "PostgresSessionRepository.FindActiveByUserIDAndTenant"

	query := `
		SELECT
			id, user_id, tenant_id, provider, status, refresh_token,
			ip_address, user_agent, device_id,
			expires_at, created_at, updated_at, last_active_at, revoked_at
		FROM identity.sessions
		WHERE user_id = $1 AND tenant_id = $2 AND status = 'active' AND expires_at > NOW()
		ORDER BY last_active_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID.String(), tenantID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	return r.scanSessions(rows, op)
}

// CountActiveByUserID returns the count of active sessions for a user.
func (r *PostgresSessionRepository) CountActiveByUserID(ctx context.Context, userID types.ID) (int, error) {
	const op = "PostgresSessionRepository.CountActiveByUserID"

	query := `
		SELECT COUNT(*)
		FROM identity.sessions
		WHERE user_id = $1 AND status = 'active' AND expires_at > NOW()
	`

	var count int
	err := r.db.QueryRow(ctx, query, userID.String()).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return count, nil
}

// DeleteExpired removes expired sessions.
func (r *PostgresSessionRepository) DeleteExpired(ctx context.Context) (int, error) {
	const op = "PostgresSessionRepository.DeleteExpired"

	query := `
		UPDATE identity.sessions
		SET status = 'expired', updated_at = NOW()
		WHERE status = 'active' AND expires_at < NOW()
	`

	result, err := r.db.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	count, _ := result.RowsAffected()
	return int(count), nil
}

// RevokeAllByUserID revokes all sessions for a user.
func (r *PostgresSessionRepository) RevokeAllByUserID(ctx context.Context, userID types.ID, reason string) error {
	const op = "PostgresSessionRepository.RevokeAllByUserID"

	query := `
		UPDATE identity.sessions
		SET status = 'revoked', revoked_at = NOW(), updated_at = NOW()
		WHERE user_id = $1 AND status = 'active'
	`

	_, err := r.db.Exec(ctx, query, userID.String())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// RevokeAllByUserIDExcept revokes all sessions for a user except the given session.
func (r *PostgresSessionRepository) RevokeAllByUserIDExcept(ctx context.Context, userID types.ID, exceptSessionID types.ID, reason string) error {
	const op = "PostgresSessionRepository.RevokeAllByUserIDExcept"

	query := `
		UPDATE identity.sessions
		SET status = 'revoked', revoked_at = NOW(), updated_at = NOW()
		WHERE user_id = $1 AND id != $2 AND status = 'active'
	`

	_, err := r.db.Exec(ctx, query, userID.String(), exceptSessionID.String())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// ============================================================================
// Row Scanning
// ============================================================================

type sessionRow struct {
	ID           string
	UserID       string
	TenantID     string
	Provider     string
	Status       string
	RefreshToken string
	IPAddress    *string
	UserAgent    *string
	DeviceID     *string
	ExpiresAt    time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastActiveAt time.Time
	RevokedAt    *time.Time
}

func (r *PostgresSessionRepository) scanSession(row *sql.Row) (*session.Session, error) {
	var sr sessionRow
	err := row.Scan(
		&sr.ID,
		&sr.UserID,
		&sr.TenantID,
		&sr.Provider,
		&sr.Status,
		&sr.RefreshToken,
		&sr.IPAddress,
		&sr.UserAgent,
		&sr.DeviceID,
		&sr.ExpiresAt,
		&sr.CreatedAt,
		&sr.UpdatedAt,
		&sr.LastActiveAt,
		&sr.RevokedAt,
	)
	if err != nil {
		return nil, err
	}

	return r.rowToSession(sr)
}

func (r *PostgresSessionRepository) scanSessions(rows *sql.Rows, op string) ([]*session.Session, error) {
	sessions := make([]*session.Session, 0)

	for rows.Next() {
		var sr sessionRow
		err := rows.Scan(
			&sr.ID,
			&sr.UserID,
			&sr.TenantID,
			&sr.Provider,
			&sr.Status,
			&sr.RefreshToken,
			&sr.IPAddress,
			&sr.UserAgent,
			&sr.DeviceID,
			&sr.ExpiresAt,
			&sr.CreatedAt,
			&sr.UpdatedAt,
			&sr.LastActiveAt,
			&sr.RevokedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan row: %w", op, err)
		}

		s, err := r.rowToSession(sr)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		sessions = append(sessions, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return sessions, nil
}

func (r *PostgresSessionRepository) rowToSession(sr sessionRow) (*session.Session, error) {
	snap := session.Snapshot{
		ID:           sr.ID,
		UserID:       sr.UserID,
		TenantID:     sr.TenantID,
		Provider:     sr.Provider,
		Status:       sr.Status,
		RefreshToken: sr.RefreshToken,
		IPAddress:    fromNullString(sr.IPAddress),
		UserAgent:    fromNullString(sr.UserAgent),
		DeviceID:     fromNullString(sr.DeviceID),
		ExpiresAt:    sr.ExpiresAt,
		CreatedAt:    sr.CreatedAt,
		UpdatedAt:    sr.UpdatedAt,
		LastActiveAt: sr.LastActiveAt,
		RevokedAt:    sr.RevokedAt,
	}

	return session.FromSnapshot(snap)
}

// ============================================================================
// Helpers
// ============================================================================

func toNullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func fromNullString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
