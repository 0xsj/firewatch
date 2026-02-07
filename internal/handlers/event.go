package handlers

import (
	"log/slog"
	"net/http"

	"github.com/0xsj/firewatch/internal/geoip"
	"github.com/0xsj/firewatch/internal/middleware"
	"github.com/0xsj/firewatch/internal/storage"
	"github.com/0xsj/firewatch/internal/storage/models"
	"github.com/0xsj/firewatch/pkg/crypto"
	"github.com/0xsj/firewatch/pkg/httputil"
	"github.com/0xsj/firewatch/pkg/timeutil"
)

// RecordEvent creates and persists a honeypot event from the
// current request. This is shared across all honeypot modules
// to avoid duplicating event construction logic.
func RecordEvent(store storage.Store, logger *slog.Logger, r *http.Request, module, severity string, signatures []string) {
	event := &models.Event{
		ID:         crypto.UUID4(),
		Timestamp:  timeutil.FormatRFC3339(timeutil.NowUTC()),
		RequestID:  middleware.RequestID(r.Context()),
		SourceIP:   httputil.ClientIP(r),
		Module:     module,
		Method:     r.Method,
		Path:       r.URL.Path,
		Query:      r.URL.RawQuery,
		Headers:    httputil.HeaderMap(r.Header),
		UserAgent:  r.UserAgent(),
		Severity:   severity,
		Signatures: signatures,
		GeoIP:      geoip.FromContext(r.Context()),
	}

	if err := store.SaveEvent(r.Context(), event); err != nil {
		logger.Error("failed to save event",
			"error", err,
			"event_id", event.ID,
			"module", module,
		)
	}
}
