package worker

// Status represents the current state of a job.
type Status string

const (
	// StatusPending indicates the job is waiting to be processed.
	StatusPending Status = "pending"

	// StatusRunning indicates the job is currently being processed.
	StatusRunning Status = "running"

	// StatusCompleted indicates the job finished successfully.
	StatusCompleted Status = "completed"

	// StatusFailed indicates the job failed after all retries.
	StatusFailed Status = "failed"

	// StatusRetrying indicates the job failed but will be retried.
	StatusRetrying Status = "retrying"

	// StatusCancelled indicates the job was cancelled.
	StatusCancelled Status = "cancelled"
)

// String returns the string representation of the status.
func (s Status) String() string {
	return string(s)
}

// IsTerminal returns true if the status is a final state.
func (s Status) IsTerminal() bool {
	switch s {
	case StatusCompleted, StatusFailed, StatusCancelled:
		return true
	default:
		return false
	}
}

// IsProcessable returns true if the job can be picked up for processing.
func (s Status) IsProcessable() bool {
	return s == StatusPending || s == StatusRetrying
}
