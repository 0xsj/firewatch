package tracing

import (
	"context"
)

// Provider is the port for distributed tracing.
// Implementations include OpenTelemetry, Jaeger, Zipkin, etc.
type Provider interface {
	// Tracer returns a tracer for creating spans.
	Tracer() Tracer

	// Shutdown gracefully shuts down the tracing provider.
	Shutdown(ctx context.Context) error
}

// Tracer creates spans for tracing operations.
type Tracer interface {
	// Start creates a new span and returns it with a new context.
	Start(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span)
}

// Span represents a single operation within a trace.
type Span interface {
	// End completes the span.
	End()

	// SetAttribute sets a key-value attribute on the span.
	SetAttribute(key string, value any)

	// SetAttributes sets multiple attributes on the span.
	SetAttributes(attrs map[string]any)

	// SetStatus sets the span status.
	SetStatus(code StatusCode, description string)

	// RecordError records an error on the span.
	RecordError(err error)

	// AddEvent adds an event to the span.
	AddEvent(name string, attrs map[string]any)

	// SpanContext returns the span's context for propagation.
	SpanContext() SpanContext
}

// SpanContext contains identifying trace information about a span.
type SpanContext struct {
	TraceID    string
	SpanID     string
	TraceFlags byte
}

// IsValid returns true if the span context has valid trace and span IDs.
func (sc SpanContext) IsValid() bool {
	return sc.TraceID != "" && sc.SpanID != ""
}

// StatusCode represents the status of a span.
type StatusCode int

const (
	StatusUnset StatusCode = iota
	StatusOK
	StatusError
)

// SpanKind represents the kind of span.
type SpanKind int

const (
	SpanKindInternal SpanKind = iota
	SpanKindServer
	SpanKindClient
	SpanKindProducer
	SpanKindConsumer
)

// SpanOption configures a span.
type SpanOption func(*SpanConfig)

// SpanConfig holds span configuration.
type SpanConfig struct {
	Kind       SpanKind
	Attributes map[string]any
}

// WithSpanKind sets the span kind.
func WithSpanKind(kind SpanKind) SpanOption {
	return func(c *SpanConfig) {
		c.Kind = kind
	}
}

// WithAttributes sets initial span attributes.
func WithAttributes(attrs map[string]any) SpanOption {
	return func(c *SpanConfig) {
		c.Attributes = attrs
	}
}

// ============================================================================
// Noop Implementation
// ============================================================================

// NoopProvider is a no-op implementation of Provider.
type NoopProvider struct{}

func (n *NoopProvider) Tracer() Tracer {
	return &noopTracer{}
}

func (n *NoopProvider) Shutdown(ctx context.Context) error {
	return nil
}

type noopTracer struct{}

func (n *noopTracer) Start(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	return ctx, &noopSpan{}
}

type noopSpan struct{}

func (n *noopSpan) End()                                          {}
func (n *noopSpan) SetAttribute(key string, value any)            {}
func (n *noopSpan) SetAttributes(attrs map[string]any)            {}
func (n *noopSpan) SetStatus(code StatusCode, description string) {}
func (n *noopSpan) RecordError(err error)                         {}
func (n *noopSpan) AddEvent(name string, attrs map[string]any)    {}
func (n *noopSpan) SpanContext() SpanContext                      { return SpanContext{} }
