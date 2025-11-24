package tracing

import "time"

// Config holds tracing configuration.
type Config struct {
	// Enabled determines if tracing is active.
	Enabled bool `env:"ENABLED"`

	// ServiceName is the name of the service in traces.
	ServiceName string `env:"SERVICE_NAME"`

	// ServiceVersion is the version of the service.
	ServiceVersion string `env:"SERVICE_VERSION"`

	// Environment is the deployment environment (dev, staging, prod).
	Environment string `env:"ENVIRONMENT"`

	// Endpoint is the OTLP collector endpoint.
	Endpoint string `env:"ENDPOINT"`

	// Insecure disables TLS for the collector connection.
	Insecure bool `env:"INSECURE"`

	// SampleRate is the fraction of traces to sample (0.0 to 1.0).
	// 1.0 means sample all traces, 0.1 means sample 10%.
	SampleRate float64 `env:"SAMPLE_RATE"`

	// Timeout is the timeout for exporting traces.
	Timeout time.Duration `env:"TIMEOUT"`

	// BatchSize is the maximum number of spans to batch before exporting.
	BatchSize int `env:"BATCH_SIZE"`

	// ExportInterval is how often to export batched spans.
	ExportInterval time.Duration `env:"EXPORT_INTERVAL"`
}

// DefaultConfig returns default tracing configuration.
// Configured for local OpenTelemetry Collector.
func DefaultConfig() Config {
	return Config{
		Enabled:        true,
		ServiceName:    "hexagonal-go",
		ServiceVersion: "0.1.0",
		Environment:    "development",
		Endpoint:       "localhost:4317",
		Insecure:       true,
		SampleRate:     1.0,
		Timeout:        10 * time.Second,
		BatchSize:      512,
		ExportInterval: 5 * time.Second,
	}
}
