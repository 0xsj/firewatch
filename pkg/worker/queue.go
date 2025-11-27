package worker

import (
	"context"
	"time"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Queue defines the interface for job queue operations.
type Queue interface {
	// Enqueue adds a job to the queue.
	Enqueue(ctx context.Context, job *Job) error

	// Dequeue fetches jobs ready for processing.
	// Returns up to limit jobs with scheduledAt <= now and processable status.
	// Jobs are atomically marked as running.
	Dequeue(ctx context.Context, limit int) ([]*Job, error)

	// Get retrieves a job by ID.
	Get(ctx context.Context, id types.ID) (*Job, error)

	// Update saves the current state of a job.
	Update(ctx context.Context, job *Job) error

	// Delete removes a job from the queue.
	Delete(ctx context.Context, id types.ID) error

	// Stats returns queue statistics.
	Stats(ctx context.Context) (*QueueStats, error)

	// Cleanup removes completed/failed jobs older than the given duration.
	Cleanup(ctx context.Context, olderThan time.Duration) (int64, error)
}

// QueueStats contains queue statistics.
type QueueStats struct {
	Pending   int64 `json:"pending"`
	Running   int64 `json:"running"`
	Completed int64 `json:"completed"`
	Failed    int64 `json:"failed"`
	Retrying  int64 `json:"retrying"`
	Cancelled int64 `json:"cancelled"`
	Total     int64 `json:"total"`
}
