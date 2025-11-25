package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/0xsj/hexagonal-go/internal/email/domain"
	"github.com/0xsj/hexagonal-go/pkg/database"
	pkgerrors "github.com/0xsj/hexagonal-go/pkg/errors"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// PostgresRepository is a PostgreSQL implementation of domain.Repository.
type PostgresRepository struct {
	db database.DB
}

// NewPostgresRepository creates a new PostgreSQL email template repository.
func NewPostgresRepository(db database.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

// Save persists a template (insert or update).
// Save persists a template (insert or update).
func (r *PostgresRepository) Save(ctx context.Context, t *domain.Template) error {
	const op = "PostgresRepository.Save"

	snapshot := t.ToSnapshot()

	variablesJSON, err := json.Marshal(snapshot.Variables)
	if err != nil {
		return fmt.Errorf("%s: failed to marshal variables: %w", op, err)
	}

	query := `
		INSERT INTO email.templates (
			id, tenant_id, slug, locale, name, description,
			subject, body_html, body_text, variables,
			status, version, is_system,
			created_by, updated_by,
			created_at, updated_at, activated_at, archived_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10,
			$11, $12, $13,
			$14, $15,
			$16, $17, $18, $19
		)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			subject = EXCLUDED.subject,
			body_html = EXCLUDED.body_html,
			body_text = EXCLUDED.body_text,
			variables = EXCLUDED.variables,
			status = EXCLUDED.status,
			version = EXCLUDED.version,
			updated_by = EXCLUDED.updated_by,
			updated_at = EXCLUDED.updated_at,
			activated_at = EXCLUDED.activated_at,
			archived_at = EXCLUDED.archived_at
	`

	_, err = r.db.Exec(ctx, query,
		snapshot.ID,
		nullableString(snapshot.TenantID),
		snapshot.Slug,
		snapshot.Locale,
		snapshot.Name,
		nullableString(snapshot.Description),
		snapshot.Subject,
		snapshot.BodyHTML,
		nullableString(snapshot.BodyText),
		variablesJSON,
		snapshot.Status,
		snapshot.Version,
		snapshot.IsSystem,
		nullableString(snapshot.CreatedBy),
		nullableString(snapshot.UpdatedBy),
		snapshot.CreatedAt,
		snapshot.UpdatedAt,
		snapshot.ActivatedAt,
		snapshot.ArchivedAt,
	)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// FindByID retrieves a template by ID.
func (r *PostgresRepository) FindByID(ctx context.Context, id types.ID) (*domain.Template, error) {
	const op = "PostgresRepository.FindByID"

	query := `
		SELECT
			id, tenant_id, slug, locale, name, description,
			subject, body_html, body_text, variables,
			status, version, created_by, updated_by,
			created_at, updated_at
		FROM email.templates
		WHERE id = $1
	`

	template, err := r.scanTemplate(ctx, query, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.NotFound(op, "template")
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return template, nil
}

// FindBySlug retrieves a template by slug, locale, and optional tenant.
func (r *PostgresRepository) FindBySlug(ctx context.Context, tenantID *string, slug string, locale domain.Locale) (*domain.Template, error) {
	const op = "PostgresRepository.FindBySlug"

	// Try tenant-specific first, then fall back to system-wide
	query := `
		SELECT
			id, tenant_id, slug, locale, name, description,
			subject, body_html, body_text, variables,
			status, version, created_by, updated_by,
			created_at, updated_at
		FROM email.templates
		WHERE slug = $1 AND locale = $2
			AND (tenant_id = $3 OR (tenant_id IS NULL AND $3 IS NULL))
		ORDER BY tenant_id NULLS LAST
		LIMIT 1
	`

	template, err := r.scanTemplate(ctx, query, slug, locale.String(), toNullString(tenantID))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrTemplateNotFound(slug, locale)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return template, nil
}

// FindActiveBySlug retrieves an active template by slug and locale.
func (r *PostgresRepository) FindActiveBySlug(ctx context.Context, tenantID *string, slug string, locale domain.Locale) (*domain.Template, error) {
	const op = "PostgresRepository.FindActiveBySlug"

	query := `
		SELECT
			id, tenant_id, slug, locale, name, description,
			subject, body_html, body_text, variables,
			status, version, created_by, updated_by,
			created_at, updated_at
		FROM email.templates
		WHERE slug = $1 AND locale = $2 AND status = 'active'
			AND (tenant_id = $3 OR tenant_id IS NULL)
		ORDER BY tenant_id NULLS LAST
		LIMIT 1
	`

	template, err := r.scanTemplate(ctx, query, slug, locale.String(), toNullString(tenantID))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrTemplateNotFound(slug, locale)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return template, nil
}

// List retrieves templates matching the given filters.
func (r *PostgresRepository) List(ctx context.Context, filters domain.ListFilters) ([]*domain.Template, error) {
	const op = "PostgresRepository.List"

	query, args := r.buildListQuery(filters, false)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	templates := make([]*domain.Template, 0)
	for rows.Next() {
		template, err := r.scanTemplateFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		templates = append(templates, template)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return templates, nil
}

// Count returns the total number of templates matching filters.
func (r *PostgresRepository) Count(ctx context.Context, filters domain.ListFilters) (int, error) {
	const op = "PostgresRepository.Count"

	query, args := r.buildListQuery(filters, true)

	var count int
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return count, nil
}

// Delete removes a template by ID.
func (r *PostgresRepository) Delete(ctx context.Context, id types.ID) error {
	const op = "PostgresRepository.Delete"

	query := `DELETE FROM email.templates WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return pkgerrors.NotFound(op, "template")
	}

	return nil
}

// ExistsBySlug checks if a template with the given slug/locale exists.
func (r *PostgresRepository) ExistsBySlug(ctx context.Context, tenantID *string, slug string, locale domain.Locale) (bool, error) {
	const op = "PostgresRepository.ExistsBySlug"

	query := `
		SELECT EXISTS(
			SELECT 1 FROM email.templates
			WHERE slug = $1 AND locale = $2
				AND (tenant_id = $3 OR (tenant_id IS NULL AND $3 IS NULL))
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, slug, locale.String(), toNullString(tenantID)).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return exists, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

func (r *PostgresRepository) buildListQuery(filters domain.ListFilters, countOnly bool) (string, []interface{}) {
	var sb strings.Builder
	args := make([]interface{}, 0)
	argIndex := 1

	if countOnly {
		sb.WriteString("SELECT COUNT(*) FROM email.templates WHERE 1=1")
	} else {
		sb.WriteString(`
			SELECT
				id, tenant_id, slug, locale, name, description,
				subject, body_html, body_text, variables,
				status, version, created_by, updated_by,
				created_at, updated_at
			FROM email.templates
			WHERE 1=1
		`)
	}

	// Tenant filter
	if filters.TenantID != nil {
		if filters.IncludeSystemTemplates {
			sb.WriteString(fmt.Sprintf(" AND (tenant_id = $%d OR tenant_id IS NULL)", argIndex))
		} else {
			sb.WriteString(fmt.Sprintf(" AND tenant_id = $%d", argIndex))
		}
		args = append(args, *filters.TenantID)
		argIndex++
	} else if !filters.IncludeSystemTemplates {
		sb.WriteString(" AND tenant_id IS NOT NULL")
	}

	// Status filter
	if filters.Status != nil {
		sb.WriteString(fmt.Sprintf(" AND status = $%d", argIndex))
		args = append(args, filters.Status.String())
		argIndex++
	}

	// Locale filter
	if filters.Locale != nil {
		sb.WriteString(fmt.Sprintf(" AND locale = $%d", argIndex))
		args = append(args, filters.Locale.String())
		argIndex++
	}

	// Slug contains filter
	if filters.SlugContains != "" {
		sb.WriteString(fmt.Sprintf(" AND slug ILIKE $%d", argIndex))
		args = append(args, "%"+filters.SlugContains+"%")
		argIndex++
	}

	// Name contains filter
	if filters.NameContains != "" {
		sb.WriteString(fmt.Sprintf(" AND name ILIKE $%d", argIndex))
		args = append(args, "%"+filters.NameContains+"%")
		argIndex++
	}

	if !countOnly {
		// Sorting
		sortBy := "created_at"
		if filters.SortBy != "" {
			sortBy = string(filters.SortBy)
		}
		sortOrder := "DESC"
		if filters.SortOrder == domain.SortOrderAsc {
			sortOrder = "ASC"
		}
		sb.WriteString(fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder))

		// Pagination
		sb.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1))
		args = append(args, filters.Limit, filters.Offset)
	}

	return sb.String(), args
}

func (r *PostgresRepository) scanTemplate(ctx context.Context, query string, args ...interface{}) (*domain.Template, error) {
	row := r.db.QueryRow(ctx, query, args...)
	return r.scanTemplateFromRow(row)
}

func (r *PostgresRepository) scanTemplateFromRow(row interface {
	Scan(dest ...interface{}) error
}) (*domain.Template, error) {
	var (
		id          string
		tenantID    sql.NullString
		slug        string
		locale      string
		name        string
		description string
		subject     string
		bodyHTML    string
		bodyText    string
		variables   []byte
		status      string
		version     int
		createdBy   sql.NullString
		updatedBy   sql.NullString
		createdAt   time.Time
		updatedAt   time.Time
	)

	err := row.Scan(
		&id,
		&tenantID,
		&slug,
		&locale,
		&name,
		&description,
		&subject,
		&bodyHTML,
		&bodyText,
		&variables,
		&status,
		&version,
		&createdBy,
		&updatedBy,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Parse ID
	templateID, err := types.ParseID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid template id: %w", err)
	}

	// Parse locale
	parsedLocale, err := domain.ParseLocale(locale)
	if err != nil {
		return nil, fmt.Errorf("invalid locale: %w", err)
	}

	// Parse status
	parsedStatus, err := domain.ParseStatus(status)
	if err != nil {
		return nil, fmt.Errorf("invalid status: %w", err)
	}

	// Parse variables
	var parsedVariables domain.Variables
	if len(variables) > 0 {
		if err := json.Unmarshal(variables, &parsedVariables); err != nil {
			return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
		}
	}

	// Parse optional IDs
	var parsedCreatedBy *types.ID
	if createdBy.Valid {
		parsed, err := types.ParseID(createdBy.String)
		if err == nil {
			parsedCreatedBy = &parsed
		}
	}

	var parsedUpdatedBy *types.ID
	if updatedBy.Valid {
		parsed, err := types.ParseID(updatedBy.String)
		if err == nil {
			parsedUpdatedBy = &parsed
		}
	}

	// Parse tenant ID
	var parsedTenantID *string
	if tenantID.Valid {
		parsedTenantID = &tenantID.String
	}

	return domain.Reconstitute(
		templateID,
		parsedTenantID,
		slug,
		parsedLocale,
		name,
		description,
		subject,
		bodyHTML,
		bodyText,
		parsedVariables,
		parsedStatus,
		version,
		parsedCreatedBy,
		parsedUpdatedBy,
		types.NewTimestamp(createdAt),
		types.NewTimestamp(updatedAt),
	), nil
}

func toNullString(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}

func toNullID(id *types.ID) interface{} {
	if id == nil {
		return nil
	}
	return id.String()
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
