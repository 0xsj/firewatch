package otel

import (
	"context"
	"fmt"
	"time"

	"github.com/0xsj/hexagonal-go/pkg/observability/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Provider is an OpenTelemetry implementation of tracing.Provider.
type Provider struct {
	config         tracing.Config
	tracerProvider *sdktrace.TracerProvider
	tracer         *Tracer
}

// New creates a new OpenTelemetry tracing provider.
func New(ctx context.Context, config tracing.Config) (*Provider, error) {
	if !config.Enabled {
		return &Provider{
			config: config,
			tracer: &Tracer{tracer: otel.Tracer(config.ServiceName)},
		}, nil
	}

	// Create OTLP exporter
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(config.Endpoint),
		otlptracegrpc.WithTimeout(config.Timeout),
	}

	if config.Insecure {
		opts = append(opts, otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource with service information
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			semconv.DeploymentEnvironment(config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create sampler based on sample rate
	var sampler sdktrace.Sampler
	if config.SampleRate >= 1.0 {
		sampler = sdktrace.AlwaysSample()
	} else if config.SampleRate <= 0.0 {
		sampler = sdktrace.NeverSample()
	} else {
		sampler = sdktrace.TraceIDRatioBased(config.SampleRate)
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(config.ExportInterval),
			sdktrace.WithMaxExportBatchSize(config.BatchSize),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Set global tracer provider and propagator
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer := &Tracer{
		tracer: tp.Tracer(config.ServiceName),
	}

	return &Provider{
		config:         config,
		tracerProvider: tp,
		tracer:         tracer,
	}, nil
}

// Tracer returns a tracer for creating spans.
func (p *Provider) Tracer() tracing.Tracer {
	return p.tracer
}

// Shutdown gracefully shuts down the tracing provider.
func (p *Provider) Shutdown(ctx context.Context) error {
	if p.tracerProvider != nil {
		return p.tracerProvider.Shutdown(ctx)
	}
	return nil
}

// ============================================================================
// Tracer Implementation
// ============================================================================

// Tracer wraps an OpenTelemetry tracer.
type Tracer struct {
	tracer trace.Tracer
}

// Start creates a new span and returns it with a new context.
func (t *Tracer) Start(ctx context.Context, name string, opts ...tracing.SpanOption) (context.Context, tracing.Span) {
	// Apply options
	config := &tracing.SpanConfig{
		Kind:       tracing.SpanKindInternal,
		Attributes: make(map[string]any),
	}
	for _, opt := range opts {
		opt(config)
	}

	// Convert span kind
	spanKind := convertSpanKind(config.Kind)

	// Start span
	ctx, otelSpan := t.tracer.Start(ctx, name,
		trace.WithSpanKind(spanKind),
	)

	// Set initial attributes
	span := &Span{span: otelSpan}
	if len(config.Attributes) > 0 {
		span.SetAttributes(config.Attributes)
	}

	return ctx, span
}

// ============================================================================
// Span Implementation
// ============================================================================

// Span wraps an OpenTelemetry span.
type Span struct {
	span trace.Span
}

// End completes the span.
func (s *Span) End() {
	s.span.End()
}

// SetAttribute sets a key-value attribute on the span.
func (s *Span) SetAttribute(key string, value any) {
	s.span.SetAttributes(convertAttribute(key, value))
}

// SetAttributes sets multiple attributes on the span.
func (s *Span) SetAttributes(attrs map[string]any) {
	otelAttrs := make([]attribute.KeyValue, 0, len(attrs))
	for k, v := range attrs {
		otelAttrs = append(otelAttrs, convertAttribute(k, v))
	}
	s.span.SetAttributes(otelAttrs...)
}

// SetStatus sets the span status.
func (s *Span) SetStatus(code tracing.StatusCode, description string) {
	s.span.SetStatus(convertStatusCode(code), description)
}

// RecordError records an error on the span.
func (s *Span) RecordError(err error) {
	s.span.RecordError(err)
	s.span.SetStatus(codes.Error, err.Error())
}

// AddEvent adds an event to the span.
func (s *Span) AddEvent(name string, attrs map[string]any) {
	otelAttrs := make([]attribute.KeyValue, 0, len(attrs))
	for k, v := range attrs {
		otelAttrs = append(otelAttrs, convertAttribute(k, v))
	}
	s.span.AddEvent(name, trace.WithAttributes(otelAttrs...))
}

// SpanContext returns the span's context for propagation.
func (s *Span) SpanContext() tracing.SpanContext {
	sc := s.span.SpanContext()
	return tracing.SpanContext{
		TraceID:    sc.TraceID().String(),
		SpanID:     sc.SpanID().String(),
		TraceFlags: byte(sc.TraceFlags()),
	}
}

// ============================================================================
// Helpers
// ============================================================================

func convertSpanKind(kind tracing.SpanKind) trace.SpanKind {
	switch kind {
	case tracing.SpanKindServer:
		return trace.SpanKindServer
	case tracing.SpanKindClient:
		return trace.SpanKindClient
	case tracing.SpanKindProducer:
		return trace.SpanKindProducer
	case tracing.SpanKindConsumer:
		return trace.SpanKindConsumer
	default:
		return trace.SpanKindInternal
	}
}

func convertStatusCode(code tracing.StatusCode) codes.Code {
	switch code {
	case tracing.StatusOK:
		return codes.Ok
	case tracing.StatusError:
		return codes.Error
	default:
		return codes.Unset
	}
}

func convertAttribute(key string, value any) attribute.KeyValue {
	switch v := value.(type) {
	case string:
		return attribute.String(key, v)
	case int:
		return attribute.Int(key, v)
	case int64:
		return attribute.Int64(key, v)
	case float64:
		return attribute.Float64(key, v)
	case bool:
		return attribute.Bool(key, v)
	case []string:
		return attribute.StringSlice(key, v)
	case []int:
		return attribute.IntSlice(key, v)
	case []int64:
		return attribute.Int64Slice(key, v)
	case []float64:
		return attribute.Float64Slice(key, v)
	case []bool:
		return attribute.BoolSlice(key, v)
	case time.Duration:
		return attribute.Int64(key, int64(v))
	case error:
		return attribute.String(key, v.Error())
	default:
		return attribute.String(key, fmt.Sprintf("%v", v))
	}
}
