package middleware

import "net/http"

// Middleware wraps an http.Handler with additional behavior.
type Middleware func(http.Handler) http.Handler

// Chain composes middlewares so they execute left to right.
// Chain(A, B, C)(handler) results in: A → B → C → handler
func Chain(mws ...Middleware) Middleware {
	return func(final http.Handler) http.Handler {
		for i := len(mws) - 1; i >= 0; i-- {
			final = mws[i](final)
		}
		return final
	}
}

// responseWriter wraps http.ResponseWriter to capture the status
// code and response size for logging.
type responseWriter struct {
	http.ResponseWriter
	status  int
	size    int
	written bool
}

// WrapResponseWriter creates a responseWriter that captures
// status code and bytes written.
func WrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (w *responseWriter) WriteHeader(code int) {
	if !w.written {
		w.status = code
		w.written = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.written = true
	}
	n, err := w.ResponseWriter.Write(b)
	w.size += n
	return n, err
}

// Status returns the HTTP status code written to the response.
func (w *responseWriter) Status() int {
	return w.status
}

// Size returns the total bytes written to the response body.
func (w *responseWriter) Size() int {
	return w.size
}

// Flush implements http.Flusher if the underlying writer supports it.
func (w *responseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
