package admin

import (
	"net/http"

	"github.com/0xsj/firewatch/internal/deception"
	"github.com/0xsj/firewatch/internal/handlers"
	"github.com/0xsj/firewatch/pkg/httputil"
)

const (
	sigAdminerProbe = "admin-adminer-001"
	sigAdminerLogin = "admin-adminer-002"
)

func (a *Admin) handleAdminerGet(w http.ResponseWriter, r *http.Request) {
	a.logger.Info("Adminer probe",
		"ip", httputil.ClientIP(r),
	)

	handlers.RecordEvent(a.store, a.logger, r, moduleName, "medium", []string{sigAdminerProbe})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Powered-By", "PHP/8.1.0")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(deception.AdminerLoginPage()))
}

func (a *Admin) handleAdminerPost(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	username := r.FormValue("auth[username]")
	password := r.FormValue("auth[password]")
	server := r.FormValue("auth[server]")

	if username != "" || password != "" {
		a.logger.Info("Adminer login attempt",
			"username", username,
			"server", server,
			"ip", httputil.ClientIP(r),
		)
	}

	handlers.RecordEvent(a.store, a.logger, r, moduleName, "high", []string{sigAdminerProbe, sigAdminerLogin})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Powered-By", "PHP/8.1.0")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(deception.AdminerLoginPage()))
}
