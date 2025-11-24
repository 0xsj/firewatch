package metrics

import "time"

// Config holds metrics configuration.
type Config struct {
	// Enabled determines if metrics collection is active.
	Enabled bool `env:"ENABLED"`

	// Port is the port to expose metrics on.
	Port int `env:"PORT"`

	// Path is the HTTP path for metrics endpoint.
	Path string `env:"PATH"`

	// Namespace is the prefix for all metrics.
	Namespace string `env:"NAMESPACE"`

	// Subsystem is an optional subsystem name.
	Subsystem string `env:"SUBSYSTEM"`

	// CollectGoMetrics enables Go runtime metrics.
	CollectGoMetrics bool `env:"COLLECT_GO_METRICS"`

	// CollectProcessMetrics enables process metrics.
	CollectProcessMetrics bool `env:"COLLECT_PROCESS_METRICS"`

	// HistogramBuckets defines custom histogram buckets for latency metrics.
	// If empty, default buckets are used.
	HistogramBuckets []float64
}

// DefaultConfig returns default metrics configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:               true,
		Port:                  9090,
		Path:                  "/metrics",
		Namespace:             "hexagonal",
		Subsystem:             "",
		CollectGoMetrics:      true,
		CollectProcessMetrics: true,
		HistogramBuckets:      DefaultHistogramBuckets(),
	}
}

// DefaultHistogramBuckets returns default latency buckets in seconds.
// Covers range from 1ms to 10s.
func DefaultHistogramBuckets() []float64 {
	return []float64{
		0.001, // 1ms
		0.005, // 5ms
		0.01,  // 10ms
		0.025, // 25ms
		0.05,  // 50ms
		0.1,   // 100ms
		0.25,  // 250ms
		0.5,   // 500ms
		1.0,   // 1s
		2.5,   // 2.5s
		5.0,   // 5s
		10.0,  // 10s
	}
}

// HTTPLatencyBuckets returns buckets optimized for HTTP request latency.
func HTTPLatencyBuckets() []float64 {
	return []float64{
		0.005, // 5ms
		0.01,  // 10ms
		0.025, // 25ms
		0.05,  // 50ms
		0.1,   // 100ms
		0.25,  // 250ms
		0.5,   // 500ms
		1.0,   // 1s
		2.5,   // 2.5s
		5.0,   // 5s
	}
}

// DatabaseLatencyBuckets returns buckets optimized for database query latency.
func DatabaseLatencyBuckets() []float64 {
	return []float64{
		0.001, // 1ms
		0.005, // 5ms
		0.01,  // 10ms
		0.025, // 25ms
		0.05,  // 50ms
		0.1,   // 100ms
		0.25,  // 250ms
		0.5,   // 500ms
		1.0,   // 1s
		float64(5*time.Second) / float64(time.Second),
	}
}
