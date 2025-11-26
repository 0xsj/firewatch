package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	tenant "github.com/0xsj/hexagonal-go/internal/tenant/domain"
	"github.com/0xsj/hexagonal-go/pkg/database"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// PostgresRepository implements tenant.Repository using PostgreSQL.
type PostgresRepository struct {
	db database.DB
}

// NewPostgresRepository creates a new PostgresRepository.
func NewPostgresRepository(db database.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

// Save persists a new tenant.
func (r *PostgresRepository) Save(ctx context.Context, t *tenant.Tenant) error {
	const op = "PostgresRepository.Save"

	settingsJSON, err := json.Marshal(t.Settings().ToMap())
	if err != nil {
		return fmt.Errorf("%s: failed to marshal settings: %w", op, err)
	}

	query := `
		INSERT INTO tenants (
			id, slug, name, plan, status, settings, owner_id,
			billing_id, trial_ends_at, created_at, updated_at, version
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
	`

	_, err = r.db.Exec(ctx, query,
		t.ID().String(),
		t.Slug().String(),
		t.Name(),
		t.Plan().String(),
		t.Status().String(),
		settingsJSON,
		t.OwnerID().String(),
		stringPtrToNullString(t.BillingID()),
		timestampPtrToNullTime(t.TrialEndsAt()),
		t.CreatedAt().Time(),
		t.UpdatedAt().Time(),
		t.Version(),
	)
	if err != nil {
		return fmt.Errorf("%s: failed to insert tenant: %w", op, err)
	}

	return nil
}

// Update persists changes to an existing tenant.
func (r *PostgresRepository) Update(ctx context.Context, t *tenant.Tenant) error {
	const op = "PostgresRepository.Update"

	settingsJSON, err := json.Marshal(t.Settings().ToMap())
	if err != nil {
		return fmt.Errorf("%s: failed to marshal settings: %w", op, err)
	}

	query := `
		UPDATE tenants SET
			name = $1,
			plan = $2,
			status = $3,
			settings = $4,
			billing_id = $5,
			trial_ends_at = $6,
			updated_at = $7,
			version = $8
		WHERE id = $9 AND version = $10
	`

	result, err := r.db.Exec(ctx, query,
		t.Name(),
		t.Plan().String(),
		t.Status().String(),
		settingsJSON,
		stringPtrToNullString(t.BillingID()),
		timestampPtrToNullTime(t.TrialEndsAt()),
		t.UpdatedAt().Time(),
		t.Version(),
		t.ID().String(),
		t.Version()-1, // Check against previous version
	)
	if err != nil {
		return fmt.Errorf("%s: failed to update tenant: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%s: tenant not found or version conflict", op)
	}

	return nil
}

// FindByID retrieves a tenant by its ID.
func (r *PostgresRepository) FindByID(ctx context.Context, id types.ID) (*tenant.Tenant, error) {
	const op = "PostgresRepository.FindByID"

	query := `
		SELECT id, slug, name, plan, status, settings, owner_id,
			   billing_id, trial_ends_at, created_at, updated_at, version
		FROM tenants
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id.String())
	return r.scanTenantFromRow(op, row)
}

// FindBySlug retrieves a tenant by its slug.
func (r *PostgresRepository) FindBySlug(ctx context.Context, slug tenant.Slug) (*tenant.Tenant, error) {
	const op = "PostgresRepository.FindBySlug"

	query := `
		SELECT id, slug, name, plan, status, settings, owner_id,
			   billing_id, trial_ends_at, created_at, updated_at, version
		FROM tenants
		WHERE slug = $1
	`

	row := r.db.QueryRow(ctx, query, slug.String())
	return r.scanTenantFromRow(op, row)
}

// SlugExists checks if a slug is already taken.
func (r *PostgresRepository) SlugExists(ctx context.Context, slug tenant.Slug) (bool, error) {
	const op = "PostgresRepository.SlugExists"

	query := `SELECT EXISTS(SELECT 1 FROM tenants WHERE slug = $1)`

	row := r.db.QueryRow(ctx, query, slug.String())

	var exists bool
	err := row.Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%s: failed to check slug: %w", op, err)
	}

	return exists, nil
}

// List retrieves tenants matching the given filters.
func (r *PostgresRepository) List(ctx context.Context, filters tenant.ListFilters) ([]*tenant.Tenant, error) {
	const op = "PostgresRepository.List"

	query, args := r.buildListQuery(filters, false)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to query tenants: %w", op, err)
	}
	defer rows.Close()

	var tenants []*tenant.Tenant
	for rows.Next() {
		t, err := r.scanTenantFromRows(op, rows)
		if err != nil {
			return nil, err
		}
		tenants = append(tenants, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows iteration error: %w", op, err)
	}

	return tenants, nil
}

// Count returns the total number of tenants matching the filters.
func (r *PostgresRepository) Count(ctx context.Context, filters tenant.ListFilters) (int64, error) {
	const op = "PostgresRepository.Count"

	query, args := r.buildListQuery(filters, true)

	row := r.db.QueryRow(ctx, query, args...)

	var count int64
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to count tenants: %w", op, err)
	}

	return count, nil
}

// Delete removes a tenant from persistence.
func (r *PostgresRepository) Delete(ctx context.Context, id types.ID) error {
	const op = "PostgresRepository.Delete"

	query := `DELETE FROM tenants WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("%s: failed to delete tenant: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}
	if rowsAffected == 0 {
		return tenant.ErrTenantNotFound(op, id.String())
	}

	return nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// buildListQuery constructs the SQL query for listing tenants.
func (r *PostgresRepository) buildListQuery(filters tenant.ListFilters, countOnly bool) (string, []any) {
	var query strings.Builder
	var args []any
	argIndex := 1

	if countOnly {
		query.WriteString("SELECT COUNT(*) FROM tenants WHERE 1=1")
	} else {
		query.WriteString(`
			SELECT id, slug, name, plan, status, settings, owner_id,
				   billing_id, trial_ends_at, created_at, updated_at, version
			FROM tenants
			WHERE 1=1
		`)
	}

	// Apply filters
	if filters.Status != nil {
		query.WriteString(fmt.Sprintf(" AND status = $%d", argIndex))
		args = append(args, filters.Status.String())
		argIndex++
	}

	if filters.Plan != nil {
		query.WriteString(fmt.Sprintf(" AND plan = $%d", argIndex))
		args = append(args, filters.Plan.String())
		argIndex++
	}

	if filters.OwnerID != nil {
		query.WriteString(fmt.Sprintf(" AND owner_id = $%d", argIndex))
		args = append(args, filters.OwnerID.String())
		argIndex++
	}

	if filters.Search != nil && *filters.Search != "" {
		query.WriteString(fmt.Sprintf(" AND (name ILIKE $%d OR slug ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+*filters.Search+"%")
		argIndex++
	}

	if !countOnly {
		query.WriteString(" ORDER BY created_at DESC")
		query.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1))
		args = append(args, filters.Limit, filters.Offset)
	}

	return query.String(), args
}

// scanTenantFromRow scans a single tenant from a database row.
func (r *PostgresRepository) scanTenantFromRow(op string, row interface {
	Scan(dest ...interface{}) error
}) (*tenant.Tenant, error) {
	var (
		id          string
		slug        string
		name        string
		plan        string
		status      string
		settings    []byte
		ownerID     string
		billingID   sql.NullString
		trialEndsAt sql.NullTime
		createdAt   sql.NullTime
		updatedAt   sql.NullTime
		version     int
	)

	err := row.Scan(
		&id,
		&slug,
		&name,
		&plan,
		&status,
		&settings,
		&ownerID,
		&billingID,
		&trialEndsAt,
		&createdAt,
		&updatedAt,
		&version,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("%s: failed to scan tenant: %w", op, err)
	}

	return r.rowToTenant(op, id, slug, name, plan, status, settings, ownerID, billingID, trialEndsAt, createdAt, updatedAt, version)
}

// scanTenantFromRows scans a tenant from multiple rows.
func (r *PostgresRepository) scanTenantFromRows(op string, rows interface {
	Scan(dest ...interface{}) error
}) (*tenant.Tenant, error) {
	var (
		id          string
		slug        string
		name        string
		plan        string
		status      string
		settings    []byte
		ownerID     string
		billingID   sql.NullString
		trialEndsAt sql.NullTime
		createdAt   sql.NullTime
		updatedAt   sql.NullTime
		version     int
	)

	err := rows.Scan(
		&id,
		&slug,
		&name,
		&plan,
		&status,
		&settings,
		&ownerID,
		&billingID,
		&trialEndsAt,
		&createdAt,
		&updatedAt,
		&version,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to scan tenant: %w", op, err)
	}

	return r.rowToTenant(op, id, slug, name, plan, status, settings, ownerID, billingID, trialEndsAt, createdAt, updatedAt, version)
}

// rowToTenant converts scanned values to a domain tenant.
func (r *PostgresRepository) rowToTenant(
	op string,
	id, slugStr, name, planStr, statusStr string,
	settingsJSON []byte,
	ownerIDStr string,
	billingID sql.NullString,
	trialEndsAt, createdAt, updatedAt sql.NullTime,
	version int,
) (*tenant.Tenant, error) {
	// Parse ID
	tenantID, err := types.ParseID(id)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid tenant id: %w", op, err)
	}

	// Parse slug
	slug, err := tenant.NewSlug(slugStr)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid slug: %w", op, err)
	}

	// Parse plan
	plan, err := tenant.ParsePlan(planStr)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid plan: %w", op, err)
	}

	// Parse status
	status, err := tenant.ParseStatus(statusStr)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid status: %w", op, err)
	}

	// Parse settings
	var settingsMap map[string]any
	if len(settingsJSON) > 0 {
		if err := json.Unmarshal(settingsJSON, &settingsMap); err != nil {
			return nil, fmt.Errorf("%s: failed to unmarshal settings: %w", op, err)
		}
	}
	settings := tenant.NewSettingsFromMap(settingsMap)

	// Parse owner ID
	ownerID, err := types.ParseID(ownerIDStr)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid owner id: %w", op, err)
	}

	// Parse optional billing ID
	var billingIDPtr *string
	if billingID.Valid {
		billingIDPtr = &billingID.String
	}

	// Parse optional trial ends at
	var trialEndsAtPtr *types.Timestamp
	if trialEndsAt.Valid {
		ts := types.NewTimestamp(trialEndsAt.Time)
		trialEndsAtPtr = &ts
	}

	// Parse timestamps
	createdAtTs := types.NewTimestamp(createdAt.Time)
	updatedAtTs := types.NewTimestamp(updatedAt.Time)

	// Reconstitute tenant
	return tenant.Reconstitute(
		tenantID,
		slug,
		name,
		plan,
		status,
		settings,
		ownerID,
		billingIDPtr,
		trialEndsAtPtr,
		createdAtTs,
		updatedAtTs,
		version,
	), nil
}

// stringPtrToNullString converts a *string to sql.NullString.
func stringPtrToNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

// timestampPtrToNullTime converts a *types.Timestamp to sql.NullTime.
func timestampPtrToNullTime(ts *types.Timestamp) sql.NullTime {
	if ts == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: ts.Time(), Valid: true}
}
