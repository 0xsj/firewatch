package admin

import (
	"net/http"

	"github.com/0xsj/firewatch/internal/deception"
	"github.com/0xsj/firewatch/internal/handlers"
	"github.com/0xsj/firewatch/pkg/httputil"
)

const (
	sigCPanelProbe = "admin-cpanel-001"
	sigCPanelLogin = "admin-cpanel-002"
)

func (a *Admin) handleCPanelGet(w http.ResponseWriter, r *http.Request) {
	a.logger.Info("cPanel probe",
		"ip", httputil.ClientIP(r),
	)

	handlers.RecordEvent(a.store, a.logger, r, moduleName, "medium", []string{sigCPanelProbe})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(deception.CPanelLoginPage()))
}

func (a *Admin) handleCPanelPost(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	username := r.FormValue("user")
	password := r.FormValue("pass")

	if username != "" || password != "" {
		a.logger.Info("cPanel login attempt",
			"username", username,
			"ip", httputil.ClientIP(r),
		)
	}

	handlers.RecordEvent(a.store, a.logger, r, moduleName, "high", []string{sigCPanelProbe, sigCPanelLogin})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(deception.CPanelLoginPage()))
}
