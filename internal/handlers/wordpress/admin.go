package wordpress

import (
	"net/http"

	"github.com/0xsj/firewatch/internal/handlers"
	"github.com/0xsj/firewatch/pkg/httputil"
)

const sigAdminProbe = "wp-admin-001"

func (wp *WordPress) handleAdmin(w http.ResponseWriter, r *http.Request) {
	wp.logger.Info("admin panel probe",
		"path", r.URL.Path,
		"ip", httputil.ClientIP(r),
	)

	handlers.RecordEvent(wp.store, wp.logger, r, moduleName, "medium", []string{sigAdminProbe})

	// Redirect to login like a real WordPress installation
	http.Redirect(w, r, "/wp-login.php?redirect_to=%2Fwp-admin%2F&reauth=1", http.StatusFound)
}

const sigWPJSON = "wp-api-001"

func (wp *WordPress) handleWPJSON(w http.ResponseWriter, r *http.Request) {
	wp.logger.Info("wp-json API probe",
		"path", r.URL.Path,
		"ip", httputil.ClientIP(r),
	)

	handlers.RecordEvent(wp.store, wp.logger, r, moduleName, "low", []string{sigWPJSON})

	httputil.JSON(w, http.StatusOK, map[string]any{
		"name":           "WordPress Site",
		"description":    "",
		"url":            "https://example.com",
		"namespaces":     []string{"wp/v2", "wp-site-health/v1"},
		"authentication": map[string]any{},
	})
}

func (wp *WordPress) handleStatic(w http.ResponseWriter, r *http.Request) {
	handlers.RecordEvent(wp.store, wp.logger, r, moduleName, "info", nil)

	w.Header().Set("X-Powered-By", "PHP/8.1.0")
	http.Error(w, "403 Forbidden", http.StatusForbidden)
}
