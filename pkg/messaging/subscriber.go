package messaging

import (
	"context"
	"fmt"
)

// Subscriber is the interface for subscribing to events from the event bus.
// This is a PORT in hexagonal architecture.
//
// Event handlers depend on this interface to register themselves.
// Infrastructure provides concrete implementations (in-memory, NATS, Kafka, etc.).
//
// Design principles:
//   - Subscribers are read-only - they receive events but don't publish
//   - Subscribers register handlers for specific event types
//   - Multiple handlers can subscribe to the same event type
//   - Handlers are invoked asynchronously
//   - Handler failures don't affect other handlers
//
// Implementations:
//   - memory.Bus - In-memory event bus
//   - nats.Subscriber - NATS JetStream
//   - kafka.Subscriber - Apache Kafka
type Subscriber interface {
	// Subscribe registers a handler for a specific event type.
	// The handler will be invoked for every event of that type.
	// Multiple handlers can subscribe to the same event type.
	//
	// Example:
	//
	//	handler := NewUserRegisteredHandler(...)
	//	err := subscriber.Subscribe("identity.user.registered", handler)
	Subscribe(eventType string, handler EventHandler) error

	// SubscribeAll registers a handler for all event types.
	// Useful for logging, metrics, or debugging.
	//
	// Example:
	//
	//	logger := NewEventLogger()
	//	err := subscriber.SubscribeAll(logger)
	SubscribeAll(handler EventHandler) error

	// Unsubscribe removes a handler for a specific event type.
	// If the handler is not subscribed, this is a no-op.
	//
	// Example:
	//
	//	err := subscriber.Unsubscribe("identity.user.registered", handler)
	Unsubscribe(eventType string, handler EventHandler) error

	// UnsubscribeAll removes all handlers.
	// Useful for testing or shutdown.
	UnsubscribeAll() error

	// Close gracefully shuts down the subscriber.
	// Waits for in-flight handlers to complete.
	Close() error
}

// EventHandler is the interface that processes events.
// Implement this interface to handle specific events.
//
// Design principles:
//   - Handlers should be idempotent (safe to retry)
//   - Handlers should be fast (offload heavy work to queues)
//   - Handlers should not panic (return errors instead)
//   - Handlers should log errors but not stop processing
//
// Example:
//
//	type UserRegisteredHandler struct {
//	    notificationCmd *SendNotificationCommand
//	}
//
//	func (h *UserRegisteredHandler) Handle(ctx context.Context, event Event) error {
//	    email := event.Data()["email"].(string)
//	    return h.notificationCmd.Send(ctx, "welcome_email", email)
//	}
type EventHandler interface {
	// Handle processes an event.
	// Return an error if processing fails (may trigger retry).
	// Context may contain correlation ID, tenant ID, etc.
	Handle(ctx context.Context, event Event) error
}

// EventHandlerFunc is an adapter to allow ordinary functions to be used as EventHandlers.
// Useful for simple handlers and testing.
//
// Example:
//
//	handler := messaging.EventHandlerFunc(func(ctx context.Context, event Event) error {
//	    log.Printf("Received: %s", event.Type())
//	    return nil
//	})
type EventHandlerFunc func(ctx context.Context, event Event) error

func (f EventHandlerFunc) Handle(ctx context.Context, event Event) error {
	return f(ctx, event)
}

// ============================================================================
// Subscription Management
// ============================================================================

// Subscription represents an active subscription to an event type.
// Can be used to unsubscribe later.
type Subscription struct {
	EventType string
	Handler   EventHandler
	Active    bool
}

// String returns a string representation of the subscription.
func (s *Subscription) String() string {
	status := "active"
	if !s.Active {
		status = "inactive"
	}
	return fmt.Sprintf("Subscription{type=%s, status=%s}", s.EventType, status)
}

// ============================================================================
// Handler Middleware
// ============================================================================

// HandlerMiddleware wraps an EventHandler with additional behavior.
// Useful for logging, metrics, error handling, retry logic, etc.
type HandlerMiddleware func(EventHandler) EventHandler

// Chain combines multiple middlewares into one.
// Middlewares are applied in order: first middleware wraps second, etc.
//
// Example:
//
//	handler := Chain(
//	    LoggingMiddleware(),
//	    MetricsMiddleware(),
//	    RetryMiddleware(3),
//	)(baseHandler)
func Chain(middlewares ...HandlerMiddleware) HandlerMiddleware {
	return func(handler EventHandler) EventHandler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			handler = middlewares[i](handler)
		}
		return handler
	}
}

// ============================================================================
// Common Middlewares
// ============================================================================

// RecoverMiddleware recovers from panics in handlers.
func RecoverMiddleware() HandlerMiddleware {
	return func(next EventHandler) EventHandler {
		return EventHandlerFunc(func(ctx context.Context, event Event) (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("handler panicked: %v", r)
				}
			}()
			return next.Handle(ctx, event)
		})
	}
}

// LoggingMiddleware logs handler execution.
func LoggingMiddleware(logger func(string, ...any)) HandlerMiddleware {
	return func(next EventHandler) EventHandler {
		return EventHandlerFunc(func(ctx context.Context, event Event) error {
			logger("handling event: type=%s, id=%s", event.Type(), event.ID())
			err := next.Handle(ctx, event)
			if err != nil {
				logger("handler error: type=%s, id=%s, err=%v", event.Type(), event.ID(), err)
			}
			return err
		})
	}
}

// ============================================================================
// Errors
// ============================================================================

// ErrSubscriptionFailed indicates that subscription failed.
type ErrSubscriptionFailed struct {
	EventType string
	Reason    string
	Err       error
}

func (e *ErrSubscriptionFailed) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("failed to subscribe to %s: %s: %v",
			e.EventType, e.Reason, e.Err)
	}
	return fmt.Sprintf("failed to subscribe to %s: %s",
		e.EventType, e.Reason)
}

func (e *ErrSubscriptionFailed) Unwrap() error {
	return e.Err
}

// NewSubscriptionError creates a new ErrSubscriptionFailed.
func NewSubscriptionError(eventType, reason string, err error) error {
	return &ErrSubscriptionFailed{
		EventType: eventType,
		Reason:    reason,
		Err:       err,
	}
}

// ErrHandlerFailed indicates that an event handler failed.
type ErrHandlerFailed struct {
	EventID   string
	EventType string
	Handler   string
	Err       error
}

func (e *ErrHandlerFailed) Error() string {
	return fmt.Sprintf("handler %s failed for event %s (%s): %v",
		e.Handler, e.EventID, e.EventType, e.Err)
}

func (e *ErrHandlerFailed) Unwrap() error {
	return e.Err
}

// NewHandlerError creates a new ErrHandlerFailed.
func NewHandlerError(event Event, handlerName string, err error) error {
	return &ErrHandlerFailed{
		EventID:   event.ID(),
		EventType: event.Type(),
		Handler:   handlerName,
		Err:       err,
	}
}

// ============================================================================
// Subscriber Options
// ============================================================================

// SubscriberOption configures a Subscriber.
type SubscriberOption func(s *SubscriberConfig)

// SubscriberConfig holds subscriber configuration.
type SubscriberConfig struct {
	// MaxConcurrency is the maximum number of concurrent handlers.
	MaxConcurrency int

	// RetryAttempts is the number of retry attempts for failed handlers.
	RetryAttempts int

	// ErrorHandler is called when a handler fails after all retries.
	ErrorHandler func(error)

	// Middlewares to apply to all handlers.
	Middlewares []HandlerMiddleware
}

// DefaultSubscriberConfig returns default subscriber configuration.
func DefaultSubscriberConfig() SubscriberConfig {
	return SubscriberConfig{
		MaxConcurrency: 10,
		RetryAttempts:  3,
		ErrorHandler:   func(err error) {}, // no-op
		Middlewares:    []HandlerMiddleware{RecoverMiddleware()},
	}
}

// WithMaxConcurrency sets the maximum concurrent handlers.
func WithMaxConcurrency(n int) SubscriberOption {
	return func(c *SubscriberConfig) {
		c.MaxConcurrency = n
	}
}

// WithRetryAttempts sets the retry attempts.
func WithRetryAttempts(n int) SubscriberOption {
	return func(c *SubscriberConfig) {
		c.RetryAttempts = n
	}
}

// WithErrorHandler sets the error handler.
func WithErrorHandler(handler func(error)) SubscriberOption {
	return func(c *SubscriberConfig) {
		c.ErrorHandler = handler
	}
}

// WithMiddleware adds a middleware to the config.
func WithMiddleware(middleware HandlerMiddleware) SubscriberOption {
	return func(c *SubscriberConfig) {
		c.Middlewares = append(c.Middlewares, middleware)
	}
}

// ApplySubscriberOptions applies options to a config.
func (c *SubscriberConfig) ApplySubscriberOptions(opts ...SubscriberOption) {
	for _, opt := range opts {
		opt(c)
	}
}
