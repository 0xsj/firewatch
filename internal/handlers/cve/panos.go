package cve

import (
	"net/http"

	"github.com/0xsj/firewatch/internal/handlers"
	"github.com/0xsj/firewatch/pkg/httputil"
)

const (
	sigPANOSProbe  = "cve-panos-001"
	sigPANOSInject = "cve-panos-002"
	panosLoginCSS  = `body{font-family:Arial,sans-serif;background:#f5f5f5}.login-panel{max-width:400px;margin:80px auto;background:#fff;padding:2em;border:1px solid #ddd;border-radius:4px}h1{color:#333;font-size:1.3em}`
)

func (c *CVE) handlePANOSGet(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("PAN-OS GlobalProtect probe",
		"ip", httputil.ClientIP(r),
	)

	handlers.RecordEvent(c.store, c.logger, r, moduleName, "medium", []string{sigPANOSProbe})

	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(panosLoginCSS))
}

func (c *CVE) handlePANOSPost(w http.ResponseWriter, r *http.Request) {
	c.logger.Warn("PAN-OS command injection attempt",
		"ip", httputil.ClientIP(r),
	)

	handlers.RecordEvent(c.store, c.logger, r, moduleName, "critical", []string{sigPANOSProbe, sigPANOSInject})

	w.WriteHeader(http.StatusOK)
}
