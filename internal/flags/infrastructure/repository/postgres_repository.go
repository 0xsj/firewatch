package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/0xsj/hexagonal-go/internal/flags/domain"
	"github.com/0xsj/hexagonal-go/pkg/database"
	pkgerrors "github.com/0xsj/hexagonal-go/pkg/errors"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// PostgresRepository is a PostgreSQL implementation of domain.Repository.
type PostgresRepository struct {
	db database.DB
}

// NewPostgresRepository creates a new PostgresRepository.
func NewPostgresRepository(db database.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

// ============================================================================
// Repository Implementation
// ============================================================================

// Save persists a flag (create or update).
func (r *PostgresRepository) Save(ctx context.Context, flag *domain.Flag) error {
	const op = "PostgresRepository.Save"

	// Check if flag already exists
	existing, err := r.FindByID(ctx, flag.ID())
	if err != nil && !pkgerrors.IsNotFound(err) {
		return fmt.Errorf("%s: failed to check existing flag: %w", op, err)
	}

	if existing != nil {
		return r.update(ctx, flag)
	}

	return r.insert(ctx, flag)
}

// FindByID retrieves a flag by its ID.
func (r *PostgresRepository) FindByID(ctx context.Context, id types.ID) (*domain.Flag, error) {
	const op = "PostgresRepository.FindByID"

	query := `
		SELECT id, tenant_id, key, name, description, enabled,
		       variants, default_variant, rules, overrides,
		       created_at, updated_at, version
		FROM flags.flags
		WHERE id = $1
	`

	var row flagRow
	err := r.db.QueryRow(ctx, query, id.String()).Scan(
		&row.ID,
		&row.TenantID,
		&row.Key,
		&row.Name,
		&row.Description,
		&row.Enabled,
		&row.Variants,
		&row.DefaultVariant,
		&row.Rules,
		&row.Overrides,
		&row.CreatedAt,
		&row.UpdatedAt,
		&row.Version,
	)

	if err == sql.ErrNoRows {
		return nil, pkgerrors.NotFound(op, fmt.Sprintf("flag not found: %s", id.String()))
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return r.rowToFlag(row)
}

// FindByKey retrieves a flag by its key within a tenant.
func (r *PostgresRepository) FindByKey(ctx context.Context, tenantID, key string) (*domain.Flag, error) {
	const op = "PostgresRepository.FindByKey"

	query := `
		SELECT id, tenant_id, key, name, description, enabled,
		       variants, default_variant, rules, overrides,
		       created_at, updated_at, version
		FROM flags.flags
		WHERE tenant_id = $1 AND key = $2
	`

	var row flagRow
	err := r.db.QueryRow(ctx, query, tenantID, key).Scan(
		&row.ID,
		&row.TenantID,
		&row.Key,
		&row.Name,
		&row.Description,
		&row.Enabled,
		&row.Variants,
		&row.DefaultVariant,
		&row.Rules,
		&row.Overrides,
		&row.CreatedAt,
		&row.UpdatedAt,
		&row.Version,
	)

	if err == sql.ErrNoRows {
		return nil, pkgerrors.NotFound(op, fmt.Sprintf("flag not found: key=%s", key))
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return r.rowToFlag(row)
}

// FindAll retrieves all flags for a tenant with optional filters.
func (r *PostgresRepository) FindAll(ctx context.Context, tenantID string, filters *domain.Filters) ([]*domain.Flag, error) {
	const op = "PostgresRepository.FindAll"

	query, args := r.buildListQuery(tenantID, filters)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	flags := make([]*domain.Flag, 0)
	for rows.Next() {
		var row flagRow
		err := rows.Scan(
			&row.ID,
			&row.TenantID,
			&row.Key,
			&row.Name,
			&row.Description,
			&row.Enabled,
			&row.Variants,
			&row.DefaultVariant,
			&row.Rules,
			&row.Overrides,
			&row.CreatedAt,
			&row.UpdatedAt,
			&row.Version,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan row: %w", op, err)
		}

		flag, err := r.rowToFlag(row)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		flags = append(flags, flag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return flags, nil
}

// Delete removes a flag by ID.
func (r *PostgresRepository) Delete(ctx context.Context, id types.ID) error {
	const op = "PostgresRepository.Delete"

	query := `DELETE FROM flags.flags WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return pkgerrors.NotFound(op, fmt.Sprintf("flag not found: %s", id.String()))
	}

	return nil
}

// Exists checks if a flag key exists within a tenant.
func (r *PostgresRepository) Exists(ctx context.Context, tenantID, key string) (bool, error) {
	const op = "PostgresRepository.Exists"

	query := `SELECT EXISTS(SELECT 1 FROM flags.flags WHERE tenant_id = $1 AND key = $2)`

	var exists bool
	err := r.db.QueryRow(ctx, query, tenantID, key).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return exists, nil
}

// ============================================================================
// Private Helper Methods
// ============================================================================

// insert creates a new flag in the database.
func (r *PostgresRepository) insert(ctx context.Context, flag *domain.Flag) error {
	const op = "PostgresRepository.insert"

	variantsJSON, err := r.serializeVariants(flag.Variants())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rulesJSON, err := r.serializeRules(flag.Rules())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	overridesJSON, err := r.serializeOverrides(flag.Overrides())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	query := `
		INSERT INTO flags.flags (
			id, tenant_id, key, name, description, enabled,
			variants, default_variant, rules, overrides,
			created_at, updated_at, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err = r.db.Exec(ctx, query,
		flag.ID().String(),
		flag.TenantID(),
		flag.Key(),
		flag.Name(),
		flag.Description(),
		flag.Enabled(),
		variantsJSON,
		flag.DefaultVariant(),
		rulesJSON,
		overridesJSON,
		flag.CreatedAt().Time(),
		flag.UpdatedAt().Time(),
		flag.Version(),
	)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// update modifies an existing flag in the database.
func (r *PostgresRepository) update(ctx context.Context, flag *domain.Flag) error {
	const op = "PostgresRepository.update"

	variantsJSON, err := r.serializeVariants(flag.Variants())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rulesJSON, err := r.serializeRules(flag.Rules())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	overridesJSON, err := r.serializeOverrides(flag.Overrides())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	query := `
		UPDATE flags.flags SET
			name = $2,
			description = $3,
			enabled = $4,
			variants = $5,
			default_variant = $6,
			rules = $7,
			overrides = $8,
			updated_at = $9,
			version = $10
		WHERE id = $1 AND version = $10 - 1
	`

	result, err := r.db.Exec(ctx, query,
		flag.ID().String(),
		flag.Name(),
		flag.Description(),
		flag.Enabled(),
		variantsJSON,
		flag.DefaultVariant(),
		rulesJSON,
		overridesJSON,
		flag.UpdatedAt().Time(),
		flag.Version(),
	)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: optimistic lock failure, flag was modified by another process", op)
	}

	return nil
}

// buildListQuery constructs a dynamic SELECT query with filters.
func (r *PostgresRepository) buildListQuery(tenantID string, filters *domain.Filters) (string, []interface{}) {
	query := `
		SELECT id, tenant_id, key, name, description, enabled,
		       variants, default_variant, rules, overrides,
		       created_at, updated_at, version
		FROM flags.flags
		WHERE tenant_id = $1
	`

	args := []interface{}{tenantID}
	argIndex := 2

	if filters != nil {
		if filters.Enabled != nil {
			query += fmt.Sprintf(" AND enabled = $%d", argIndex)
			args = append(args, *filters.Enabled)
			argIndex++
		}

		if len(filters.Keys) > 0 {
			query += fmt.Sprintf(" AND key = ANY($%d)", argIndex)
			args = append(args, filters.Keys)
			argIndex++
		}

		if filters.Search != "" {
			query += fmt.Sprintf(" AND (key ILIKE $%d OR name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex, argIndex)
			args = append(args, "%"+filters.Search+"%")
			argIndex++
		}
	}

	query += " ORDER BY created_at DESC"

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

// rowToFlag converts a database row to a domain Flag.
func (r *PostgresRepository) rowToFlag(row flagRow) (*domain.Flag, error) {
	const op = "PostgresRepository.rowToFlag"

	id, err := types.ParseID(row.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid flag ID: %w", op, err)
	}

	variants, err := r.deserializeVariants(row.Variants)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid variants: %w", op, err)
	}

	rules, err := r.deserializeRules(row.Rules)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid rules: %w", op, err)
	}

	overrides, err := r.deserializeOverrides(row.Overrides)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid overrides: %w", op, err)
	}

	return domain.Reconstitute(
		id,
		row.TenantID,
		row.Key,
		row.Name,
		row.Description,
		row.Enabled,
		variants,
		row.DefaultVariant,
		rules,
		overrides,
		types.TimestampFromTime(row.CreatedAt),
		types.TimestampFromTime(row.UpdatedAt),
		row.Version,
	), nil
}

// ============================================================================
// Database Row Type
// ============================================================================

// flagRow represents a row from the flags table.
type flagRow struct {
	ID             string
	TenantID       string
	Key            string
	Name           string
	Description    string
	Enabled        bool
	Variants       []byte
	DefaultVariant string
	Rules          []byte
	Overrides      []byte
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Version        int
}

// ============================================================================
// JSON Serialization Helpers
// ============================================================================

// variantJSON represents a variant in JSON format.
type variantJSON struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Weight      int    `json:"weight"`
	Description string `json:"description"`
}

// ruleJSON represents a rule in JSON format.
type ruleJSON struct {
	ID         string   `json:"id"`
	Type       string   `json:"type"`
	Attribute  string   `json:"attribute"`
	Operator   string   `json:"operator"`
	Values     []string `json:"values"`
	Percentage int      `json:"percentage"`
	VariantKey string   `json:"variant_key"`
	Priority   int      `json:"priority"`
}

// overrideJSON represents an override in JSON format.
type overrideJSON struct {
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	VariantKey string `json:"variant_key"`
}

// serializeVariants converts variants to JSON.
func (r *PostgresRepository) serializeVariants(variants []domain.Variant) ([]byte, error) {
	vjs := make([]variantJSON, len(variants))
	for i, v := range variants {
		vjs[i] = variantJSON{
			Key:         v.Key(),
			Value:       v.Value(),
			Weight:      v.Weight(),
			Description: v.Description(),
		}
	}
	return json.Marshal(vjs)
}

// deserializeVariants converts JSON to variants.
func (r *PostgresRepository) deserializeVariants(data []byte) ([]domain.Variant, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var vjs []variantJSON
	if err := json.Unmarshal(data, &vjs); err != nil {
		return nil, err
	}

	variants := make([]domain.Variant, len(vjs))
	for i, vj := range vjs {
		v, err := domain.NewVariant(vj.Key, vj.Value, vj.Weight, vj.Description)
		if err != nil {
			return nil, err
		}
		variants[i] = v
	}

	return variants, nil
}

// serializeRules converts rules to JSON.
func (r *PostgresRepository) serializeRules(rules []domain.Rule) ([]byte, error) {
	rjs := make([]ruleJSON, len(rules))
	for i, rule := range rules {
		rjs[i] = ruleJSON{
			ID:         rule.ID().String(),
			Type:       rule.Type().String(),
			Attribute:  rule.Attribute(),
			Operator:   rule.Operator().String(),
			Values:     rule.Values(),
			Percentage: rule.Percentage(),
			VariantKey: rule.VariantKey(),
			Priority:   rule.Priority(),
		}
	}
	return json.Marshal(rjs)
}

// deserializeRules converts JSON to rules.
func (r *PostgresRepository) deserializeRules(data []byte) ([]domain.Rule, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var rjs []ruleJSON
	if err := json.Unmarshal(data, &rjs); err != nil {
		return nil, err
	}

	rules := make([]domain.Rule, len(rjs))
	for i, rj := range rjs {
		id, err := types.ParseID(rj.ID)
		if err != nil {
			return nil, err
		}

		rule, err := domain.NewRule(
			id,
			domain.RuleType(rj.Type),
			rj.Attribute,
			domain.Operator(rj.Operator),
			rj.Values,
			rj.Percentage,
			rj.VariantKey,
			rj.Priority,
		)
		if err != nil {
			return nil, err
		}
		rules[i] = rule
	}

	return rules, nil
}

// serializeOverrides converts overrides to JSON.
func (r *PostgresRepository) serializeOverrides(overrides map[string]domain.Override) ([]byte, error) {
	ojs := make([]overrideJSON, 0, len(overrides))
	for _, o := range overrides {
		ojs = append(ojs, overrideJSON{
			TargetType: o.TargetType,
			TargetID:   o.TargetID,
			VariantKey: o.VariantKey,
		})
	}
	return json.Marshal(ojs)
}

// deserializeOverrides converts JSON to overrides.
func (r *PostgresRepository) deserializeOverrides(data []byte) (map[string]domain.Override, error) {
	if len(data) == 0 {
		return make(map[string]domain.Override), nil
	}

	var ojs []overrideJSON
	if err := json.Unmarshal(data, &ojs); err != nil {
		return nil, err
	}

	overrides := make(map[string]domain.Override, len(ojs))
	for _, oj := range ojs {
		key := oj.TargetType + ":" + oj.TargetID
		overrides[key] = domain.Override{
			TargetType: oj.TargetType,
			TargetID:   oj.TargetID,
			VariantKey: oj.VariantKey,
		}
	}

	return overrides, nil
}
