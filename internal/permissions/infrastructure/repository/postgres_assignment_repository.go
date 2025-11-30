package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/0xsj/hexagonal-go/internal/permissions/domain"
	"github.com/0xsj/hexagonal-go/pkg/database"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// PostgresAssignmentRepository is a PostgreSQL implementation of domain.AssignmentRepository.
type PostgresAssignmentRepository struct {
	db database.DB
}

// NewPostgresAssignmentRepository creates a new PostgreSQL assignment repository.
func NewPostgresAssignmentRepository(db database.DB) *PostgresAssignmentRepository {
	return &PostgresAssignmentRepository{
		db: db,
	}
}

// ============================================================================
// Repository Implementation
// ============================================================================

// Save persists an assignment.
func (r *PostgresAssignmentRepository) Save(ctx context.Context, assignment *domain.Assignment) error {
	const op = "PostgresAssignmentRepository.Save"

	query := `
		INSERT INTO permissions.assignments (
			id, user_id, role_id, tenant_id, resource_id,
			expires_at, created_at, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id, role_id, tenant_id, COALESCE(resource_id, '')) DO UPDATE SET
			expires_at = EXCLUDED.expires_at
	`

	_, err := r.db.Exec(ctx, query,
		assignment.ID().String(),
		assignment.UserID().String(),
		assignment.RoleID().String(),
		assignment.TenantID(),
		nullString(stringPtr(assignment.ResourceID())),
		nullTime(assignment.ExpiresAt()),
		assignment.CreatedAt().Time(),
		assignment.CreatedBy(),
	)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// FindByID retrieves an assignment by its ID.
func (r *PostgresAssignmentRepository) FindByID(ctx context.Context, id types.ID) (*domain.Assignment, error) {
	const op = "PostgresAssignmentRepository.FindByID"

	query := `
		SELECT id, user_id, role_id, tenant_id, resource_id,
		       expires_at, created_at, created_by
		FROM permissions.assignments
		WHERE id = $1
	`

	var row assignmentRow
	err := r.db.QueryRow(ctx, query, id.String()).Scan(
		&row.ID,
		&row.UserID,
		&row.RoleID,
		&row.TenantID,
		&row.ResourceID,
		&row.ExpiresAt,
		&row.CreatedAt,
		&row.CreatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrAssignmentNotFound(op, id.String())
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return r.rowToAssignment(row)
}

// FindByUserAndRole retrieves an assignment by user, role, and scope.
func (r *PostgresAssignmentRepository) FindByUserAndRole(ctx context.Context, userID, roleID types.ID, tenantID, resourceID string) (*domain.Assignment, error) {
	const op = "PostgresAssignmentRepository.FindByUserAndRole"

	query := `
		SELECT id, user_id, role_id, tenant_id, resource_id,
		       expires_at, created_at, created_by
		FROM permissions.assignments
		WHERE user_id = $1 AND role_id = $2 AND tenant_id = $3 
		  AND COALESCE(resource_id, '') = COALESCE($4, '')
	`

	var row assignmentRow
	err := r.db.QueryRow(ctx, query, userID.String(), roleID.String(), tenantID, nullString(stringPtr(resourceID))).Scan(
		&row.ID,
		&row.UserID,
		&row.RoleID,
		&row.TenantID,
		&row.ResourceID,
		&row.ExpiresAt,
		&row.CreatedAt,
		&row.CreatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrAssignmentNotFound(op, "")
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return r.rowToAssignment(row)
}

// FindByUser retrieves all assignments for a user.
func (r *PostgresAssignmentRepository) FindByUser(ctx context.Context, userID types.ID, filters *domain.AssignmentFilters) ([]*domain.Assignment, error) {
	const op = "PostgresAssignmentRepository.FindByUser"

	query, args := r.buildUserQuery(userID, filters)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	return r.scanAssignments(op, rows)
}

// FindByRole retrieves all assignments for a role.
func (r *PostgresAssignmentRepository) FindByRole(ctx context.Context, roleID types.ID, filters *domain.AssignmentFilters) ([]*domain.Assignment, error) {
	const op = "PostgresAssignmentRepository.FindByRole"

	query, args := r.buildRoleQuery(roleID, filters)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	return r.scanAssignments(op, rows)
}

// FindByTenant retrieves all assignments within a tenant.
func (r *PostgresAssignmentRepository) FindByTenant(ctx context.Context, tenantID string, filters *domain.AssignmentFilters) ([]*domain.Assignment, error) {
	const op = "PostgresAssignmentRepository.FindByTenant"

	query, args := r.buildTenantQuery(tenantID, filters)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	return r.scanAssignments(op, rows)
}

// Delete removes an assignment by ID.
func (r *PostgresAssignmentRepository) Delete(ctx context.Context, id types.ID) error {
	const op = "PostgresAssignmentRepository.Delete"

	query := `DELETE FROM permissions.assignments WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return domain.ErrAssignmentNotFound(op, id.String())
	}

	return nil
}

// DeleteByUserAndRole removes an assignment by user, role, and scope.
func (r *PostgresAssignmentRepository) DeleteByUserAndRole(ctx context.Context, userID, roleID types.ID, tenantID, resourceID string) error {
	const op = "PostgresAssignmentRepository.DeleteByUserAndRole"

	query := `
		DELETE FROM permissions.assignments
		WHERE user_id = $1 AND role_id = $2 AND tenant_id = $3
		  AND COALESCE(resource_id, '') = COALESCE($4, '')
	`

	result, err := r.db.Exec(ctx, query, userID.String(), roleID.String(), tenantID, nullString(stringPtr(resourceID)))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return domain.ErrAssignmentNotFound(op, "")
	}

	return nil
}

// DeleteByUser removes all assignments for a user.
func (r *PostgresAssignmentRepository) DeleteByUser(ctx context.Context, userID types.ID) error {
	const op = "PostgresAssignmentRepository.DeleteByUser"

	query := `DELETE FROM permissions.assignments WHERE user_id = $1`

	_, err := r.db.Exec(ctx, query, userID.String())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// DeleteByRole removes all assignments for a role.
func (r *PostgresAssignmentRepository) DeleteByRole(ctx context.Context, roleID types.ID) error {
	const op = "PostgresAssignmentRepository.DeleteByRole"

	query := `DELETE FROM permissions.assignments WHERE role_id = $1`

	_, err := r.db.Exec(ctx, query, roleID.String())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// DeleteExpired removes all expired assignments.
func (r *PostgresAssignmentRepository) DeleteExpired(ctx context.Context) (int64, error) {
	const op = "PostgresAssignmentRepository.DeleteExpired"

	query := `DELETE FROM permissions.assignments WHERE expires_at IS NOT NULL AND expires_at < NOW()`

	result, err := r.db.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}

	return rowsAffected, nil
}

// Exists checks if an assignment exists.
func (r *PostgresAssignmentRepository) Exists(ctx context.Context, userID, roleID types.ID, tenantID, resourceID string) (bool, error) {
	const op = "PostgresAssignmentRepository.Exists"

	query := `
		SELECT EXISTS(
			SELECT 1 FROM permissions.assignments
			WHERE user_id = $1 AND role_id = $2 AND tenant_id = $3
			  AND COALESCE(resource_id, '') = COALESCE($4, '')
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, userID.String(), roleID.String(), tenantID, nullString(stringPtr(resourceID))).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return exists, nil
}

// ============================================================================
// Private Helper Methods
// ============================================================================

// buildUserQuery constructs a SELECT query for user assignments.
func (r *PostgresAssignmentRepository) buildUserQuery(userID types.ID, filters *domain.AssignmentFilters) (string, []interface{}) {
	query := `
		SELECT id, user_id, role_id, tenant_id, resource_id,
		       expires_at, created_at, created_by
		FROM permissions.assignments
		WHERE user_id = $1
	`

	args := []interface{}{userID.String()}
	argIndex := 2

	if filters != nil {
		query, args, argIndex = r.applyFilters(query, args, argIndex, filters)
	}

	query += " ORDER BY created_at DESC"

	if filters != nil {
		query, args = r.applyPagination(query, args, argIndex, filters)
	}

	return query, args
}

// buildRoleQuery constructs a SELECT query for role assignments.
func (r *PostgresAssignmentRepository) buildRoleQuery(roleID types.ID, filters *domain.AssignmentFilters) (string, []interface{}) {
	query := `
		SELECT id, user_id, role_id, tenant_id, resource_id,
		       expires_at, created_at, created_by
		FROM permissions.assignments
		WHERE role_id = $1
	`

	args := []interface{}{roleID.String()}
	argIndex := 2

	if filters != nil {
		query, args, argIndex = r.applyFilters(query, args, argIndex, filters)
	}

	query += " ORDER BY created_at DESC"

	if filters != nil {
		query, args = r.applyPagination(query, args, argIndex, filters)
	}

	return query, args
}

// buildTenantQuery constructs a SELECT query for tenant assignments.
func (r *PostgresAssignmentRepository) buildTenantQuery(tenantID string, filters *domain.AssignmentFilters) (string, []interface{}) {
	query := `
		SELECT id, user_id, role_id, tenant_id, resource_id,
		       expires_at, created_at, created_by
		FROM permissions.assignments
		WHERE tenant_id = $1
	`

	args := []interface{}{tenantID}
	argIndex := 2

	if filters != nil {
		if filters.ResourceID != "" {
			query += fmt.Sprintf(" AND resource_id = $%d", argIndex)
			args = append(args, filters.ResourceID)
			argIndex++
		}

		if !filters.IncludeExpired {
			query += " AND (expires_at IS NULL OR expires_at > NOW())"
		}
	}

	query += " ORDER BY created_at DESC"

	if filters != nil {
		query, args = r.applyPagination(query, args, argIndex, filters)
	}

	return query, args
}

// applyFilters adds common filter conditions to a query.
func (r *PostgresAssignmentRepository) applyFilters(query string, args []interface{}, argIndex int, filters *domain.AssignmentFilters) (string, []interface{}, int) {
	if filters.TenantID != "" {
		query += fmt.Sprintf(" AND tenant_id = $%d", argIndex)
		args = append(args, filters.TenantID)
		argIndex++
	}

	if filters.ResourceID != "" {
		query += fmt.Sprintf(" AND resource_id = $%d", argIndex)
		args = append(args, filters.ResourceID)
		argIndex++
	}

	if !filters.IncludeExpired {
		query += " AND (expires_at IS NULL OR expires_at > NOW())"
	}

	return query, args, argIndex
}

// applyPagination adds LIMIT and OFFSET to a query.
func (r *PostgresAssignmentRepository) applyPagination(query string, args []interface{}, argIndex int, filters *domain.AssignmentFilters) (string, []interface{}) {
	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filters.Offset)
	}

	return query, args
}

// scanAssignments scans multiple assignment rows.
func (r *PostgresAssignmentRepository) scanAssignments(op string, rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]*domain.Assignment, error) {
	assignments := make([]*domain.Assignment, 0)

	for rows.Next() {
		var row assignmentRow
		err := rows.Scan(
			&row.ID,
			&row.UserID,
			&row.RoleID,
			&row.TenantID,
			&row.ResourceID,
			&row.ExpiresAt,
			&row.CreatedAt,
			&row.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan row: %w", op, err)
		}

		assignment, err := r.rowToAssignment(row)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		assignments = append(assignments, assignment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return assignments, nil
}

// rowToAssignment converts a database row to a domain Assignment.
func (r *PostgresAssignmentRepository) rowToAssignment(row assignmentRow) (*domain.Assignment, error) {
	const op = "PostgresAssignmentRepository.rowToAssignment"

	id, err := types.ParseID(row.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid assignment ID: %w", op, err)
	}

	userID, err := types.ParseID(row.UserID)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid user ID: %w", op, err)
	}

	roleID, err := types.ParseID(row.RoleID)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid role ID: %w", op, err)
	}

	return domain.ReconstituteAssignment(
		id,
		userID,
		roleID,
		row.TenantID,
		nullableString(row.ResourceID),
		nullableTime(row.ExpiresAt),
		types.TimestampFromTime(row.CreatedAt),
		row.CreatedBy,
	), nil
}

// ============================================================================
// Database Row Type
// ============================================================================

// assignmentRow represents a row from the assignments table.
type assignmentRow struct {
	ID         string
	UserID     string
	RoleID     string
	TenantID   string
	ResourceID sql.NullString
	ExpiresAt  sql.NullTime
	CreatedAt  time.Time
	CreatedBy  string
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

// nullableString converts sql.NullString to string (empty if null).
func nullableString(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}

// stringPtr converts a string to *string (nil if empty).
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
