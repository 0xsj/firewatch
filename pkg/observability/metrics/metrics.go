package metrics

import (
	"context"
	"net/http"
)

// Provider is the port for metrics collection.
// Implementations include Prometheus, DataDog, CloudWatch, etc.
type Provider interface {
	// Counter creates or retrieves a counter metric.
	Counter(name, help string, labels ...string) Counter

	// Gauge creates or retrieves a gauge metric.
	Gauge(name, help string, labels ...string) Gauge

	// Histogram creates or retrieves a histogram metric.
	Histogram(name, help string, buckets []float64, labels ...string) Histogram

	// Handler returns an HTTP handler for exposing metrics.
	Handler() http.Handler

	// Start starts a metrics server on the configured port.
	Start(ctx context.Context) error

	// Close gracefully shuts down the metrics provider.
	Close() error
}

// Counter is a monotonically increasing metric.
// Use for: request counts, errors, completed tasks.
type Counter interface {
	// Inc increments the counter by 1.
	Inc(labels ...string)

	// Add adds the given value to the counter.
	Add(value float64, labels ...string)
}

// Gauge is a metric that can go up and down.
// Use for: current connections, queue size, temperature.
type Gauge interface {
	// Set sets the gauge to the given value.
	Set(value float64, labels ...string)

	// Inc increments the gauge by 1.
	Inc(labels ...string)

	// Dec decrements the gauge by 1.
	Dec(labels ...string)

	// Add adds the given value to the gauge.
	Add(value float64, labels ...string)

	// Sub subtracts the given value from the gauge.
	Sub(value float64, labels ...string)
}

// Histogram measures the distribution of values.
// Use for: request latency, response size.
type Histogram interface {
	// Observe records a value in the histogram.
	Observe(value float64, labels ...string)
}

// Timer is a helper for measuring duration.
type Timer struct {
	histogram Histogram
	labels    []string
	startTime int64
}

// NoopProvider is a no-op implementation of Provider.
// Useful for testing or when metrics are disabled.
type NoopProvider struct{}

func (n *NoopProvider) Counter(name, help string, labels ...string) Counter {
	return &noopCounter{}
}

func (n *NoopProvider) Gauge(name, help string, labels ...string) Gauge {
	return &noopGauge{}
}

func (n *NoopProvider) Histogram(name, help string, buckets []float64, labels ...string) Histogram {
	return &noopHistogram{}
}

func (n *NoopProvider) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func (n *NoopProvider) Start(ctx context.Context) error {
	return nil
}

func (n *NoopProvider) Close() error {
	return nil
}

// Noop implementations

type noopCounter struct{}

func (n *noopCounter) Inc(labels ...string)                {}
func (n *noopCounter) Add(value float64, labels ...string) {}

type noopGauge struct{}

func (n *noopGauge) Set(value float64, labels ...string) {}
func (n *noopGauge) Inc(labels ...string)                {}
func (n *noopGauge) Dec(labels ...string)                {}
func (n *noopGauge) Add(value float64, labels ...string) {}
func (n *noopGauge) Sub(value float64, labels ...string) {}

type noopHistogram struct{}

func (n *noopHistogram) Observe(value float64, labels ...string) {}
