package middleware

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// Logger is a middleware that logs HTTP requests and responses.
// It logs:
//   - Request: method, path, remote_addr, request_id
//   - Response: status, duration, bytes_written
//
// Usage with Chi:
//
//	r := chi.NewRouter()
//	r.Use(middleware.RequestID)
//	r.Use(middleware.Logger(log))
func Logger(log logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Capture start time
			start := time.Now()

			// Wrap response writer to capture status code and bytes
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // Default to 200
				bytesWritten:   0,
			}

			// Extract request ID from context
			requestID := GetRequestID(r.Context())

			// Log request start
			log.Info("http request started",
				logger.String("method", r.Method),
				logger.String("path", r.URL.Path),
				logger.String("remote_addr", r.RemoteAddr),
				logger.String("request_id", requestID),
			)

			// Call next handler
			next.ServeHTTP(wrapped, r)

			// Calculate duration
			duration := time.Since(start)

			// Determine log level based on status code
			logFunc := log.Info
			if wrapped.statusCode >= 500 {
				logFunc = log.Error
			} else if wrapped.statusCode >= 400 {
				logFunc = log.Warn
			}

			// Log request completion
			logFunc("http request completed",
				logger.String("method", r.Method),
				logger.String("path", r.URL.Path),
				logger.Int("status", wrapped.statusCode),
				logger.Duration("duration", duration),
				logger.Int("bytes", wrapped.bytesWritten),
				logger.String("request_id", requestID),
			)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code and bytes written.
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
	wroteHeader  bool
}

// WriteHeader captures the status code.
func (rw *responseWriter) WriteHeader(statusCode int) {
	if !rw.wroteHeader {
		rw.statusCode = statusCode
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(statusCode)
	}
}

// Write captures bytes written and ensures header is written.
func (rw *responseWriter) Write(b []byte) (int, error) {
	// Ensure header is written (WriteHeader may not have been called explicitly)
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}

	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

// Flush implements http.Flusher if the underlying ResponseWriter supports it.
func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack implements http.Hijacker if the underlying ResponseWriter supports it.
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("underlying ResponseWriter does not support hijacking")
}
