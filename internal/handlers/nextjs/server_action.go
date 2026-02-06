package nextjs

import (
	"net/http"

	"github.com/0xsj/firewatch/pkg/httputil"
)

// Signatures for server action probes.
const (
	sigServerActionProbe   = "nextjs-action-001"
	sigServerActionPayload = "nextjs-action-002"
)

// handleServerAction handles POST requests that include a
// next-action header. This detects scanners probing for Next.js
// Server Action vulnerabilities (e.g., CVE-2025-29927 patterns).
func (n *NextJS) handleServerAction(w http.ResponseWriter, r *http.Request) {
	actionID := r.Header.Get("Next-Action")
	if actionID == "" {
		// POST without next-action — less interesting but still log it
		n.recordEvent(r, "low", nil)
		servePage(w)
		return
	}

	// Server Action probe detected
	sigs := []string{sigServerActionProbe}

	// Check if there's a payload (higher severity)
	body, _ := httputil.ReadBody(r, 0)
	if len(body) > 0 {
		sigs = append(sigs, sigServerActionPayload)
	}

	n.logger.Info("server action probe",
		"action_id", actionID,
		"body_size", len(body),
		"ip", httputil.ClientIP(r),
	)

	n.recordEvent(r, "high", sigs)

	// Return a plausible server action response
	w.Header().Set("Content-Type", "text/x-component")
	w.Header().Set("X-Action-Revalidated", "[[],0,0]")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`1:{"actionResult":"$undefined","redirectURL":null}`))
}
