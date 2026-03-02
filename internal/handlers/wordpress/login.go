package wordpress

import (
	"net/http"

	"github.com/0xsj/firewatch/internal/deception"
	"github.com/0xsj/firewatch/internal/handlers"
	"github.com/0xsj/firewatch/pkg/httputil"
)

const (
	sigLoginProbe      = "wp-login-001"
	sigBruteForce      = "wp-bruteforce-001"
	sigBruteForceMulti = "wp-bruteforce-002"
)

func (wp *WordPress) handleLoginGet(w http.ResponseWriter, r *http.Request) {
	wp.logger.Info("login page probe",
		"ip", httputil.ClientIP(r),
	)

	handlers.RecordEvent(wp.store, wp.logger, r, moduleName, "medium", []string{sigLoginProbe})

	html := deception.WordPressLoginPage(wp.cfg.FakeVersion)
	if wp.deception.Breadcrumbs {
		html = deception.InjectHTML(html, moduleName, wp.breadcrumbCfg())
		deception.BreadcrumbHeaders(w, moduleName, wp.breadcrumbCfg())
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Powered-By", "PHP/8.1.0")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(html))
}

func (wp *WordPress) handleLoginPost(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	username := r.FormValue("log")
	password := r.FormValue("pwd")

	sigs := []string{sigBruteForce}
	severity := "high"

	if username != "" || password != "" {
		wp.logger.Info("brute force attempt",
			"username", username,
			"ip", httputil.ClientIP(r),
		)
	}

	handlers.RecordEvent(wp.store, wp.logger, r, moduleName, severity, sigs)

	// Return a failed login response that looks real
	html := deception.WordPressLoginPage(wp.cfg.FakeVersion)
	if wp.deception.Breadcrumbs {
		html = deception.InjectHTML(html, moduleName, wp.breadcrumbCfg())
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Powered-By", "PHP/8.1.0")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(html))
}

func (wp *WordPress) breadcrumbCfg() deception.BreadcrumbConfig {
	return deception.BreadcrumbConfig{
		Domain:         "",
		EnabledModules: []string{"admin", "api", "cloud", "cve", "exposure", "nextjs", "wordpress"},
	}
}
