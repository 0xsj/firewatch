package cve

import (
	"io"
	"net/http"
	"strings"

	"github.com/0xsj/firewatch/internal/deception"
	"github.com/0xsj/firewatch/internal/handlers"
	"github.com/0xsj/firewatch/pkg/httputil"
)

const (
	sigLog4ShellProbe   = "cve-log4shell-001"
	sigLog4ShellPayload = "cve-log4shell-002"
)

func (c *CVE) handleSolrGet(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("solr admin probe",
		"ip", httputil.ClientIP(r),
	)

	handlers.RecordEvent(c.store, c.logger, r, moduleName, "medium", []string{sigLog4ShellProbe})

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(deception.SolrAdminPage()))
}

func (c *CVE) handleSolrPost(w http.ResponseWriter, r *http.Request) {
	severity := "high"
	sigs := []string{sigLog4ShellProbe}

	// Check for JNDI injection payload in body or headers.
	body, _ := io.ReadAll(io.LimitReader(r.Body, 8192))
	bodyStr := string(body)
	hasJNDI := strings.Contains(bodyStr, "${jndi:") || strings.Contains(r.Header.Get("X-Api-Version"), "${jndi:")

	if hasJNDI {
		severity = "critical"
		sigs = append(sigs, sigLog4ShellPayload)
		c.logger.Warn("log4shell JNDI payload detected",
			"ip", httputil.ClientIP(r),
		)
	} else {
		c.logger.Info("solr admin POST",
			"ip", httputil.ClientIP(r),
		)
	}

	handlers.RecordEvent(c.store, c.logger, r, moduleName, severity, sigs)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(deception.SolrAdminPage()))
}
