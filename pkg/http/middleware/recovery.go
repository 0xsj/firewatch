package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// Recovery is a middleware that recovers from panics and returns a 500 error.
// It logs the panic with stack trace and request context.
//
// This prevents the entire server from crashing when a handler panics.
//
// Usage with Chi:
//
//	r := chi.NewRouter()
//	r.Use(middleware.RequestID)
//	r.Use(middleware.Logger(log))
//	r.Use(middleware.Recovery(log))
func Recovery(log logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Capture stack trace
					stack := debug.Stack()

					// Extract request ID for correlation
					requestID := GetRequestID(r.Context())

					// Log the panic with full context
					log.Error("panic recovered",
						logger.String("error", fmt.Sprintf("%v", err)),
						logger.String("request_id", requestID),
						logger.String("method", r.Method),
						logger.String("path", r.URL.Path),
						logger.String("remote_addr", r.RemoteAddr),
						logger.String("stack_trace", string(stack)),
					)

					// Return 500 Internal Server Error
					// Check if headers were already written
					if !isHeaderWritten(w) {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusInternalServerError)

						// Return safe error message (don't expose panic details to client)
						response := fmt.Sprintf(`{"error":"Internal server error","code":"INTERNAL_ERROR","request_id":"%s"}`, requestID)
						w.Write([]byte(response))
					}
				}
			}()

			// Call next handler
			next.ServeHTTP(w, r)
		})
	}
}

// isHeaderWritten checks if HTTP headers have already been written.
// This prevents "http: superfluous response.WriteHeader" errors.
func isHeaderWritten(w http.ResponseWriter) bool {
	// Try to cast to our wrapped response writer first
	if rw, ok := w.(*responseWriter); ok {
		return rw.wroteHeader
	}

	// For unwrapped response writers, we can't reliably check,
	// so assume false and let http package handle any errors
	return false
}
