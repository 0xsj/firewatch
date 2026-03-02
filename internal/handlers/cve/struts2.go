package cve

import (
	"net/http"
	"strings"

	"github.com/0xsj/firewatch/internal/deception"
	"github.com/0xsj/firewatch/internal/handlers"
	"github.com/0xsj/firewatch/pkg/httputil"
)

const (
	sigStruts2Probe = "cve-struts2-001"
	sigStruts2OGNL  = "cve-struts2-002"
)

func (c *CVE) handleStrutsGet(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("struts2 showcase probe",
		"ip", httputil.ClientIP(r),
	)

	handlers.RecordEvent(c.store, c.logger, r, moduleName, "medium", []string{sigStruts2Probe})

	html := deception.StrutsShowcasePage()
	if c.deception.Breadcrumbs {
		html = deception.InjectHTML(html, moduleName, c.breadcrumbCfg())
		deception.BreadcrumbHeaders(w, moduleName, c.breadcrumbCfg())
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

func (c *CVE) handleStrutsPost(w http.ResponseWriter, r *http.Request) {
	severity := "high"
	sigs := []string{sigStruts2Probe}

	// Check Content-Type for OGNL expression injection (CVE-2017-5638 vector).
	ct := r.Header.Get("Content-Type")
	if strings.Contains(ct, "%{") || strings.Contains(ct, "${") {
		severity = "critical"
		sigs = append(sigs, sigStruts2OGNL)
		c.logger.Warn("OGNL expression in Content-Type",
			"ip", httputil.ClientIP(r),
			"content_type", ct,
		)
	} else {
		c.logger.Info("struts2 showcase POST",
			"ip", httputil.ClientIP(r),
		)
	}

	handlers.RecordEvent(c.store, c.logger, r, moduleName, severity, sigs)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(deception.StrutsShowcasePage()))
}
