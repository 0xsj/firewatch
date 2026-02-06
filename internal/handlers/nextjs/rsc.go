package nextjs

import (
	"net/http"

	"github.com/0xsj/firewatch/internal/deception"
	"github.com/0xsj/firewatch/pkg/httputil"
)

// Signatures for RSC probes.
const (
	sigRSCProbe      = "nextjs-rsc-001"
	sigRSCHeader     = "nextjs-rsc-002"
	sigDebugEndpoint = "nextjs-rsc-003"
)

// handleRSC handles requests to React Server Component endpoints
// and requests carrying RSC-related headers. These indicate
// scanners probing for Next.js internals.
func (n *NextJS) handleRSC(w http.ResponseWriter, r *http.Request) {
	sigs := []string{sigRSCProbe}

	// RSC header present — scanner knows about Next.js internals
	if r.Header.Get("Rsc") != "" {
		sigs = append(sigs, sigRSCHeader)
	}

	// Debug endpoint probe — higher severity
	severity := "medium"
	if r.URL.Path == "/__nextjs_original-stack-frame" {
		sigs = append(sigs, sigDebugEndpoint)
		severity = "high"
	}

	n.logger.Info("rsc probe",
		"path", r.URL.Path,
		"has_rsc_header", r.Header.Get("Rsc") != "",
		"ip", httputil.ClientIP(r),
	)

	n.recordEvent(r, severity, sigs)

	// Return fake RSC flight response
	w.Header().Set("Content-Type", "text/x-component")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(deception.NextJSRSCPayload()))
}
