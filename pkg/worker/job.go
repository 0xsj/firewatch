package worker

import (
	"encoding/json"
	"time"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Job represents a unit of work to be processed.
type Job struct {
	id          types.ID
	jobType     string
	payload     []byte
	status      Status
	priority    int
	attempts    int
	maxRetries  int
	lastError   string
	scheduledAt time.Time
	startedAt   *time.Time
	completedAt *time.Time
	createdAt   time.Time
	updatedAt   time.Time
}

// NewJob creates a new job with the given type and payload.
func NewJob(jobType string, payload []byte) *Job {
	now := time.Now().UTC()
	return &Job{
		id:          types.NewID(),
		jobType:     jobType,
		payload:     payload,
		status:      StatusPending,
		priority:    0,
		attempts:    0,
		maxRetries:  3,
		scheduledAt: now,
		createdAt:   now,
		updatedAt:   now,
	}
}

// NewJobWithData creates a new job, marshaling the data to JSON.
func NewJobWithData(jobType string, data any) (*Job, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return NewJob(jobType, payload), nil
}

// Reconstitute recreates a job from persistence.
func Reconstitute(
	id types.ID,
	jobType string,
	payload []byte,
	status Status,
	priority int,
	attempts int,
	maxRetries int,
	lastError string,
	scheduledAt time.Time,
	startedAt *time.Time,
	completedAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) *Job {
	return &Job{
		id:          id,
		jobType:     jobType,
		payload:     payload,
		status:      status,
		priority:    priority,
		attempts:    attempts,
		maxRetries:  maxRetries,
		lastError:   lastError,
		scheduledAt: scheduledAt,
		startedAt:   startedAt,
		completedAt: completedAt,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

// ============================================================================
// Getters
// ============================================================================

// ID returns the job's unique identifier.
func (j *Job) ID() types.ID {
	return j.id
}

// Type returns the job type.
func (j *Job) Type() string {
	return j.jobType
}

// Payload returns the raw job payload.
func (j *Job) Payload() []byte {
	return j.payload
}

// UnmarshalPayload unmarshals the payload into the given target.
func (j *Job) UnmarshalPayload(target any) error {
	return json.Unmarshal(j.payload, target)
}

// Status returns the current job status.
func (j *Job) Status() Status {
	return j.status
}

// Priority returns the job priority (higher = more important).
func (j *Job) Priority() int {
	return j.priority
}

// Attempts returns the number of processing attempts.
func (j *Job) Attempts() int {
	return j.attempts
}

// MaxRetries returns the maximum number of retry attempts.
func (j *Job) MaxRetries() int {
	return j.maxRetries
}

// LastError returns the last error message if the job failed.
func (j *Job) LastError() string {
	return j.lastError
}

// ScheduledAt returns when the job is scheduled to run.
func (j *Job) ScheduledAt() time.Time {
	return j.scheduledAt
}

// StartedAt returns when the job started processing.
func (j *Job) StartedAt() *time.Time {
	return j.startedAt
}

// CompletedAt returns when the job completed.
func (j *Job) CompletedAt() *time.Time {
	return j.completedAt
}

// CreatedAt returns when the job was created.
func (j *Job) CreatedAt() time.Time {
	return j.createdAt
}

// UpdatedAt returns when the job was last updated.
func (j *Job) UpdatedAt() time.Time {
	return j.updatedAt
}

// CanRetry returns true if the job can be retried.
func (j *Job) CanRetry() bool {
	return j.attempts < j.maxRetries
}

// ============================================================================
// Setters / State Transitions
// ============================================================================

// WithPriority sets the job priority.
func (j *Job) WithPriority(priority int) *Job {
	j.priority = priority
	j.updatedAt = time.Now().UTC()
	return j
}

// WithMaxRetries sets the maximum retry attempts.
func (j *Job) WithMaxRetries(maxRetries int) *Job {
	j.maxRetries = maxRetries
	j.updatedAt = time.Now().UTC()
	return j
}

// WithScheduledAt sets when the job should run.
func (j *Job) WithScheduledAt(t time.Time) *Job {
	j.scheduledAt = t
	j.updatedAt = time.Now().UTC()
	return j
}

// MarkRunning marks the job as currently running.
func (j *Job) MarkRunning() error {
	if !j.status.IsProcessable() {
		return ErrJobNotProcessable
	}

	now := time.Now().UTC()
	j.status = StatusRunning
	j.startedAt = &now
	j.attempts++
	j.updatedAt = now
	return nil
}

// MarkCompleted marks the job as successfully completed.
func (j *Job) MarkCompleted() {
	now := time.Now().UTC()
	j.status = StatusCompleted
	j.completedAt = &now
	j.updatedAt = now
}

// MarkFailed marks the job as failed.
// If retries remain, it will be marked as retrying instead.
func (j *Job) MarkFailed(errMsg string, backoff time.Duration) {
	now := time.Now().UTC()
	j.lastError = errMsg
	j.updatedAt = now

	if j.CanRetry() {
		j.status = StatusRetrying
		scheduledAt := now.Add(backoff)
		j.scheduledAt = scheduledAt
	} else {
		j.status = StatusFailed
		j.completedAt = &now
	}
}

// MarkCancelled marks the job as cancelled.
func (j *Job) MarkCancelled() {
	now := time.Now().UTC()
	j.status = StatusCancelled
	j.completedAt = &now
	j.updatedAt = now
}
