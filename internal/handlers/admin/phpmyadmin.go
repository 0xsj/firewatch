package admin

import (
	"net/http"

	"github.com/0xsj/firewatch/internal/deception"
	"github.com/0xsj/firewatch/internal/handlers"
	"github.com/0xsj/firewatch/pkg/httputil"
)

const (
	sigPhpMyAdminProbe = "admin-phpmyadmin-001"
	sigPhpMyAdminLogin = "admin-phpmyadmin-002"
)

func (a *Admin) handlePhpMyAdminGet(w http.ResponseWriter, r *http.Request) {
	a.logger.Info("phpMyAdmin probe",
		"ip", httputil.ClientIP(r),
	)

	handlers.RecordEvent(a.store, a.logger, r, moduleName, "medium", []string{sigPhpMyAdminProbe})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Powered-By", "PHP/8.1.0")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(deception.PhpMyAdminLoginPage()))
}

func (a *Admin) handlePhpMyAdminPost(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	username := r.FormValue("pma_username")
	password := r.FormValue("pma_password")

	if username != "" || password != "" {
		a.logger.Info("phpMyAdmin login attempt",
			"username", username,
			"ip", httputil.ClientIP(r),
		)
	}

	handlers.RecordEvent(a.store, a.logger, r, moduleName, "high", []string{sigPhpMyAdminProbe, sigPhpMyAdminLogin})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Powered-By", "PHP/8.1.0")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(deception.PhpMyAdminLoginPage()))
}
