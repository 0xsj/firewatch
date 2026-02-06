package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/0xsj/firewatch/pkg/httputil"
)

// Logging logs every request with structured fields: method, path,
// status, duration, client IP, user agent, and request ID.
func Logging(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := WrapResponseWriter(w)

			next.ServeHTTP(wrapped, r)

			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"query", r.URL.RawQuery,
				"status", wrapped.Status(),
				"size", wrapped.Size(),
				"duration_ms", time.Since(start).Milliseconds(),
				"ip", httputil.ClientIP(r),
				"user_agent", r.UserAgent(),
				"request_id", RequestID(r.Context()),
			)
		})
	}
}
