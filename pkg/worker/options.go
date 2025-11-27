package worker

import "time"

// Options configures the worker behavior.
type Options struct {
	// Concurrency is the number of jobs to process in parallel.
	// Default: 5
	Concurrency int

	// PollInterval is how often to check for new jobs.
	// Default: 1 second
	PollInterval time.Duration

	// MaxRetries is the maximum number of retry attempts for a failed job.
	// Default: 3
	MaxRetries int

	// RetryBackoff is the base duration for exponential backoff.
	// Actual delay: RetryBackoff * 2^(attempt-1)
	// Default: 5 seconds
	RetryBackoff time.Duration

	// MaxBackoff is the maximum backoff duration.
	// Default: 5 minutes
	MaxBackoff time.Duration

	// JobTimeout is the maximum duration a job can run before being cancelled.
	// Default: 5 minutes
	JobTimeout time.Duration

	// ShutdownTimeout is how long to wait for running jobs to complete on shutdown.
	// Default: 30 seconds
	ShutdownTimeout time.Duration

	// BatchSize is the number of jobs to fetch per poll.
	// Default: 10
	BatchSize int
}

// DefaultOptions returns sensible default options.
func DefaultOptions() Options {
	return Options{
		Concurrency:     5,
		PollInterval:    1 * time.Second,
		MaxRetries:      3,
		RetryBackoff:    5 * time.Second,
		MaxBackoff:      5 * time.Minute,
		JobTimeout:      5 * time.Minute,
		ShutdownTimeout: 30 * time.Second,
		BatchSize:       10,
	}
}

// Option is a function that modifies Options.
type Option func(*Options)

// WithConcurrency sets the number of concurrent workers.
func WithConcurrency(n int) Option {
	return func(o *Options) {
		if n > 0 {
			o.Concurrency = n
		}
	}
}

// WithPollInterval sets the polling interval.
func WithPollInterval(d time.Duration) Option {
	return func(o *Options) {
		if d > 0 {
			o.PollInterval = d
		}
	}
}

// WithMaxRetries sets the maximum retry attempts.
func WithMaxRetries(n int) Option {
	return func(o *Options) {
		if n >= 0 {
			o.MaxRetries = n
		}
	}
}

// WithRetryBackoff sets the base retry backoff duration.
func WithRetryBackoff(d time.Duration) Option {
	return func(o *Options) {
		if d > 0 {
			o.RetryBackoff = d
		}
	}
}

// WithMaxBackoff sets the maximum backoff duration.
func WithMaxBackoff(d time.Duration) Option {
	return func(o *Options) {
		if d > 0 {
			o.MaxBackoff = d
		}
	}
}

// WithJobTimeout sets the job execution timeout.
func WithJobTimeout(d time.Duration) Option {
	return func(o *Options) {
		if d > 0 {
			o.JobTimeout = d
		}
	}
}

// WithShutdownTimeout sets the graceful shutdown timeout.
func WithShutdownTimeout(d time.Duration) Option {
	return func(o *Options) {
		if d > 0 {
			o.ShutdownTimeout = d
		}
	}
}

// WithBatchSize sets the number of jobs to fetch per poll.
func WithBatchSize(n int) Option {
	return func(o *Options) {
		if n > 0 {
			o.BatchSize = n
		}
	}
}

// CalculateBackoff calculates the backoff duration for a given attempt.
// Uses exponential backoff: base * 2^(attempt-1), capped at MaxBackoff.
func (o Options) CalculateBackoff(attempt int) time.Duration {
	if attempt <= 0 {
		attempt = 1
	}

	// Calculate exponential backoff
	backoff := o.RetryBackoff
	for i := 1; i < attempt; i++ {
		backoff *= 2
		if backoff > o.MaxBackoff {
			return o.MaxBackoff
		}
	}

	return backoff
}
