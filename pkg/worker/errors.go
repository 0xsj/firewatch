package worker

import "errors"

var (
	// ErrJobNotFound indicates the job does not exist.
	ErrJobNotFound = errors.New("job not found")

	// ErrJobAlreadyExists indicates a job with the same ID already exists.
	ErrJobAlreadyExists = errors.New("job already exists")

	// ErrJobNotProcessable indicates the job is not in a processable state.
	ErrJobNotProcessable = errors.New("job is not in a processable state")

	// ErrJobCancelled indicates the job has been cancelled.
	ErrJobCancelled = errors.New("job has been cancelled")

	// ErrHandlerNotFound indicates no handler is registered for the job type.
	ErrHandlerNotFound = errors.New("no handler registered for job type")

	// ErrQueueFull indicates the job queue is at capacity.
	ErrQueueFull = errors.New("job queue is full")

	// ErrInvalidJob indicates the job data is invalid.
	ErrInvalidJob = errors.New("invalid job")

	// ErrMaxRetriesExceeded indicates the job has exceeded maximum retry attempts.
	ErrMaxRetriesExceeded = errors.New("job has exceeded maximum retry attempts")

	// ErrWorkerStopped indicates the worker has been stopped.
	ErrWorkerStopped = errors.New("worker has been stopped")
)
