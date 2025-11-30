package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/0xsj/hexagonal-go/internal/permissions/domain"
	"github.com/0xsj/hexagonal-go/pkg/database"
	pkgerrors "github.com/0xsj/hexagonal-go/pkg/errors"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// PostgresRoleRepository is a PostgreSQL implementation of domain.RoleRepository.
type PostgresRoleRepository struct {
	db database.DB
}

// NewPostgresRoleRepository creates a new PostgreSQL role repository.
func NewPostgresRoleRepository(db database.DB) *PostgresRoleRepository {
	return &PostgresRoleRepository{
		db: db,
	}
}

// ============================================================================
// Repository Implementation
// ============================================================================

// Save persists a role (create or update).
func (r *PostgresRoleRepository) Save(ctx context.Context, role *domain.Role) error {
	const op = "PostgresRoleRepository.Save"

	existing, err := r.FindByID(ctx, role.ID())
	if err != nil && !pkgerrors.IsNotFound(err) {
		return fmt.Errorf("%s: failed to check existing role: %w", op, err)
	}

	if existing != nil {
		return r.update(ctx, role)
	}

	return r.insert(ctx, role)
}

// FindByID retrieves a role by its ID.
func (r *PostgresRoleRepository) FindByID(ctx context.Context, id types.ID) (*domain.Role, error) {
	const op = "PostgresRoleRepository.FindByID"

	query := `
		SELECT id, tenant_id, name, description, permissions, is_system,
		       created_at, updated_at, version
		FROM permissions.roles
		WHERE id = $1
	`

	var row roleRow
	err := r.db.QueryRow(ctx, query, id.String()).Scan(
		&row.ID,
		&row.TenantID,
		&row.Name,
		&row.Description,
		&row.Permissions,
		&row.IsSystem,
		&row.CreatedAt,
		&row.UpdatedAt,
		&row.Version,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrRoleNotFound(op, id.String())
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return r.rowToRole(row)
}

// FindByName retrieves a role by name within a tenant.
func (r *PostgresRoleRepository) FindByName(ctx context.Context, tenantID, name string) (*domain.Role, error) {
	const op = "PostgresRoleRepository.FindByName"

	query := `
		SELECT id, tenant_id, name, description, permissions, is_system,
		       created_at, updated_at, version
		FROM permissions.roles
		WHERE (tenant_id = $1 OR tenant_id = '*') AND name = $2
		ORDER BY tenant_id DESC
		LIMIT 1
	`

	var row roleRow
	err := r.db.QueryRow(ctx, query, tenantID, name).Scan(
		&row.ID,
		&row.TenantID,
		&row.Name,
		&row.Description,
		&row.Permissions,
		&row.IsSystem,
		&row.CreatedAt,
		&row.UpdatedAt,
		&row.Version,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrRoleNotFoundByName(op, name)
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return r.rowToRole(row)
}

// FindAll retrieves all roles for a tenant.
func (r *PostgresRoleRepository) FindAll(ctx context.Context, tenantID string, filters *domain.RoleFilters) ([]*domain.Role, error) {
	const op = "PostgresRoleRepository.FindAll"

	query, args := r.buildListQuery(tenantID, filters)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	roles := make([]*domain.Role, 0)
	for rows.Next() {
		var row roleRow
		err := rows.Scan(
			&row.ID,
			&row.TenantID,
			&row.Name,
			&row.Description,
			&row.Permissions,
			&row.IsSystem,
			&row.CreatedAt,
			&row.UpdatedAt,
			&row.Version,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan row: %w", op, err)
		}

		role, err := r.rowToRole(row)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return roles, nil
}

// FindSystemRoles retrieves all system roles.
func (r *PostgresRoleRepository) FindSystemRoles(ctx context.Context) ([]*domain.Role, error) {
	const op = "PostgresRoleRepository.FindSystemRoles"

	query := `
		SELECT id, tenant_id, name, description, permissions, is_system,
		       created_at, updated_at, version
		FROM permissions.roles
		WHERE is_system = true
		ORDER BY name ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	roles := make([]*domain.Role, 0)
	for rows.Next() {
		var row roleRow
		err := rows.Scan(
			&row.ID,
			&row.TenantID,
			&row.Name,
			&row.Description,
			&row.Permissions,
			&row.IsSystem,
			&row.CreatedAt,
			&row.UpdatedAt,
			&row.Version,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan row: %w", op, err)
		}

		role, err := r.rowToRole(row)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return roles, nil
}

// Delete removes a role by ID.
func (r *PostgresRoleRepository) Delete(ctx context.Context, id types.ID) error {
	const op = "PostgresRoleRepository.Delete"

	query := `DELETE FROM permissions.roles WHERE id = $1 AND is_system = false`

	result, err := r.db.Exec(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return domain.ErrRoleNotFound(op, id.String())
	}

	return nil
}

// Exists checks if a role name exists within a tenant.
func (r *PostgresRoleRepository) Exists(ctx context.Context, tenantID, name string) (bool, error) {
	const op = "PostgresRoleRepository.Exists"

	query := `SELECT EXISTS(SELECT 1 FROM permissions.roles WHERE (tenant_id = $1 OR tenant_id = '*') AND name = $2)`

	var exists bool
	err := r.db.QueryRow(ctx, query, tenantID, name).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return exists, nil
}

// ============================================================================
// Private Helper Methods
// ============================================================================

// insert creates a new role in the database.
func (r *PostgresRoleRepository) insert(ctx context.Context, role *domain.Role) error {
	const op = "PostgresRoleRepository.insert"

	permissionsJSON, err := r.serializePermissions(role.Permissions())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	query := `
		INSERT INTO permissions.roles (
			id, tenant_id, name, description, permissions, is_system,
			created_at, updated_at, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = r.db.Exec(ctx, query,
		role.ID().String(),
		role.TenantID(),
		role.Name(),
		role.Description(),
		permissionsJSON,
		role.IsSystem(),
		role.CreatedAt().Time(),
		role.UpdatedAt().Time(),
		role.Version(),
	)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// update modifies an existing role in the database.
func (r *PostgresRoleRepository) update(ctx context.Context, role *domain.Role) error {
	const op = "PostgresRoleRepository.update"

	permissionsJSON, err := r.serializePermissions(role.Permissions())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	query := `
		UPDATE permissions.roles SET
			name = $2,
			description = $3,
			permissions = $4,
			updated_at = $5,
			version = $6
		WHERE id = $1 AND version = $6 - 1 AND is_system = false
	`

	result, err := r.db.Exec(ctx, query,
		role.ID().String(),
		role.Name(),
		role.Description(),
		permissionsJSON,
		role.UpdatedAt().Time(),
		role.Version(),
	)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: optimistic lock failure or system role modification attempted", op)
	}

	return nil
}

// buildListQuery constructs a dynamic SELECT query with filters.
func (r *PostgresRoleRepository) buildListQuery(tenantID string, filters *domain.RoleFilters) (string, []interface{}) {
	query := `
		SELECT id, tenant_id, name, description, permissions, is_system,
		       created_at, updated_at, version
		FROM permissions.roles
		WHERE (tenant_id = $1 OR tenant_id = '*')
	`

	args := []interface{}{tenantID}
	argIndex := 2

	if filters != nil {
		if !filters.IncludeSystem {
			query += " AND is_system = false"
		}

		if filters.Search != "" {
			query += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex)
			args = append(args, "%"+filters.Search+"%")
			argIndex++
		}
	}

	query += " ORDER BY is_system DESC, name ASC"

	if filters != nil {
		if filters.Limit > 0 {
			query += fmt.Sprintf(" LIMIT $%d", argIndex)
			args = append(args, filters.Limit)
			argIndex++
		}

		if filters.Offset > 0 {
			query += fmt.Sprintf(" OFFSET $%d", argIndex)
			args = append(args, filters.Offset)
		}
	}

	return query, args
}

// rowToRole converts a database row to a domain Role.
func (r *PostgresRoleRepository) rowToRole(row roleRow) (*domain.Role, error) {
	const op = "PostgresRoleRepository.rowToRole"

	id, err := types.ParseID(row.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid role ID: %w", op, err)
	}

	permissions, err := r.deserializePermissions(row.Permissions)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid permissions: %w", op, err)
	}

	return domain.Reconstitute(
		id,
		row.TenantID,
		row.Name,
		row.Description,
		permissions,
		row.IsSystem,
		types.TimestampFromTime(row.CreatedAt),
		types.TimestampFromTime(row.UpdatedAt),
		row.Version,
	), nil
}

// ============================================================================
// Database Row Type
// ============================================================================

// roleRow represents a row from the roles table.
type roleRow struct {
	ID          string
	TenantID    string
	Name        string
	Description string
	Permissions []byte
	IsSystem    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Version     int
}

// ============================================================================
// JSON Serialization Helpers
// ============================================================================

// serializePermissions converts permissions to JSON.
func (r *PostgresRoleRepository) serializePermissions(permissions domain.PermissionSet) ([]byte, error) {
	return json.Marshal(permissions.Strings())
}

// deserializePermissions converts JSON to permissions.
func (r *PostgresRoleRepository) deserializePermissions(data []byte) (domain.PermissionSet, error) {
	if len(data) == 0 {
		return domain.PermissionSet{}, nil
	}

	var strings []string
	if err := json.Unmarshal(data, &strings); err != nil {
		return nil, err
	}

	return domain.ParsePermissionSet(strings)
}
