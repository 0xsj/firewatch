package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/0xsj/hexagonal-go/pkg/database"
	"github.com/0xsj/hexagonal-go/pkg/types"
	"github.com/0xsj/hexagonal-go/pkg/worker"
)

// Queue is a PostgreSQL-backed implementation of worker.Queue.
// Uses SELECT FOR UPDATE SKIP LOCKED for atomic, concurrent-safe dequeue.
type Queue struct {
	db database.DB
}

// NewQueue creates a new PostgreSQL queue.
func NewQueue(db database.DB) *Queue {
	return &Queue{db: db}
}

// Enqueue adds a job to the queue.
func (q *Queue) Enqueue(ctx context.Context, job *worker.Job) error {
	const op = "postgres.Queue.Enqueue"

	query := `
		INSERT INTO jobs (
			id, type, payload, status, priority, attempts, max_retries,
			last_error, scheduled_at, started_at, completed_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`

	_, err := q.db.Exec(ctx, query,
		job.ID().String(),
		job.Type(),
		job.Payload(),
		job.Status().String(),
		job.Priority(),
		job.Attempts(),
		job.MaxRetries(),
		stringToNullString(job.LastError()),
		job.ScheduledAt(),
		timeToNullTime(job.StartedAt()),
		timeToNullTime(job.CompletedAt()),
		job.CreatedAt(),
		job.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("%s: failed to enqueue job: %w", op, err)
	}

	return nil
}

// Dequeue fetches jobs ready for processing using SELECT FOR UPDATE SKIP LOCKED.
// This ensures atomic dequeue across multiple workers.
func (q *Queue) Dequeue(ctx context.Context, limit int) ([]*worker.Job, error) {
	const op = "postgres.Queue.Dequeue"

	// Use CTE to atomically select and update jobs
	query := `
		WITH selected AS (
			SELECT id
			FROM jobs
			WHERE status IN ('pending', 'retrying')
			  AND scheduled_at <= NOW()
			ORDER BY priority DESC, scheduled_at ASC
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE jobs
		SET status = 'running',
		    started_at = NOW(),
		    attempts = attempts + 1,
		    updated_at = NOW()
		WHERE id IN (SELECT id FROM selected)
		RETURNING id, type, payload, status, priority, attempts, max_retries,
		          last_error, scheduled_at, started_at, completed_at, created_at, updated_at
	`

	rows, err := q.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to dequeue jobs: %w", op, err)
	}
	defer rows.Close()

	var jobs []*worker.Job
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan job: %w", op, err)
		}
		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows iteration error: %w", op, err)
	}

	return jobs, nil
}

// Get retrieves a job by ID.
func (q *Queue) Get(ctx context.Context, id types.ID) (*worker.Job, error) {
	const op = "postgres.Queue.Get"

	query := `
		SELECT id, type, payload, status, priority, attempts, max_retries,
		       last_error, scheduled_at, started_at, completed_at, created_at, updated_at
		FROM jobs
		WHERE id = $1
	`

	row := q.db.QueryRow(ctx, query, id.String())
	job, err := scanJobRow(row)
	if err == sql.ErrNoRows {
		return nil, worker.ErrJobNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get job: %w", op, err)
	}

	return job, nil
}

// Update saves the current state of a job.
func (q *Queue) Update(ctx context.Context, job *worker.Job) error {
	const op = "postgres.Queue.Update"

	query := `
		UPDATE jobs
		SET status = $1,
		    priority = $2,
		    attempts = $3,
		    max_retries = $4,
		    last_error = $5,
		    scheduled_at = $6,
		    started_at = $7,
		    completed_at = $8,
		    updated_at = $9
		WHERE id = $10
	`

	result, err := q.db.Exec(ctx, query,
		job.Status().String(),
		job.Priority(),
		job.Attempts(),
		job.MaxRetries(),
		stringToNullString(job.LastError()),
		job.ScheduledAt(),
		timeToNullTime(job.StartedAt()),
		timeToNullTime(job.CompletedAt()),
		job.UpdatedAt(),
		job.ID().String(),
	)
	if err != nil {
		return fmt.Errorf("%s: failed to update job: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}
	if rowsAffected == 0 {
		return worker.ErrJobNotFound
	}

	return nil
}

// Delete removes a job from the queue.
func (q *Queue) Delete(ctx context.Context, id types.ID) error {
	const op = "postgres.Queue.Delete"

	query := `DELETE FROM jobs WHERE id = $1`

	result, err := q.db.Exec(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("%s: failed to delete job: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}
	if rowsAffected == 0 {
		return worker.ErrJobNotFound
	}

	return nil
}

// Stats returns queue statistics.
func (q *Queue) Stats(ctx context.Context) (*worker.QueueStats, error) {
	const op = "postgres.Queue.Stats"

	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END), 0) as pending,
			COALESCE(SUM(CASE WHEN status = 'running' THEN 1 ELSE 0 END), 0) as running,
			COALESCE(SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END), 0) as completed,
			COALESCE(SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END), 0) as failed,
			COALESCE(SUM(CASE WHEN status = 'retrying' THEN 1 ELSE 0 END), 0) as retrying,
			COALESCE(SUM(CASE WHEN status = 'cancelled' THEN 1 ELSE 0 END), 0) as cancelled,
			COUNT(*) as total
		FROM jobs
	`

	row := q.db.QueryRow(ctx, query)

	var stats worker.QueueStats
	err := row.Scan(
		&stats.Pending,
		&stats.Running,
		&stats.Completed,
		&stats.Failed,
		&stats.Retrying,
		&stats.Cancelled,
		&stats.Total,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get stats: %w", op, err)
	}

	return &stats, nil
}

// Cleanup removes completed/failed jobs older than the given duration.
func (q *Queue) Cleanup(ctx context.Context, olderThan time.Duration) (int64, error) {
	const op = "postgres.Queue.Cleanup"

	query := `
		DELETE FROM jobs
		WHERE status IN ('completed', 'failed', 'cancelled')
		  AND completed_at < $1
	`

	cutoff := time.Now().UTC().Add(-olderThan)
	result, err := q.db.Exec(ctx, query, cutoff)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to cleanup jobs: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}

	return rowsAffected, nil
}

// ============================================================================
// Helpers
// ============================================================================

// scanJob scans a job from rows.
func scanJob(rows interface {
	Scan(dest ...interface{}) error
}) (*worker.Job, error) {
	var (
		id          string
		jobType     string
		payload     []byte
		status      string
		priority    int
		attempts    int
		maxRetries  int
		lastError   sql.NullString
		scheduledAt time.Time
		startedAt   sql.NullTime
		completedAt sql.NullTime
		createdAt   time.Time
		updatedAt   time.Time
	)

	err := rows.Scan(
		&id,
		&jobType,
		&payload,
		&status,
		&priority,
		&attempts,
		&maxRetries,
		&lastError,
		&scheduledAt,
		&startedAt,
		&completedAt,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	jobID, err := types.ParseID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid job id: %w", err)
	}

	return worker.Reconstitute(
		jobID,
		jobType,
		payload,
		worker.Status(status),
		priority,
		attempts,
		maxRetries,
		lastError.String,
		scheduledAt,
		nullTimeToPtr(startedAt),
		nullTimeToPtr(completedAt),
		createdAt,
		updatedAt,
	), nil
}

// scanJobRow scans a job from a single row.
func scanJobRow(row interface {
	Scan(dest ...interface{}) error
}) (*worker.Job, error) {
	return scanJob(row)
}

// stringToNullString converts a string to sql.NullString.
func stringToNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// timeToNullTime converts a *time.Time to sql.NullTime.
func timeToNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

// nullTimeToPtr converts sql.NullTime to *time.Time.
func nullTimeToPtr(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

// ============================================================================
// Additional Query Methods
// ============================================================================

// GetByType retrieves jobs by type with optional status filter.
func (q *Queue) GetByType(ctx context.Context, jobType string, status *worker.Status, limit int) ([]*worker.Job, error) {
	const op = "postgres.Queue.GetByType"

	var query string
	var args []interface{}

	if status != nil {
		query = `
			SELECT id, type, payload, status, priority, attempts, max_retries,
			       last_error, scheduled_at, started_at, completed_at, created_at, updated_at
			FROM jobs
			WHERE type = $1 AND status = $2
			ORDER BY created_at DESC
			LIMIT $3
		`
		args = []interface{}{jobType, status.String(), limit}
	} else {
		query = `
			SELECT id, type, payload, status, priority, attempts, max_retries,
			       last_error, scheduled_at, started_at, completed_at, created_at, updated_at
			FROM jobs
			WHERE type = $1
			ORDER BY created_at DESC
			LIMIT $2
		`
		args = []interface{}{jobType, limit}
	}

	rows, err := q.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to query jobs: %w", op, err)
	}
	defer rows.Close()

	var jobs []*worker.Job
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan job: %w", op, err)
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// RequeueStale requeues jobs that have been running for too long (likely crashed workers).
func (q *Queue) RequeueStale(ctx context.Context, staleDuration time.Duration) (int64, error) {
	const op = "postgres.Queue.RequeueStale"

	query := `
		UPDATE jobs
		SET status = 'retrying',
		    scheduled_at = NOW(),
		    updated_at = NOW()
		WHERE status = 'running'
		  AND started_at < $1
	`

	cutoff := time.Now().UTC().Add(-staleDuration)
	result, err := q.db.Exec(ctx, query, cutoff)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to requeue stale jobs: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}

	return rowsAffected, nil
}
