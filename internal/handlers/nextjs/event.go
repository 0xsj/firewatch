package nextjs

import (
	"net/http"

	"github.com/0xsj/firewatch/internal/deception"
	"github.com/0xsj/firewatch/internal/middleware"
	"github.com/0xsj/firewatch/internal/storage/models"
	"github.com/0xsj/firewatch/pkg/crypto"
	"github.com/0xsj/firewatch/pkg/httputil"
	"github.com/0xsj/firewatch/pkg/timeutil"
)

// recordEvent creates and persists an event from the current request.
func (n *NextJS) recordEvent(r *http.Request, severity string, signatures []string) {
	event := &models.Event{
		ID:         crypto.UUID4(),
		Timestamp:  timeutil.FormatRFC3339(timeutil.NowUTC()),
		RequestID:  middleware.RequestID(r.Context()),
		SourceIP:   httputil.ClientIP(r),
		Module:     moduleName,
		Method:     r.Method,
		Path:       r.URL.Path,
		Query:      r.URL.RawQuery,
		Headers:    httputil.HeaderMap(r.Header),
		UserAgent:  r.UserAgent(),
		Severity:   severity,
		Signatures: signatures,
	}

	if err := n.store.SaveEvent(r.Context(), event); err != nil {
		n.logger.Error("failed to save event",
			"error", err,
			"event_id", event.ID,
		)
	}
}

// servePage writes a fake Next.js HTML page.
func servePage(w http.ResponseWriter) {
	nonce, _ := crypto.RandomHex(16)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Powered-By", "Next.js")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(deception.NextJSPage("App", nonce)))
}
