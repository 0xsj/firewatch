package middleware

import (
	"log/slog"
	"net/http"

	"github.com/0xsj/firewatch/internal/fingerprint"
)

// Fingerprint runs the fingerprinting engine on every request,
// storing the result in the request context for downstream handlers.
func Fingerprint(engine *fingerprint.Engine, logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			result := engine.Analyze(r)

			ctx := fingerprint.WithResult(r.Context(), result)
			r = r.WithContext(ctx)

			if result.KnownClient != "" {
				logger.Debug("known client detected",
					"client", result.KnownClient,
					"ja3_hash", result.JA3Hash,
					"anomalies", result.Anomalies,
					"request_id", RequestID(r.Context()),
				)
			}

			next.ServeHTTP(w, r)
		})
	}
}
