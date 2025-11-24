package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/0xsj/hexagonal-go/internal/notifications/domain"
	"github.com/0xsj/hexagonal-go/pkg/database"
	pkgerrors "github.com/0xsj/hexagonal-go/pkg/errors"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// PostgresRepository is a PostgreSQL implementation of domain.Repository.
type PostgresRepository struct {
	db database.DB
}

// NewPostgresRepository creates a new PostgreSQL notification repository.
func NewPostgresRepository(db database.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

// Save persists a notification (insert or update).
func (r *PostgresRepository) Save(ctx context.Context, notification *domain.Notification) error {
	const op = "notifications.PostgresRepository.Save"

	query := `
		INSERT INTO notifications.notifications (
			id, tenant_id, channel, recipient, subject, body,
			status, attempts, last_error, sent_at,
			user_id, correlation_id, event_type,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10,
			$11, $12, $13,
			$14, $15
		)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			attempts = EXCLUDED.attempts,
			last_error = EXCLUDED.last_error,
			sent_at = EXCLUDED.sent_at,
			updated_at = EXCLUDED.updated_at
	`

	snapshot := notification.ToSnapshot()

	_, err := r.db.Exec(ctx, query,
		snapshot.ID,
		nullableString(snapshot.TenantID),
		snapshot.Channel,
		snapshot.Recipient,
		snapshot.Subject,
		snapshot.Body,
		snapshot.Status,
		snapshot.Attempts,
		nullableString(snapshot.LastError),
		snapshot.SentAt,
		nullableString(snapshot.UserID),
		nullableString(snapshot.CorrelationID),
		nullableString(snapshot.EventType),
		snapshot.CreatedAt,
		snapshot.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// FindByID retrieves a notification by ID.
func (r *PostgresRepository) FindByID(ctx context.Context, id types.ID) (*domain.Notification, error) {
	const op = "notifications.PostgresRepository.FindByID"

	query := `
		SELECT
			id, tenant_id, channel, recipient, subject, body,
			status, attempts, last_error, sent_at,
			user_id, correlation_id, event_type,
			created_at, updated_at
		FROM notifications.notifications
		WHERE id = $1
	`

	var row notificationRow
	err := r.db.QueryRow(ctx, query, id.String()).Scan(
		&row.ID,
		&row.TenantID,
		&row.Channel,
		&row.Recipient,
		&row.Subject,
		&row.Body,
		&row.Status,
		&row.Attempts,
		&row.LastError,
		&row.SentAt,
		&row.UserID,
		&row.CorrelationID,
		&row.EventType,
		&row.CreatedAt,
		&row.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, pkgerrors.NotFound(op, "notification")
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return r.rowToNotification(row)
}

// FindPending retrieves pending notifications for retry processing.
func (r *PostgresRepository) FindPending(ctx context.Context, limit int) ([]*domain.Notification, error) {
	const op = "notifications.PostgresRepository.FindPending"

	query := `
		SELECT
			id, tenant_id, channel, recipient, subject, body,
			status, attempts, last_error, sent_at,
			user_id, correlation_id, event_type,
			created_at, updated_at
		FROM notifications.notifications
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, domain.StatusPending.String(), limit)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	return r.scanRows(rows, op)
}

// List retrieves notifications matching the filters.
func (r *PostgresRepository) List(ctx context.Context, filters domain.Filters, page domain.Page) (*domain.PagedResult, error) {
	const op = "notifications.PostgresRepository.List"

	query, args := r.buildListQuery(filters, page)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	notifications, err := r.scanRows(rows, op)
	if err != nil {
		return nil, err
	}

	total, err := r.Count(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get count: %w", op, err)
	}

	return &domain.PagedResult{
		Notifications: notifications,
		Total:         total,
		Limit:         page.Limit,
		Offset:        page.Offset,
	}, nil
}

// Count returns the total number of notifications matching the filters.
func (r *PostgresRepository) Count(ctx context.Context, filters domain.Filters) (int, error) {
	const op = "notifications.PostgresRepository.Count"

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
			id, tenant_id, channel, recipient, subject, body,
			status, attempts, last_error, sent_at,
			user_id, correlation_id, event_type,
			created_at, updated_at
		FROM notifications.notifications
		WHERE 1=1
	`

	args := make([]interface{}, 0)
	argIndex := 1

	query, args, argIndex = r.applyFilters(query, args, argIndex, filters)

	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, page.Limit, page.Offset)

	return query, args
}

func (r *PostgresRepository) buildCountQuery(filters domain.Filters) (string, []interface{}) {
	query := `SELECT COUNT(*) FROM notifications.notifications WHERE 1=1`

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

	if filters.Channel != nil {
		query += fmt.Sprintf(" AND channel = $%d", argIndex)
		args = append(args, filters.Channel.String())
		argIndex++
	}

	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, filters.Status.String())
		argIndex++
	}

	if filters.EventType != nil {
		query += fmt.Sprintf(" AND event_type = $%d", argIndex)
		args = append(args, *filters.EventType)
		argIndex++
	}

	if filters.Recipient != nil {
		query += fmt.Sprintf(" AND recipient = $%d", argIndex)
		args = append(args, *filters.Recipient)
		argIndex++
	}

	return query, args, argIndex
}

// ============================================================================
// Row Mapping
// ============================================================================

type notificationRow struct {
	ID            string
	TenantID      *string
	Channel       string
	Recipient     string
	Subject       string
	Body          string
	Status        string
	Attempts      int
	LastError     *string
	SentAt        *time.Time
	UserID        *string
	CorrelationID *string
	EventType     *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (r *PostgresRepository) rowToNotification(row notificationRow) (*domain.Notification, error) {
	snapshot := domain.Snapshot{
		ID:            row.ID,
		TenantID:      derefString(row.TenantID),
		Channel:       row.Channel,
		Recipient:     row.Recipient,
		Subject:       row.Subject,
		Body:          row.Body,
		Status:        row.Status,
		Attempts:      row.Attempts,
		LastError:     derefString(row.LastError),
		SentAt:        row.SentAt,
		UserID:        derefString(row.UserID),
		CorrelationID: derefString(row.CorrelationID),
		EventType:     derefString(row.EventType),
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}

	return domain.FromSnapshot(snapshot)
}

func (r *PostgresRepository) scanRows(rows *sql.Rows, op string) ([]*domain.Notification, error) {
	notifications := make([]*domain.Notification, 0)

	for rows.Next() {
		var row notificationRow
		err := rows.Scan(
			&row.ID,
			&row.TenantID,
			&row.Channel,
			&row.Recipient,
			&row.Subject,
			&row.Body,
			&row.Status,
			&row.Attempts,
			&row.LastError,
			&row.SentAt,
			&row.UserID,
			&row.CorrelationID,
			&row.EventType,
			&row.CreatedAt,
			&row.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan row: %w", op, err)
		}

		notification, err := r.rowToNotification(row)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return notifications, nil
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
