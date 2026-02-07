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

// ResponseWriter wraps http.ResponseWriter to capture the status
// code and response size for logging.
type ResponseWriter struct {
	http.ResponseWriter
	status  int
	size    int
	written bool
}

// WrapResponseWriter creates a ResponseWriter that captures
// status code and bytes written.
func WrapResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (w *ResponseWriter) WriteHeader(code int) {
	if !w.written {
		w.status = code
		w.written = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *ResponseWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.written = true
	}
	n, err := w.ResponseWriter.Write(b)
	w.size += n
	return n, err
}

// Status returns the HTTP status code written to the response.
func (w *ResponseWriter) Status() int {
	return w.status
}

// Size returns the total bytes written to the response body.
func (w *ResponseWriter) Size() int {
	return w.size
}

// Flush implements http.Flusher if the underlying writer supports it.
func (w *ResponseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
