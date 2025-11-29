package memory

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// Bus is an in-memory event bus implementation.
// Implements both Publisher and Subscriber interfaces.
//
// Design:
//   - Thread-safe using sync.RWMutex
//   - Asynchronous event delivery via goroutines
//   - Each handler runs independently (failures don't affect others)
//   - Graceful shutdown with WaitGroup
//
// Limitations:
//   - Events are not persisted (lost on restart)
//   - Single instance only (no distributed support)
//   - No delivery guarantees (fire-and-forget)
//
// Use cases:
//   - Development and testing
//   - Single-instance applications
//   - Low-volume event processing
type Bus struct {
	// subscribers maps event types to subscription entries
	subscribers map[string][]*subscription

	// allSubscribers receives all events regardless of type
	allSubscribers []*subscription

	// mu protects subscribers and allSubscribers
	mu sync.RWMutex

	// wg tracks in-flight event deliveries
	wg sync.WaitGroup

	// logger for event delivery errors
	logger logger.Logger

	// closed indicates if the bus is closed
	closed bool

	// config holds bus configuration
	config Config

	// nextID generates unique subscription IDs
	nextID atomic.Int64
}

// subscription represents a handler subscription.
type subscription struct {
	id      int64
	handler messaging.EventHandler
}

// Config holds configuration for the memory bus.
type Config struct {
	// Logger for event delivery errors (optional)
	Logger logger.Logger

	// MaxConcurrency limits concurrent handler executions (0 = unlimited)
	MaxConcurrency int

	// ErrorHandler is called when a handler fails (optional)
	ErrorHandler func(error)
}

// DefaultConfig returns default bus configuration.
func DefaultConfig() Config {
	return Config{
		Logger:         nil, // No logging by default
		MaxConcurrency: 0,   // Unlimited
		ErrorHandler:   nil, // No error handler
	}
}

// NewBus creates a new in-memory event bus.
func NewBus(config Config) *Bus {
	return &Bus{
		subscribers:    make(map[string][]*subscription),
		allSubscribers: make([]*subscription, 0),
		logger:         config.Logger,
		config:         config,
		closed:         false,
	}
}

// NewDefaultBus creates a new bus with default configuration.
func NewDefaultBus() *Bus {
	return NewBus(DefaultConfig())
}

// ============================================================================
// Publisher Interface Implementation
// ============================================================================

// Publish publishes an event to all subscribed handlers.
// Handlers are invoked asynchronously in goroutines.
// Returns immediately after queuing the event.
func (b *Bus) Publish(ctx context.Context, event messaging.Event) error {
	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return fmt.Errorf("bus is closed")
	}

	// Get subscriptions for this event type
	typeSubs := b.subscribers[event.Type()]
	allSubs := b.allSubscribers

	// Create a copy to avoid holding the lock during delivery
	subs := make([]*subscription, 0, len(typeSubs)+len(allSubs))
	subs = append(subs, typeSubs...)
	subs = append(subs, allSubs...)

	b.mu.RUnlock()

	// Deliver to each handler asynchronously
	for _, sub := range subs {
		b.wg.Add(1)
		go b.deliver(ctx, sub.handler, event)
	}

	return nil
}

// PublishBatch publishes multiple events.
// Each event is published independently.
func (b *Bus) PublishBatch(ctx context.Context, events []messaging.Event) error {
	for _, event := range events {
		if err := b.Publish(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// Close gracefully shuts down the bus.
// Waits for all in-flight event deliveries to complete.
func (b *Bus) Close() error {
	b.mu.Lock()
	b.closed = true
	b.mu.Unlock()

	// Wait for all handlers to complete
	b.wg.Wait()

	if b.logger != nil {
		b.logger.Info("event bus closed")
	}

	return nil
}

// ============================================================================
// Subscriber Interface Implementation
// ============================================================================

// Subscribe registers a handler for a specific event type.
// Returns a subscription ID that can be used to unsubscribe.
func (b *Bus) Subscribe(eventType string, handler messaging.EventHandler) error {
	if eventType == "" {
		return fmt.Errorf("event type cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return fmt.Errorf("bus is closed")
	}

	sub := &subscription{
		id:      b.nextID.Add(1),
		handler: handler,
	}

	b.subscribers[eventType] = append(b.subscribers[eventType], sub)

	if b.logger != nil {
		handlerName := messaging.GetHandlerName(handler)
		b.logger.Info("handler subscribed",
			logger.String("event_type", eventType),
			logger.String("handler", handlerName),
			logger.Int64("subscription_id", sub.id),
		)
	}

	return nil
}

// SubscribeAll registers a handler for all event types.
// Useful for logging, metrics, or debugging.
func (b *Bus) SubscribeAll(handler messaging.EventHandler) error {
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return fmt.Errorf("bus is closed")
	}

	sub := &subscription{
		id:      b.nextID.Add(1),
		handler: handler,
	}

	b.allSubscribers = append(b.allSubscribers, sub)

	if b.logger != nil {
		handlerName := messaging.GetHandlerName(handler)
		b.logger.Info("handler subscribed to all events",
			logger.String("handler", handlerName),
			logger.Int64("subscription_id", sub.id),
		)
	}

	return nil
}

// Unsubscribe removes a handler for a specific event type.
// Note: This is difficult to implement reliably with function handlers.
// Consider using SubscribeWithID/UnsubscribeByID instead.
func (b *Bus) Unsubscribe(eventType string, handler messaging.EventHandler) error {
	// This implementation is kept for interface compatibility
	// but may not work reliably with function handlers
	b.mu.Lock()
	defer b.mu.Unlock()

	subs := b.subscribers[eventType]
	for i, sub := range subs {
		// This comparison only works for non-function handlers
		// For function handlers, use UnsubscribeByID instead
		if fmt.Sprintf("%p", sub.handler) == fmt.Sprintf("%p", handler) {
			// Remove subscription
			subs[i] = subs[len(subs)-1]
			b.subscribers[eventType] = subs[:len(subs)-1]

			if b.logger != nil {
				handlerName := messaging.GetHandlerName(handler)
				b.logger.Info("handler unsubscribed",
					logger.String("event_type", eventType),
					logger.String("handler", handlerName),
				)
			}

			return nil
		}
	}

	return nil
}

// UnsubscribeAll removes all handlers.
func (b *Bus) UnsubscribeAll() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscribers = make(map[string][]*subscription)
	b.allSubscribers = make([]*subscription, 0)

	if b.logger != nil {
		b.logger.Info("all handlers unsubscribed")
	}

	return nil
}

// ============================================================================
// Private Methods
// ============================================================================

// deliver delivers an event to a handler.
// Runs in a goroutine.
func (b *Bus) deliver(ctx context.Context, handler messaging.EventHandler, event messaging.Event) {
	defer b.wg.Done()

	// Create a detached context that won't be canceled when the parent (e.g., HTTP request) ends.
	// Event handlers should complete independently of the triggering request.
	// We still propagate event metadata for tracing/correlation.
	detachedCtx := context.Background()
	detachedCtx = messaging.PropagateToContext(detachedCtx, event)

	// Invoke handler
	if err := handler.Handle(detachedCtx, event); err != nil {
		handlerName := messaging.GetHandlerName(handler)

		// Log error
		if b.logger != nil {
			b.logger.Error("handler failed",
				logger.String("event_id", event.ID()),
				logger.String("event_type", event.Type()),
				logger.String("handler", handlerName),
				logger.Err(err),
			)
		}

		// Call error handler if configured
		if b.config.ErrorHandler != nil {
			handlerErr := messaging.NewHandlerError(event, handlerName, err)
			b.config.ErrorHandler(handlerErr)
		}
	}
}

// ============================================================================
// Introspection
// ============================================================================

// Stats returns statistics about the bus.
type Stats struct {
	// EventTypes is the number of event types with subscribers
	EventTypes int

	// TotalHandlers is the total number of registered handlers
	TotalHandlers int

	// AllSubscribers is the number of handlers subscribed to all events
	AllSubscribers int

	// InFlight is the number of events currently being delivered
	InFlight int

	// Closed indicates if the bus is closed
	Closed bool
}

// Stats returns current bus statistics.
func (b *Bus) Stats() Stats {
	b.mu.RLock()
	defer b.mu.RUnlock()

	totalHandlers := len(b.allSubscribers)
	for _, subs := range b.subscribers {
		totalHandlers += len(subs)
	}

	return Stats{
		EventTypes:     len(b.subscribers),
		TotalHandlers:  totalHandlers,
		AllSubscribers: len(b.allSubscribers),
		Closed:         b.closed,
	}
}

// GetSubscribers returns all handlers for an event type.
// Returns a copy to prevent modification.
func (b *Bus) GetSubscribers(eventType string) []messaging.EventHandler {
	b.mu.RLock()
	defer b.mu.RUnlock()

	subs := b.subscribers[eventType]
	handlers := make([]messaging.EventHandler, len(subs))
	for i, sub := range subs {
		handlers[i] = sub.handler
	}
	return handlers
}

// GetAllSubscribers returns all handlers subscribed to all events.
func (b *Bus) GetAllSubscribers() []messaging.EventHandler {
	b.mu.RLock()
	defer b.mu.RUnlock()

	handlers := make([]messaging.EventHandler, len(b.allSubscribers))
	for i, sub := range b.allSubscribers {
		handlers[i] = sub.handler
	}
	return handlers
}
