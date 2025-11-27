package memory

import (
	"context"
	"sync"
	"time"

	"github.com/0xsj/hexagonal-go/pkg/types"
	"github.com/0xsj/hexagonal-go/pkg/worker"
)

// Queue is an in-memory implementation of worker.Queue.
// Useful for testing and development. Not suitable for production
// as jobs are lost on restart.
type Queue struct {
	mu   sync.RWMutex
	jobs map[types.ID]*worker.Job
}

// NewQueue creates a new in-memory queue.
func NewQueue() *Queue {
	return &Queue{
		jobs: make(map[types.ID]*worker.Job),
	}
}

// Enqueue adds a job to the queue.
func (q *Queue) Enqueue(ctx context.Context, job *worker.Job) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, exists := q.jobs[job.ID()]; exists {
		return worker.ErrJobAlreadyExists
	}

	q.jobs[job.ID()] = job
	return nil
}

// Dequeue fetches jobs ready for processing.
func (q *Queue) Dequeue(ctx context.Context, limit int) ([]*worker.Job, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now().UTC()
	var ready []*worker.Job

	for _, job := range q.jobs {
		if len(ready) >= limit {
			break
		}

		// Check if job is processable and scheduled time has passed
		if job.Status().IsProcessable() && !job.ScheduledAt().After(now) {
			if err := job.MarkRunning(); err == nil {
				ready = append(ready, job)
			}
		}
	}

	return ready, nil
}

// Get retrieves a job by ID.
func (q *Queue) Get(ctx context.Context, id types.ID) (*worker.Job, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	job, exists := q.jobs[id]
	if !exists {
		return nil, worker.ErrJobNotFound
	}

	return job, nil
}

// Update saves the current state of a job.
func (q *Queue) Update(ctx context.Context, job *worker.Job) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, exists := q.jobs[job.ID()]; !exists {
		return worker.ErrJobNotFound
	}

	q.jobs[job.ID()] = job
	return nil
}

// Delete removes a job from the queue.
func (q *Queue) Delete(ctx context.Context, id types.ID) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, exists := q.jobs[id]; !exists {
		return worker.ErrJobNotFound
	}

	delete(q.jobs, id)
	return nil
}

// Stats returns queue statistics.
func (q *Queue) Stats(ctx context.Context) (*worker.QueueStats, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	stats := &worker.QueueStats{}

	for _, job := range q.jobs {
		switch job.Status() {
		case worker.StatusPending:
			stats.Pending++
		case worker.StatusRunning:
			stats.Running++
		case worker.StatusCompleted:
			stats.Completed++
		case worker.StatusFailed:
			stats.Failed++
		case worker.StatusRetrying:
			stats.Retrying++
		case worker.StatusCancelled:
			stats.Cancelled++
		}
		stats.Total++
	}

	return stats, nil
}

// Cleanup removes completed/failed jobs older than the given duration.
func (q *Queue) Cleanup(ctx context.Context, olderThan time.Duration) (int64, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	cutoff := time.Now().UTC().Add(-olderThan)
	var removed int64

	for id, job := range q.jobs {
		if job.Status().IsTerminal() && job.CompletedAt() != nil && job.CompletedAt().Before(cutoff) {
			delete(q.jobs, id)
			removed++
		}
	}

	return removed, nil
}

// Size returns the total number of jobs in the queue.
func (q *Queue) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.jobs)
}

// Clear removes all jobs from the queue.
func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.jobs = make(map[types.ID]*worker.Job)
}
