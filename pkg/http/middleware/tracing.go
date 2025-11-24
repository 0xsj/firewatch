package middleware

import (
	"fmt"
	"net/http"

	"github.com/0xsj/hexagonal-go/pkg/observability/tracing"
)

// Tracing returns middleware that creates spans for HTTP requests.
func Tracing(tracer tracing.Tracer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create span name from method and path
			spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)

			// Start span with server kind
			ctx, span := tracer.Start(r.Context(), spanName,
				tracing.WithSpanKind(tracing.SpanKindServer),
				tracing.WithAttributes(map[string]any{
					"http.method":     r.Method,
					"http.url":        r.URL.String(),
					"http.target":     r.URL.Path,
					"http.host":       r.Host,
					"http.scheme":     scheme(r),
					"http.user_agent": r.UserAgent(),
					"http.request_id": GetRequestID(r.Context()),
					"net.peer.ip":     r.RemoteAddr,
				}),
			)
			defer span.End()

			// Wrap response writer to capture status code
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Add trace ID to response headers for debugging
			sc := span.SpanContext()
			if sc.IsValid() {
				w.Header().Set("X-Trace-ID", sc.TraceID)
			}

			// Process request with traced context
			next.ServeHTTP(wrapped, r.WithContext(ctx))

			// Set response attributes
			span.SetAttribute("http.status_code", wrapped.statusCode)
			span.SetAttribute("http.response_size", wrapped.bytesWritten)

			// Set span status based on HTTP status code
			if wrapped.statusCode >= 500 {
				span.SetStatus(tracing.StatusError, fmt.Sprintf("HTTP %d", wrapped.statusCode))
			} else if wrapped.statusCode >= 400 {
				span.SetStatus(tracing.StatusError, fmt.Sprintf("HTTP %d", wrapped.statusCode))
			} else {
				span.SetStatus(tracing.StatusOK, "")
			}
		})
	}
}

// scheme returns the request scheme (http or https).
func scheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	if s := r.Header.Get("X-Forwarded-Proto"); s != "" {
		return s
	}
	return "http"
}
