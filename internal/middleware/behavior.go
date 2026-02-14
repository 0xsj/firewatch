package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/0xsj/firewatch/internal/detection"
	"github.com/0xsj/firewatch/internal/storage"
	"github.com/0xsj/firewatch/internal/storage/models"
	"github.com/0xsj/firewatch/pkg/crypto"
	"github.com/0xsj/firewatch/pkg/httputil"
	"github.com/0xsj/firewatch/pkg/timeutil"
)

// Behavior returns middleware that tracks request sequences per-IP
// and generates events when behavioral patterns are detected.
func Behavior(tracker *detection.BehaviorTracker, store storage.Store, logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if tracker == nil {
				next.ServeHTTP(w, r)
				return
			}

			ip := httputil.ClientIP(r)

			// Record this request
			tracker.Record(ip, detection.RequestRecord{
				Timestamp: time.Now(),
				Path:      r.URL.Path,
				Module:    "", // module not known at middleware level
			})

			// Analyze behavioral patterns
			result := tracker.Analyze(ip)
			if result != nil {
				logger.Info("behavioral detection",
					"ip", ip,
					"signatures", result.Signatures,
					"severity", result.Severity,
					"request_id", RequestID(r.Context()),
				)

				event := &models.Event{
					ID:         crypto.UUID4(),
					Timestamp:  timeutil.FormatRFC3339(timeutil.NowUTC()),
					RequestID:  RequestID(r.Context()),
					SourceIP:   ip,
					Module:     "behavior",
					Method:     r.Method,
					Path:       r.URL.Path,
					Query:      r.URL.RawQuery,
					Headers:    httputil.HeaderMap(r.Header),
					UserAgent:  r.UserAgent(),
					Severity:   result.Severity,
					Signatures: result.Signatures,
				}

				if err := store.SaveEvent(r.Context(), event); err != nil {
					logger.Error("failed to save behavioral event",
						"error", err,
						"event_id", event.ID,
					)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
