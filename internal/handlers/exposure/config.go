package exposure

import (
	"net/http"

	"github.com/0xsj/firewatch/internal/handlers"
	"github.com/0xsj/firewatch/pkg/httputil"
)

const sigConfigProbe = "exposure-config-001"

func (e *Exposure) handleConfig(w http.ResponseWriter, r *http.Request) {
	e.logger.Info("config file probe",
		"path", r.URL.Path,
		"ip", httputil.ClientIP(r),
	)

	handlers.RecordEvent(e.store, e.logger, r, moduleName, "medium", []string{sigConfigProbe})

	// Most config files should return 403 — enough to confirm
	// existence without being too obviously fake
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.Error(w, "403 Forbidden", http.StatusForbidden)
}
