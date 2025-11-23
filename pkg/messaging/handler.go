package messaging

import (
	"context"
	"fmt"
	"time"
)

// HandlerRegistry manages event handlers and their metadata.
// Useful for introspection, debugging, and health checks.
type HandlerRegistry struct {
	handlers map[string][]HandlerInfo
}

// HandlerInfo contains metadata about a registered handler.
type HandlerInfo struct {
	Name            string
	EventType       string
	RegisteredAt    time.Time
	LastInvoked     *time.Time
	InvocationCount int64
	FailureCount    int64
}

// NewHandlerRegistry creates a new handler registry.
func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		handlers: make(map[string][]HandlerInfo),
	}
}

// Register adds a handler to the registry.
func (r *HandlerRegistry) Register(eventType, name string) {
	info := HandlerInfo{
		Name:         name,
		EventType:    eventType,
		RegisteredAt: time.Now(),
	}
	r.handlers[eventType] = append(r.handlers[eventType], info)
}

// GetHandlers returns all handlers for an event type.
func (r *HandlerRegistry) GetHandlers(eventType string) []HandlerInfo {
	return r.handlers[eventType]
}

// GetAllHandlers returns all registered handlers.
func (r *HandlerRegistry) GetAllHandlers() map[string][]HandlerInfo {
	return r.handlers
}

// RecordInvocation records a successful handler invocation.
func (r *HandlerRegistry) RecordInvocation(eventType, name string) {
	handlers := r.handlers[eventType]
	for i := range handlers {
		if handlers[i].Name == name {
			now := time.Now()
			handlers[i].LastInvoked = &now
			handlers[i].InvocationCount++
			break
		}
	}
}

// RecordFailure records a failed handler invocation.
func (r *HandlerRegistry) RecordFailure(eventType, name string) {
	handlers := r.handlers[eventType]
	for i := range handlers {
		if handlers[i].Name == name {
			handlers[i].FailureCount++
			break
		}
	}
}

// ============================================================================
// Named Handler
// ============================================================================

// NamedHandler is an EventHandler with a name for identification.
// Useful for logging, metrics, and debugging.
type NamedHandler interface {
	EventHandler
	Name() string
}

// namedHandler wraps an EventHandler with a name.
type namedHandler struct {
	name    string
	handler EventHandler
}

// WithName wraps an EventHandler with a name.
//
// Example:
//
//	handler := messaging.WithName("UserRegisteredHandler", baseHandler)
func WithName(name string, handler EventHandler) NamedHandler {
	// If already named, return as-is
	if nh, ok := handler.(NamedHandler); ok {
		return nh
	}
	return &namedHandler{
		name:    name,
		handler: handler,
	}
}

func (h *namedHandler) Handle(ctx context.Context, event Event) error {
	return h.handler.Handle(ctx, event)
}

func (h *namedHandler) Name() string {
	return h.name
}

// ============================================================================
// Typed Handler
// ============================================================================

// TypedHandler is an EventHandler that only handles specific event types.
// Automatically filters events by type.
type TypedHandler interface {
	EventHandler
	EventTypes() []string
}

// typedHandler wraps an EventHandler with event type filtering.
type typedHandler struct {
	eventTypes map[string]bool
	handler    EventHandler
}

// ForTypes creates a handler that only processes specific event types.
// Other events are ignored (no error).
//
// Example:
//
//	handler := messaging.ForTypes(
//	    []string{"identity.user.registered", "identity.user.verified"},
//	    baseHandler,
//	)
func ForTypes(eventTypes []string, handler EventHandler) TypedHandler {
	types := make(map[string]bool)
	for _, t := range eventTypes {
		types[t] = true
	}
	return &typedHandler{
		eventTypes: types,
		handler:    handler,
	}
}

// ForType creates a handler for a single event type.
//
// Example:
//
//	handler := messaging.ForType("identity.user.registered", baseHandler)
func ForType(eventType string, handler EventHandler) TypedHandler {
	return ForTypes([]string{eventType}, handler)
}

func (h *typedHandler) Handle(ctx context.Context, event Event) error {
	// Filter by event type
	if !h.eventTypes[event.Type()] {
		return nil // Skip this event
	}
	return h.handler.Handle(ctx, event)
}

func (h *typedHandler) EventTypes() []string {
	types := make([]string, 0, len(h.eventTypes))
	for t := range h.eventTypes {
		types = append(types, t)
	}
	return types
}

// ============================================================================
// Conditional Handler
// ============================================================================

// Predicate is a function that determines if an event should be processed.
type Predicate func(Event) bool

// ConditionalHandler only processes events that match a predicate.
type ConditionalHandler struct {
	predicate Predicate
	handler   EventHandler
}

// When creates a handler that only processes events matching a predicate.
//
// Example:
//
//	// Only process events from "identity" source
//	handler := messaging.When(
//	    func(e Event) bool { return e.Source() == "identity" },
//	    baseHandler,
//	)
func When(predicate Predicate, handler EventHandler) EventHandler {
	return &ConditionalHandler{
		predicate: predicate,
		handler:   handler,
	}
}

func (h *ConditionalHandler) Handle(ctx context.Context, event Event) error {
	if !h.predicate(event) {
		return nil // Skip this event
	}
	return h.handler.Handle(ctx, event)
}

// ============================================================================
// Common Predicates
// ============================================================================

// FromSource creates a predicate that matches events from a specific source.
func FromSource(source string) Predicate {
	return func(e Event) bool {
		return e.Source() == source
	}
}

// WithMetadata creates a predicate that matches events with specific metadata.
func WithMetadata(key string, value any) Predicate {
	return func(e Event) bool {
		if v, ok := e.Metadata()[key]; ok {
			return v == value
		}
		return false
	}
}

// HasMetadataKey creates a predicate that matches events with a metadata key.
func HasMetadataKey(key string) Predicate {
	return func(e Event) bool {
		_, ok := e.Metadata()[key]
		return ok
	}
}

// And combines multiple predicates with AND logic.
func And(predicates ...Predicate) Predicate {
	return func(e Event) bool {
		for _, p := range predicates {
			if !p(e) {
				return false
			}
		}
		return true
	}
}

// Or combines multiple predicates with OR logic.
func Or(predicates ...Predicate) Predicate {
	return func(e Event) bool {
		for _, p := range predicates {
			if p(e) {
				return true
			}
		}
		return false
	}
}

// Not negates a predicate.
func Not(predicate Predicate) Predicate {
	return func(e Event) bool {
		return !predicate(e)
	}
}

// ============================================================================
// Multi Handler
// ============================================================================

// MultiHandler invokes multiple handlers in sequence.
// If any handler fails, the error is returned but other handlers still run.
type MultiHandler struct {
	handlers []EventHandler
}

// Multi creates a handler that invokes multiple handlers.
// All handlers are invoked even if one fails.
// Returns the first error encountered.
//
// Example:
//
//	handler := messaging.Multi(
//	    emailHandler,
//	    analyticsHandler,
//	    loggingHandler,
//	)
func Multi(handlers ...EventHandler) EventHandler {
	return &MultiHandler{handlers: handlers}
}

func (h *MultiHandler) Handle(ctx context.Context, event Event) error {
	var firstErr error
	for _, handler := range h.handlers {
		if err := handler.Handle(ctx, event); err != nil && firstErr == nil {
			firstErr = err
			// Continue with other handlers
		}
	}
	return firstErr
}

// ============================================================================
// Async Handler
// ============================================================================

// AsyncHandler runs a handler asynchronously in a goroutine.
// Errors are logged but not returned.
type AsyncHandler struct {
	handler     EventHandler
	errorLogger func(error)
}

// Async creates an asynchronous handler.
// The handler runs in a goroutine and returns immediately.
//
// Example:
//
//	handler := messaging.Async(slowHandler, func(err error) {
//	    log.Printf("Async handler error: %v", err)
//	})
func Async(handler EventHandler, errorLogger func(error)) EventHandler {
	return &AsyncHandler{
		handler:     handler,
		errorLogger: errorLogger,
	}
}

func (h *AsyncHandler) Handle(ctx context.Context, event Event) error {
	go func() {
		if err := h.handler.Handle(ctx, event); err != nil {
			if h.errorLogger != nil {
				h.errorLogger(err)
			}
		}
	}()
	return nil
}

// ============================================================================
// Handler Helpers
// ============================================================================

// GetHandlerName returns the name of a handler if it implements NamedHandler.
// Otherwise returns a generic name.
func GetHandlerName(handler EventHandler) string {
	if nh, ok := handler.(NamedHandler); ok {
		return nh.Name()
	}
	return fmt.Sprintf("%T", handler)
}

// GetEventTypes returns the event types a handler processes.
// If handler implements TypedHandler, returns its types.
// Otherwise returns empty slice.
func GetEventTypes(handler EventHandler) []string {
	if th, ok := handler.(TypedHandler); ok {
		return th.EventTypes()
	}
	return []string{}
}
