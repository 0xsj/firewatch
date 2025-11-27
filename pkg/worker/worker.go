package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// Worker processes jobs from a queue.
type Worker struct {
	queue    Queue
	handlers map[string]Handler
	options  Options
	logger   logger.Logger

	mu      sync.RWMutex
	running bool
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

// New creates a new worker with the given queue and options.
func New(queue Queue, logger logger.Logger, opts ...Option) *Worker {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	return &Worker{
		queue:    queue,
		handlers: make(map[string]Handler),
		options:  options,
		logger:   logger,
		stopCh:   make(chan struct{}),
	}
}

// Register registers a handler for a job type.
func (w *Worker) Register(jobType string, handler Handler) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.handlers[jobType] = handler
	w.logger.Info("registered job handler", logger.String("job_type", jobType))
}

// RegisterFunc registers a handler function for a job type.
func (w *Worker) RegisterFunc(jobType string, fn HandlerFunc) {
	w.Register(jobType, fn)
}

// Start begins processing jobs.
// This is a blocking call - use a goroutine if needed.
func (w *Worker) Start(ctx context.Context) error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return nil
	}
	w.running = true
	w.stopCh = make(chan struct{})
	w.mu.Unlock()

	w.logger.Info("worker starting",
		logger.Int("concurrency", w.options.Concurrency),
		logger.Duration("poll_interval", w.options.PollInterval),
	)

	// Start worker goroutines
	jobCh := make(chan *Job, w.options.Concurrency)

	for i := 0; i < w.options.Concurrency; i++ {
		w.wg.Add(1)
		go w.processLoop(ctx, jobCh, i)
	}

	// Poll loop
	ticker := time.NewTicker(w.options.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("worker context cancelled, shutting down")
			return w.shutdown(jobCh)

		case <-w.stopCh:
			w.logger.Info("worker stop requested, shutting down")
			return w.shutdown(jobCh)

		case <-ticker.C:
			w.poll(ctx, jobCh)
		}
	}
}

// Stop gracefully stops the worker.
func (w *Worker) Stop() {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return
	}
	w.mu.Unlock()

	close(w.stopCh)
}

// Enqueue adds a job to the queue.
func (w *Worker) Enqueue(ctx context.Context, job *Job) error {
	return w.queue.Enqueue(ctx, job)
}

// EnqueueType creates and enqueues a job with the given type and data.
func (w *Worker) EnqueueType(ctx context.Context, jobType string, data any) (*Job, error) {
	job, err := NewJobWithData(jobType, data)
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	if err := w.queue.Enqueue(ctx, job); err != nil {
		return nil, err
	}

	return job, nil
}

// Stats returns queue statistics.
func (w *Worker) Stats(ctx context.Context) (*QueueStats, error) {
	return w.queue.Stats(ctx)
}

// ============================================================================
// Internal
// ============================================================================

// poll fetches jobs from the queue and sends them to the job channel.
func (w *Worker) poll(ctx context.Context, jobCh chan<- *Job) {
	jobs, err := w.queue.Dequeue(ctx, w.options.BatchSize)
	if err != nil {
		w.logger.Error("failed to dequeue jobs", logger.Err(err))
		return
	}

	for _, job := range jobs {
		select {
		case jobCh <- job:
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		}
	}
}

// processLoop processes jobs from the channel.
func (w *Worker) processLoop(ctx context.Context, jobCh <-chan *Job, workerID int) {
	defer w.wg.Done()

	w.logger.Debug("worker goroutine started", logger.Int("worker_id", workerID))

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobCh:
			if !ok {
				return
			}
			w.processJob(ctx, job, workerID)
		}
	}
}

// processJob processes a single job.
func (w *Worker) processJob(ctx context.Context, job *Job, workerID int) {
	w.mu.RLock()
	handler, exists := w.handlers[job.Type()]
	w.mu.RUnlock()

	if !exists {
		w.logger.Error("no handler for job type",
			logger.String("job_id", job.ID().String()),
			logger.String("job_type", job.Type()),
		)
		job.MarkFailed("no handler registered for job type", 0)
		if err := w.queue.Update(ctx, job); err != nil {
			w.logger.Error("failed to update job", logger.Err(err))
		}
		return
	}

	w.logger.Debug("processing job",
		logger.String("job_id", job.ID().String()),
		logger.String("job_type", job.Type()),
		logger.Int("attempt", job.Attempts()),
		logger.Int("worker_id", workerID),
	)

	// Create context with timeout
	jobCtx, cancel := context.WithTimeout(ctx, w.options.JobTimeout)
	defer cancel()

	// Execute handler
	err := handler.Handle(jobCtx, job)

	if err != nil {
		backoff := w.options.CalculateBackoff(job.Attempts())
		job.MarkFailed(err.Error(), backoff)

		w.logger.Warn("job failed",
			logger.String("job_id", job.ID().String()),
			logger.String("job_type", job.Type()),
			logger.Int("attempt", job.Attempts()),
			logger.Bool("will_retry", job.Status() == StatusRetrying),
			logger.Err(err),
		)
	} else {
		job.MarkCompleted()

		w.logger.Info("job completed",
			logger.String("job_id", job.ID().String()),
			logger.String("job_type", job.Type()),
			logger.Int("attempt", job.Attempts()),
		)
	}

	// Update job in queue
	if err := w.queue.Update(ctx, job); err != nil {
		w.logger.Error("failed to update job after processing",
			logger.String("job_id", job.ID().String()),
			logger.Err(err),
		)
	}
}

// shutdown gracefully shuts down the worker.
func (w *Worker) shutdown(jobCh chan *Job) error {
	w.mu.Lock()
	w.running = false
	w.mu.Unlock()

	// Close job channel to signal workers to stop
	close(jobCh)

	// Wait for workers with timeout
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		w.logger.Info("worker shutdown complete")
		return nil
	case <-time.After(w.options.ShutdownTimeout):
		w.logger.Warn("worker shutdown timed out, some jobs may not have completed")
		return fmt.Errorf("shutdown timed out after %v", w.options.ShutdownTimeout)
	}
}
