package cve

import (
	"net/http"

	"github.com/0xsj/firewatch/internal/deception"
	"github.com/0xsj/firewatch/internal/handlers"
	"github.com/0xsj/firewatch/pkg/httputil"
)

const (
	sigConfluenceProbe  = "cve-confluence-001"
	sigConfluenceCreate = "cve-confluence-002"
)

func (c *CVE) handleConfluenceGet(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("confluence wiki probe",
		"ip", httputil.ClientIP(r),
	)

	handlers.RecordEvent(c.store, c.logger, r, moduleName, "medium", []string{sigConfluenceProbe})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(deception.ConfluenceLoginPage()))
}

func (c *CVE) handleConfluenceInfo(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("confluence server-info probe",
		"ip", httputil.ClientIP(r),
	)

	handlers.RecordEvent(c.store, c.logger, r, moduleName, "high", []string{sigConfluenceProbe})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(deception.ConfluenceLoginPage()))
}

func (c *CVE) handleConfluencePost(w http.ResponseWriter, r *http.Request) {
	c.logger.Warn("confluence admin creation attempt",
		"ip", httputil.ClientIP(r),
	)

	handlers.RecordEvent(c.store, c.logger, r, moduleName, "critical", []string{sigConfluenceProbe, sigConfluenceCreate})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(deception.ConfluenceLoginPage()))
}
