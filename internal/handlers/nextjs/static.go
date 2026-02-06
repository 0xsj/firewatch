package nextjs

import (
	"net/http"
	"strings"

	"github.com/0xsj/firewatch/internal/deception"
	"github.com/0xsj/firewatch/pkg/httputil"
)

// Signatures for static asset probes.
const (
	sigStaticEnumeration = "nextjs-static-001"
	sigBuildManifest     = "nextjs-static-002"
	sigSourceMap         = "nextjs-static-003"
)

// handleStatic handles requests to /_next/static/, /_next/data/,
// and /_next/image. These detect scanners enumerating Next.js
// build artifacts.
func (n *NextJS) handleStatic(w http.ResponseWriter, r *http.Request) {
	sigs := []string{sigStaticEnumeration}
	severity := "low"
	path := r.URL.Path

	switch {
	case strings.HasSuffix(path, "buildManifest.js"):
		sigs = append(sigs, sigBuildManifest)
		severity = "medium"
		n.logger.Info("build manifest probe",
			"path", path,
			"ip", httputil.ClientIP(r),
		)
		w.Header().Set("Content-Type", "application/javascript")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(deception.NextJSBuildManifest()))

	case strings.HasSuffix(path, ".map"):
		sigs = append(sigs, sigSourceMap)
		severity = "medium"
		n.logger.Info("source map probe",
			"path", path,
			"ip", httputil.ClientIP(r),
		)
		httputil.JSON(w, http.StatusNotFound, map[string]string{
			"error": "Not Found",
		})

	default:
		// Generic static asset request
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("/* chunk */\n"))
	}

	n.recordEvent(r, severity, sigs)
}
