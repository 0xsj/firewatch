package messaging

import (
	"context"
	"fmt"
)

// Publisher is the interface for publishing events to the event bus.
// This is a PORT in hexagonal architecture.
//
// Domain code depends on this interface, not the implementation.
// Infrastructure provides concrete implementations (in-memory, NATS, Kafka, etc.).
//
// Design principles:
//   - Publishers are write-only - they emit events but don't receive responses
//   - Publishing is fire-and-forget from the domain's perspective
//   - Publishers handle serialization and transport
//   - Publishers may buffer/batch events for performance
//
// Implementations:
//   - memory.Bus - In-memory event bus (single instance)
//   - nats.Publisher - NATS JetStream (distributed, persistent)
//   - kafka.Publisher - Apache Kafka (high throughput)
type Publisher interface {
	// Publish publishes a single event to the bus.
	// Returns immediately after queuing the event.
	// Does not guarantee delivery (fire-and-forget).
	//
	// Example:
	//
	//	event := messaging.NewEvent("user.registered", "identity", data)
	//	err := publisher.Publish(ctx, event)
	Publish(ctx context.Context, event Event) error

	// PublishBatch publishes multiple events atomically.
	// Either all events are published or none are.
	// More efficient than calling Publish multiple times.
	//
	// Example:
	//
	//	events := []Event{event1, event2, event3}
	//	err := publisher.PublishBatch(ctx, events)
	PublishBatch(ctx context.Context, events []Event) error

	// Close gracefully shuts down the publisher.
	// Waits for pending events to be published.
	Close() error
}

// PublisherFunc is an adapter to allow ordinary functions to be used as Publishers.
// Useful for testing and simple use cases.
//
// Example:
//
//	publisher := messaging.PublisherFunc(func(ctx context.Context, event Event) error {
//	    log.Printf("Published: %s", event.Type())
//	    return nil
//	})
type PublisherFunc func(ctx context.Context, event Event) error

func (f PublisherFunc) Publish(ctx context.Context, event Event) error {
	return f(ctx, event)
}

func (f PublisherFunc) PublishBatch(ctx context.Context, events []Event) error {
	for _, event := range events {
		if err := f(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

func (f PublisherFunc) Close() error {
	return nil
}

// NoOpPublisher is a publisher that does nothing.
// Useful for testing when you don't care about events.
type NoOpPublisher struct{}

func (p *NoOpPublisher) Publish(ctx context.Context, event Event) error {
	return nil
}

func (p *NoOpPublisher) PublishBatch(ctx context.Context, events []Event) error {
	return nil
}

func (p *NoOpPublisher) Close() error {
	return nil
}

// ============================================================================
// Errors
// ============================================================================

// ErrPublishFailed indicates that event publishing failed.
type ErrPublishFailed struct {
	EventID   string
	EventType string
	Reason    string
	Err       error
}

func (e *ErrPublishFailed) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("failed to publish event %s (%s): %s: %v",
			e.EventID, e.EventType, e.Reason, e.Err)
	}
	return fmt.Sprintf("failed to publish event %s (%s): %s",
		e.EventID, e.EventType, e.Reason)
}

func (e *ErrPublishFailed) Unwrap() error {
	return e.Err
}

// NewPublishError creates a new ErrPublishFailed.
func NewPublishError(event Event, reason string, err error) error {
	return &ErrPublishFailed{
		EventID:   event.ID(),
		EventType: event.Type(),
		Reason:    reason,
		Err:       err,
	}
}

// ============================================================================
// Publisher Options
// ============================================================================

// PublisherOption configures a Publisher.
type PublisherOption func(p *PublisherConfig)

// PublisherConfig holds publisher configuration.
type PublisherConfig struct {
	// MaxRetries is the maximum number of retry attempts for failed publishes.
	MaxRetries int

	// BatchSize is the maximum number of events in a batch.
	BatchSize int

	// BufferSize is the size of the internal event buffer.
	BufferSize int

	// Async determines if publishing is asynchronous.
	Async bool
}

// DefaultPublisherConfig returns default publisher configuration.
func DefaultPublisherConfig() PublisherConfig {
	return PublisherConfig{
		MaxRetries: 3,
		BatchSize:  100,
		BufferSize: 1000,
		Async:      true,
	}
}

// WithMaxRetries sets the maximum retry attempts.
func WithMaxRetries(retries int) PublisherOption {
	return func(c *PublisherConfig) {
		c.MaxRetries = retries
	}
}

// WithBatchSize sets the batch size.
func WithBatchSize(size int) PublisherOption {
	return func(c *PublisherConfig) {
		c.BatchSize = size
	}
}

// WithBufferSize sets the buffer size.
func WithBufferSize(size int) PublisherOption {
	return func(c *PublisherConfig) {
		c.BufferSize = size
	}
}

// WithAsync enables or disables asynchronous publishing.
func WithAsync(async bool) PublisherOption {
	return func(c *PublisherConfig) {
		c.Async = async
	}
}

// ApplyOptions applies options to a config.
func (c *PublisherConfig) ApplyOptions(opts ...PublisherOption) {
	for _, opt := range opts {
		opt(c)
	}
}
