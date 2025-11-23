package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/0xsj/hexagonal-go/internal/identity/domain/user"
	"github.com/0xsj/hexagonal-go/pkg/database"
	pkgerrors "github.com/0xsj/hexagonal-go/pkg/errors"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// PostgresUserRepository is a PostgreSQL implementation of user.Repository.
// It provides persistent storage for user aggregates.
type PostgresUserRepository struct {
	db database.DB
}

// NewPostgresUserRepository creates a new PostgreSQL user repository.
func NewPostgresUserRepository(db database.DB) *PostgresUserRepository {
	return &PostgresUserRepository{
		db: db,
	}
}

// ============================================================================
// Repository Implementation
// ============================================================================

// Save creates or updates a user in the database.
func (r *PostgresUserRepository) Save(ctx context.Context, u *user.User) error {
	const op = "PostgresUserRepository.Save"

	// Check if user already exists
	existing, err := r.FindByID(ctx, u.ID())
	if err != nil && !pkgerrors.IsNotFound(err) {
		return fmt.Errorf("%s: failed to check existing user: %w", op, err)
	}

	if existing != nil {
		// Update existing user
		return r.update(ctx, u)
	}

	// Insert new user
	return r.insert(ctx, u)
}

// FindByID retrieves a user by their ID.
func (r *PostgresUserRepository) FindByID(ctx context.Context, id types.ID) (*user.User, error) {
	const op = "PostgresUserRepository.FindByID"

	query := `
		SELECT 
			id, tenant_id, email, password_hash,
			status, role, email_verified, email_verified_at,
			failed_login_attempts, locked_until,
			last_login_at, last_login_ip, last_login_user_agent,
			created_at, updated_at
		FROM identity.users
		WHERE id = $1
	`

	var row userRow
	err := r.db.QueryRow(ctx, query, id.String()).Scan(
		&row.ID,
		&row.TenantID,
		&row.Email,
		&row.PasswordHash,
		&row.Status,
		&row.Role,
		&row.EmailVerified,
		&row.EmailVerifiedAt,
		&row.FailedLoginAttempts,
		&row.LockedUntil,
		&row.LastLoginAt,
		&row.LastLoginIP,
		&row.LastLoginUserAgent,
		&row.CreatedAt,
		&row.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, pkgerrors.NotFound(op, fmt.Sprintf("user not found: %s", id.String()))
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return r.rowToUser(row)
}

// FindByEmail retrieves a user by their email.
// Note: In production, should filter by tenant_id from context.
// For now, queries without tenant filtering.
func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email user.Email) (*user.User, error) {
	const op = "PostgresUserRepository.FindByEmail"

	query := `
		SELECT 
			id, tenant_id, email, password_hash,
			status, role, email_verified, email_verified_at,
			failed_login_attempts, locked_until,
			last_login_at, last_login_ip, last_login_user_agent,
			created_at, updated_at
		FROM identity.users
		WHERE email = $1
		LIMIT 1
	`

	var row userRow
	err := r.db.QueryRow(ctx, query, email.String()).Scan(
		&row.ID,
		&row.TenantID,
		&row.Email,
		&row.PasswordHash,
		&row.Status,
		&row.Role,
		&row.EmailVerified,
		&row.EmailVerifiedAt,
		&row.FailedLoginAttempts,
		&row.LockedUntil,
		&row.LastLoginAt,
		&row.LastLoginIP,
		&row.LastLoginUserAgent,
		&row.CreatedAt,
		&row.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, pkgerrors.NotFound(op, fmt.Sprintf("user not found: email=%s", email.String()))
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return r.rowToUser(row)
}

// EmailExists checks if an email is already registered.
// Note: In production, should filter by tenant_id from context.
func (r *PostgresUserRepository) EmailExists(ctx context.Context, email user.Email) (bool, error) {
	const op = "PostgresUserRepository.EmailExists"

	query := `SELECT EXISTS(SELECT 1 FROM identity.users WHERE email = $1)`

	var exists bool
	err := r.db.QueryRow(ctx, query, email.String()).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return exists, nil
}

// List retrieves users matching the given filters.
func (r *PostgresUserRepository) List(ctx context.Context, filters user.Filters) ([]*user.User, error) {
	const op = "PostgresUserRepository.List"

	// Build dynamic query with filters
	query, args := r.buildListQuery(filters)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	users := make([]*user.User, 0)
	for rows.Next() {
		var row userRow
		err := rows.Scan(
			&row.ID,
			&row.TenantID,
			&row.Email,
			&row.PasswordHash,
			&row.Status,
			&row.Role,
			&row.EmailVerified,
			&row.EmailVerifiedAt,
			&row.FailedLoginAttempts,
			&row.LockedUntil,
			&row.LastLoginAt,
			&row.LastLoginIP,
			&row.LastLoginUserAgent,
			&row.CreatedAt,
			&row.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan row: %w", op, err)
		}

		u, err := r.rowToUser(row)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return users, nil
}

// Count returns the total number of users matching the given filters.
func (r *PostgresUserRepository) Count(ctx context.Context, filters user.Filters) (int, error) {
	const op = "PostgresUserRepository.Count"

	// Build dynamic count query
	query, args := r.buildCountQuery(filters)

	var count int
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return count, nil
}

// Delete removes a user from the database.
func (r *PostgresUserRepository) Delete(ctx context.Context, id types.ID) error {
	const op = "PostgresUserRepository.Delete"

	query := `DELETE FROM identity.users WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return pkgerrors.NotFound(op, fmt.Sprintf("user not found: %s", id.String()))
	}

	return nil
}

// ============================================================================
// Private Helper Methods
// ============================================================================

// insert creates a new user in the database.
func (r *PostgresUserRepository) insert(ctx context.Context, u *user.User) error {
	const op = "PostgresUserRepository.insert"

	query := `
		INSERT INTO identity.users (
			id, tenant_id, email, password_hash,
			status, role, email_verified, email_verified_at,
			failed_login_attempts, locked_until,
			last_login_at, last_login_ip, last_login_user_agent,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`

	snapshot := u.Snapshot()

	_, err := r.db.Exec(ctx, query,
		snapshot.ID,
		snapshot.TenantID,
		snapshot.Email,
		snapshot.PasswordHash,
		snapshot.Status,
		snapshot.Role,
		snapshot.EmailVerified,
		timestampToNullTime(snapshot.EmailVerifiedAt),
		snapshot.FailedLoginAttempts,
		nullTime(snapshot.LockedUntil),
		timestampToNullTime(snapshot.LastLoginAt),
		nullString(snapshot.LastLoginIP),
		nullString(snapshot.LastLoginUserAgent),
		snapshot.CreatedAt.Time(),
		snapshot.UpdatedAt.Time(),
	)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// update modifies an existing user in the database.
func (r *PostgresUserRepository) update(ctx context.Context, u *user.User) error {
	const op = "PostgresUserRepository.update"

	query := `
		UPDATE identity.users SET
			tenant_id = $2,
			email = $3,
			password_hash = $4,
			status = $5,
			role = $6,
			email_verified = $7,
			email_verified_at = $8,
			failed_login_attempts = $9,
			locked_until = $10,
			last_login_at = $11,
			last_login_ip = $12,
			last_login_user_agent = $13,
			updated_at = $14
		WHERE id = $1
	`

	snapshot := u.Snapshot()

	result, err := r.db.Exec(ctx, query,
		snapshot.ID,
		snapshot.TenantID,
		snapshot.Email,
		snapshot.PasswordHash,
		snapshot.Status,
		snapshot.Role,
		snapshot.EmailVerified,
		timestampToNullTime(snapshot.EmailVerifiedAt),
		snapshot.FailedLoginAttempts,
		nullTime(snapshot.LockedUntil),
		timestampToNullTime(snapshot.LastLoginAt),
		nullString(snapshot.LastLoginIP),
		nullString(snapshot.LastLoginUserAgent),
		snapshot.UpdatedAt.Time(),
	)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return pkgerrors.NotFound(op, fmt.Sprintf("user not found: %s", snapshot.ID))
	}

	return nil
}

// buildListQuery constructs a dynamic SELECT query with filters.
func (r *PostgresUserRepository) buildListQuery(filters user.Filters) (string, []interface{}) {
	query := `
		SELECT 
			id, tenant_id, email, password_hash,
			status, role, email_verified, email_verified_at,
			failed_login_attempts, locked_until,
			last_login_at, last_login_ip, last_login_user_agent,
			created_at, updated_at
		FROM identity.users
		WHERE 1=1
	`

	args := make([]interface{}, 0)
	argIndex := 1

	// Apply filters
	if filters.TenantID != nil {
		query += fmt.Sprintf(" AND tenant_id = $%d", argIndex)
		args = append(args, *filters.TenantID)
		argIndex++
	}

	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, string(*filters.Status))
		argIndex++
	}

	if filters.Role != nil {
		query += fmt.Sprintf(" AND role = $%d", argIndex)
		args = append(args, string(*filters.Role))
		argIndex++
	}

	if filters.EmailVerified != nil {
		query += fmt.Sprintf(" AND email_verified = $%d", argIndex)
		args = append(args, *filters.EmailVerified)
		argIndex++
	}

	if filters.EmailContains != "" {
		query += fmt.Sprintf(" AND email ILIKE $%d", argIndex)
		args = append(args, "%"+filters.EmailContains+"%")
		argIndex++
	}

	// Sorting
	sortBy := "created_at"
	if filters.SortBy != "" {
		sortBy = string(filters.SortBy)
	}

	sortOrder := "DESC"
	if filters.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	// Pagination
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, filters.Limit, filters.Offset)

	return query, args
}

// buildCountQuery constructs a dynamic COUNT query with filters.
func (r *PostgresUserRepository) buildCountQuery(filters user.Filters) (string, []interface{}) {
	query := `SELECT COUNT(*) FROM identity.users WHERE 1=1`

	args := make([]interface{}, 0)
	argIndex := 1

	// Apply same filters as List (without sorting/pagination)
	if filters.TenantID != nil {
		query += fmt.Sprintf(" AND tenant_id = $%d", argIndex)
		args = append(args, *filters.TenantID)
		argIndex++
	}

	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, string(*filters.Status))
		argIndex++
	}

	if filters.Role != nil {
		query += fmt.Sprintf(" AND role = $%d", argIndex)
		args = append(args, string(*filters.Role))
		argIndex++
	}

	if filters.EmailVerified != nil {
		query += fmt.Sprintf(" AND email_verified = $%d", argIndex)
		args = append(args, *filters.EmailVerified)
		argIndex++
	}

	if filters.EmailContains != "" {
		query += fmt.Sprintf(" AND email ILIKE $%d", argIndex)
		args = append(args, "%"+filters.EmailContains+"%")
		argIndex++
	}

	return query, args
}

// rowToUser converts a database row to a User domain entity.
func (r *PostgresUserRepository) rowToUser(row userRow) (*user.User, error) {
	const op = "PostgresUserRepository.rowToUser"

	// Parse ID
	id, err := types.ParseID(row.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid user ID: %w", op, err)
	}

	// Parse email
	email, err := user.NewEmail(row.Email)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid email: %w", op, err)
	}

	// Parse status and role
	status := user.Status(row.Status)
	role := user.Role(row.Role)

	// Build snapshot
	snapshot := user.Snapshot{
		ID:                  id.String(),
		TenantID:            row.TenantID,
		Email:               email.String(),
		PasswordHash:        nullableString(row.PasswordHash),
		Status:              string(status),
		Role:                string(role),
		EmailVerified:       row.EmailVerified,
		EmailVerifiedAt:     nullTimeToTimestamp(row.EmailVerifiedAt),
		FailedLoginAttempts: row.FailedLoginAttempts,
		LockedUntil:         nullableTime(row.LockedUntil),
		LastLoginAt:         nullTimeToTimestamp(row.LastLoginAt),
		LastLoginIP:         nullableString(row.LastLoginIP),
		LastLoginUserAgent:  nullableString(row.LastLoginUserAgent),
		CreatedAt:           types.TimestampFromTime(row.CreatedAt),
		UpdatedAt:           types.TimestampFromTime(row.UpdatedAt),
	}

	return user.LoadFromSnapshot(snapshot)
}

// ============================================================================
// Database Row Type
// ============================================================================

// userRow represents a row from the users table.
type userRow struct {
	ID                  string
	TenantID            string
	Email               string
	PasswordHash        sql.NullString
	Status              string
	Role                string
	EmailVerified       bool
	EmailVerifiedAt     sql.NullTime
	FailedLoginAttempts int
	LockedUntil         sql.NullTime
	LastLoginAt         sql.NullTime
	LastLoginIP         sql.NullString
	LastLoginUserAgent  sql.NullString
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// ============================================================================
// Helper Functions for NULL Handling
// ============================================================================

// nullTime converts a *time.Time to sql.NullTime.
func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

// nullString converts a *string to sql.NullString.
func nullString(s *string) sql.NullString {
	if s == nil || *s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

// nullableTime converts sql.NullTime to *time.Time.
func nullableTime(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	return &nt.Time
}

// nullableString converts sql.NullString to *string.
func nullableString(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

// timestampToNullTime converts *types.Timestamp to sql.NullTime.
func timestampToNullTime(ts *types.Timestamp) sql.NullTime {
	if ts == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: ts.Time(), Valid: true}
}

// nullTimeToTimestamp converts sql.NullTime to *types.Timestamp.
func nullTimeToTimestamp(nt sql.NullTime) *types.Timestamp {
	if !nt.Valid {
		return nil
	}
	ts := types.TimestampFromTime(nt.Time)
	return &ts
}
