package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/0xsj/hexagonal-go/internal/audit/domain"
	"github.com/0xsj/hexagonal-go/pkg/database"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// PostgresRepository is a PostgreSQL implementation of domain.Repository.
type PostgresRepository struct {
	db database.DB
}

// NewPostgresRepository creates a new PostgreSQL audit repository.
func NewPostgresRepository(db database.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

// Save persists an audit entry.
func (r *PostgresRepository) Save(ctx context.Context, entry *domain.AuditEntry) error {
	const op = "audit.PostgresRepository.Save"

	// Serialize payload and metadata to JSON
	payload, err := json.Marshal(entry.Payload())
	if err != nil {
		return fmt.Errorf("%s: failed to marshal payload: %w", op, err)
	}

	metadata, err := json.Marshal(entry.Metadata())
	if err != nil {
		return fmt.Errorf("%s: failed to marshal metadata: %w", op, err)
	}

	query := `
		INSERT INTO audit.entries (
			id, event_id, event_type, source, timestamp,
			tenant_id, user_id, correlation_id,
			payload, metadata, created_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11
		)
	`

	_, err = r.db.Exec(ctx, query,
		entry.ID().String(),
		entry.EventID(),
		entry.EventType(),
		entry.Source(),
		entry.Timestamp(),
		nullableString(entry.TenantID()),
		nullableString(entry.UserID()),
		nullableString(entry.CorrelationID()),
		payload,
		metadata,
		entry.CreatedAt(),
	)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// FindByID retrieves an audit entry by ID.
func (r *PostgresRepository) FindByID(ctx context.Context, id types.ID) (*domain.AuditEntry, error) {
	const op = "audit.PostgresRepository.FindByID"

	query := `
		SELECT 
			id, event_id, event_type, source, timestamp,
			tenant_id, user_id, correlation_id,
			payload, metadata, created_at
		FROM audit.entries
		WHERE id = $1
	`

	var row entryRow
	err := r.db.QueryRow(ctx, query, id.String()).Scan(
		&row.ID,
		&row.EventID,
		&row.EventType,
		&row.Source,
		&row.Timestamp,
		&row.TenantID,
		&row.UserID,
		&row.CorrelationID,
		&row.Payload,
		&row.Metadata,
		&row.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return r.rowToEntry(row)
}

// List retrieves audit entries matching the filters.
func (r *PostgresRepository) List(ctx context.Context, filters domain.Filters, page domain.Page) (*domain.PagedResult, error) {
	const op = "audit.PostgresRepository.List"

	// Build query
	query, args := r.buildListQuery(filters, page)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	entries := make([]*domain.AuditEntry, 0)
	for rows.Next() {
		var row entryRow
		err := rows.Scan(
			&row.ID,
			&row.EventID,
			&row.EventType,
			&row.Source,
			&row.Timestamp,
			&row.TenantID,
			&row.UserID,
			&row.CorrelationID,
			&row.Payload,
			&row.Metadata,
			&row.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan row: %w", op, err)
		}

		entry, err := r.rowToEntry(row)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Get total count
	total, err := r.Count(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get count: %w", op, err)
	}

	return &domain.PagedResult{
		Entries: entries,
		Total:   total,
		Limit:   page.Limit,
		Offset:  page.Offset,
	}, nil
}

// Count returns the total number of entries matching the filters.
func (r *PostgresRepository) Count(ctx context.Context, filters domain.Filters) (int, error) {
	const op = "audit.PostgresRepository.Count"

	query, args := r.buildCountQuery(filters)

	var count int
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return count, nil
}

// ============================================================================
// Query Builders
// ============================================================================

func (r *PostgresRepository) buildListQuery(filters domain.Filters, page domain.Page) (string, []interface{}) {
	query := `
		SELECT 
			id, event_id, event_type, source, timestamp,
			tenant_id, user_id, correlation_id,
			payload, metadata, created_at
		FROM audit.entries
		WHERE 1=1
	`

	args := make([]interface{}, 0)
	argIndex := 1

	query, args, argIndex = r.applyFilters(query, args, argIndex, filters)

	// Order by timestamp descending (most recent first)
	query += " ORDER BY timestamp DESC"

	// Pagination
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, page.Limit, page.Offset)

	return query, args
}

func (r *PostgresRepository) buildCountQuery(filters domain.Filters) (string, []interface{}) {
	query := `SELECT COUNT(*) FROM audit.entries WHERE 1=1`

	args := make([]interface{}, 0)
	argIndex := 1

	query, args, _ = r.applyFilters(query, args, argIndex, filters)

	return query, args
}

func (r *PostgresRepository) applyFilters(query string, args []interface{}, argIndex int, filters domain.Filters) (string, []interface{}, int) {
	if filters.TenantID != nil {
		query += fmt.Sprintf(" AND tenant_id = $%d", argIndex)
		args = append(args, *filters.TenantID)
		argIndex++
	}

	if filters.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *filters.UserID)
		argIndex++
	}

	if filters.EventType != nil {
		query += fmt.Sprintf(" AND event_type = $%d", argIndex)
		args = append(args, *filters.EventType)
		argIndex++
	}

	if filters.Source != nil {
		query += fmt.Sprintf(" AND source = $%d", argIndex)
		args = append(args, *filters.Source)
		argIndex++
	}

	if filters.CorrelationID != nil {
		query += fmt.Sprintf(" AND correlation_id = $%d", argIndex)
		args = append(args, *filters.CorrelationID)
		argIndex++
	}

	if filters.FromTimestamp != nil {
		query += fmt.Sprintf(" AND timestamp >= $%d", argIndex)
		args = append(args, filters.FromTimestamp.Time())
		argIndex++
	}

	if filters.ToTimestamp != nil {
		query += fmt.Sprintf(" AND timestamp <= $%d", argIndex)
		args = append(args, filters.ToTimestamp.Time())
		argIndex++
	}

	return query, args, argIndex
}

// ============================================================================
// Row Mapping
// ============================================================================

type entryRow struct {
	ID            string
	EventID       string
	EventType     string
	Source        string
	Timestamp     time.Time
	TenantID      *string
	UserID        *string
	CorrelationID *string
	Payload       []byte
	Metadata      []byte
	CreatedAt     time.Time
}

func (r *PostgresRepository) rowToEntry(row entryRow) (*domain.AuditEntry, error) {
	// Parse payload
	var payload map[string]any
	if err := json.Unmarshal(row.Payload, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Parse metadata
	var metadata map[string]any
	if err := json.Unmarshal(row.Metadata, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	snapshot := domain.Snapshot{
		ID:            row.ID,
		EventID:       row.EventID,
		EventType:     row.EventType,
		Source:        row.Source,
		Timestamp:     row.Timestamp,
		TenantID:      derefString(row.TenantID),
		UserID:        derefString(row.UserID),
		CorrelationID: derefString(row.CorrelationID),
		Payload:       payload,
		Metadata:      metadata,
		CreatedAt:     row.CreatedAt,
	}

	return domain.FromSnapshot(snapshot)
}

// ============================================================================
// Helpers
// ============================================================================

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
