package cve

import (
	"net/http"

	"github.com/0xsj/firewatch/internal/deception"
	"github.com/0xsj/firewatch/internal/handlers"
	"github.com/0xsj/firewatch/pkg/httputil"
)

const (
	sigMOVEitProbe   = "cve-moveit-001"
	sigMOVEitExploit = "cve-moveit-002"
)

func (c *CVE) handleMOVEitGet(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("MOVEit Transfer probe",
		"ip", httputil.ClientIP(r),
	)

	handlers.RecordEvent(c.store, c.logger, r, moduleName, "medium", []string{sigMOVEitProbe})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(deception.MOVEitLoginPage()))
}

func (c *CVE) handleMOVEitPost(w http.ResponseWriter, r *http.Request) {
	c.logger.Warn("MOVEit exploit attempt",
		"ip", httputil.ClientIP(r),
		"path", r.URL.Path,
	)

	handlers.RecordEvent(c.store, c.logger, r, moduleName, "high", []string{sigMOVEitProbe, sigMOVEitExploit})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(deception.MOVEitLoginPage()))
}
