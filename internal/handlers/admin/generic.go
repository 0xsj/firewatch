package admin

import (
	"net/http"

	"github.com/0xsj/firewatch/internal/deception"
	"github.com/0xsj/firewatch/internal/handlers"
	"github.com/0xsj/firewatch/pkg/httputil"
)

const (
	sigGenericProbe = "admin-generic-001"
	sigGenericLogin = "admin-generic-002"
)

func (a *Admin) handleGenericGet(w http.ResponseWriter, r *http.Request) {
	a.logger.Info("admin panel probe",
		"ip", httputil.ClientIP(r),
		"path", r.URL.Path,
	)

	handlers.RecordEvent(a.store, a.logger, r, moduleName, "medium", []string{sigGenericProbe})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(deception.GenericAdminLoginPage()))
}

func (a *Admin) handleGenericPost(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username != "" || password != "" {
		a.logger.Info("admin login attempt",
			"username", username,
			"ip", httputil.ClientIP(r),
		)
	}

	handlers.RecordEvent(a.store, a.logger, r, moduleName, "high", []string{sigGenericProbe, sigGenericLogin})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(deception.GenericAdminLoginPage()))
}
