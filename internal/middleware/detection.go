package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"

	"github.com/0xsj/firewatch/internal/detection"
	"github.com/0xsj/firewatch/internal/storage"
	"github.com/0xsj/firewatch/internal/storage/models"
	"github.com/0xsj/firewatch/pkg/crypto"
	"github.com/0xsj/firewatch/pkg/httputil"
	"github.com/0xsj/firewatch/pkg/timeutil"
)

// Detection runs the detection engine on every request. When
// signatures or patterns match, an event is recorded (which
// triggers alerts via the AlertingStore).
//
// The middleware buffers the request body so both the detector
// and downstream handlers can read it.
func Detection(det *detection.Detector, store storage.Store, logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Buffer the body so it can be read by both the detector
			// and the downstream handler.
			var body string
			if r.Body != nil && r.ContentLength != 0 {
				data, err := io.ReadAll(io.LimitReader(r.Body, httputil.DefaultMaxBodySize))
				if err == nil && len(data) > 0 {
					body = string(data)
					r.Body = io.NopCloser(bytes.NewReader(data))
				}
			}

			result := det.Detect(r, body)

			if result.Matched() {
				logger.Info("detection hit",
					"signatures", result.SignatureIDs(),
					"severity", result.Severity,
					"ip", httputil.ClientIP(r),
					"path", r.URL.Path,
					"request_id", RequestID(r.Context()),
				)

				// Record a detection event directly to avoid an
				// import cycle with the handlers package.
				event := &models.Event{
					ID:         crypto.UUID4(),
					Timestamp:  timeutil.FormatRFC3339(timeutil.NowUTC()),
					RequestID:  RequestID(r.Context()),
					SourceIP:   httputil.ClientIP(r),
					Module:     "detection",
					Method:     r.Method,
					Path:       r.URL.Path,
					Query:      r.URL.RawQuery,
					Headers:    httputil.HeaderMap(r.Header),
					UserAgent:  r.UserAgent(),
					Severity:   result.Severity,
					Signatures: result.SignatureIDs(),
				}

				if err := store.SaveEvent(r.Context(), event); err != nil {
					logger.Error("failed to save detection event",
						"error", err,
						"event_id", event.ID,
					)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
