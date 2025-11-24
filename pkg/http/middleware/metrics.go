package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/0xsj/hexagonal-go/pkg/observability/metrics"
)

// HTTPMetrics holds the metrics for HTTP requests.
type HTTPMetrics struct {
	requestsTotal    metrics.Counter
	requestDuration  metrics.Histogram
	requestsInFlight metrics.Gauge
}

// NewHTTPMetrics creates HTTP metrics using the provided metrics provider.
func NewHTTPMetrics(provider metrics.Provider) *HTTPMetrics {
	return &HTTPMetrics{
		requestsTotal: provider.Counter(
			"http_requests_total",
			"Total number of HTTP requests",
			"method", "path", "status",
		),
		requestDuration: provider.Histogram(
			"http_request_duration_seconds",
			"HTTP request duration in seconds",
			metrics.HTTPLatencyBuckets(),
			"method", "path", "status",
		),
		requestsInFlight: provider.Gauge(
			"http_requests_in_flight",
			"Number of HTTP requests currently being processed",
			"method",
		),
	}
}

// Metrics returns middleware that records HTTP metrics.
func Metrics(m *HTTPMetrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Track in-flight requests
			m.requestsInFlight.Inc(r.Method)
			defer m.requestsInFlight.Dec(r.Method)

			// Wrap response writer to capture status code (reuse from logger.go)
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Record start time
			start := time.Now()

			// Process request
			next.ServeHTTP(wrapped, r)

			// Record metrics
			duration := time.Since(start).Seconds()
			status := strconv.Itoa(wrapped.statusCode)
			path := normalizePath(r.URL.Path)

			m.requestsTotal.Inc(r.Method, path, status)
			m.requestDuration.Observe(duration, r.Method, path, status)
		})
	}
}

// normalizePath normalizes URL paths to prevent high-cardinality labels.
// Replaces dynamic segments (UUIDs, IDs) with placeholders.
func normalizePath(path string) string {
	if path == "" {
		return "/"
	}

	// For now, return the path as-is
	// In production, implement proper path normalization
	// to avoid high cardinality issues
	return path
}
