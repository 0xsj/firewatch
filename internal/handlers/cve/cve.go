package cve

import (
	"log/slog"

	"github.com/0xsj/firewatch/internal/config"
	"github.com/0xsj/firewatch/internal/handlers"
	"github.com/0xsj/firewatch/internal/storage"
)

const moduleName = "cve"

// CVE is a honeypot module that emulates endpoints vulnerable to
// well-known CVEs (Log4Shell, Spring4Shell, MOVEit, PAN-OS, Struts2,
// Confluence) to detect scanners probing for these vulnerabilities.
type CVE struct {
	cfg    config.CVEModuleConfig
	store  storage.Store
	logger *slog.Logger
}

// New creates a CVE honeypot module.
func New(cfg config.CVEModuleConfig, store storage.Store, logger *slog.Logger) *CVE {
	return &CVE{
		cfg:    cfg,
		store:  store,
		logger: logger.With("module", moduleName),
	}
}

func (c *CVE) Name() string { return moduleName }

// cveEnabled reports whether the given CVE ID should be active.
// If the config CVEs list is empty, all CVEs are enabled.
func (c *CVE) cveEnabled(id string) bool {
	if len(c.cfg.CVEs) == 0 {
		return true
	}
	for _, cve := range c.cfg.CVEs {
		if cve == id {
			return true
		}
	}
	return false
}

func (c *CVE) Routes() []handlers.Route {
	var routes []handlers.Route

	if c.cveEnabled("CVE-2021-44228") {
		routes = append(routes,
			handlers.Route{Pattern: "GET /solr/admin/cores", Handler: c.handleSolrGet},
			handlers.Route{Pattern: "POST /solr/admin/cores", Handler: c.handleSolrPost},
		)
	}

	if c.cveEnabled("CVE-2022-22965") {
		routes = append(routes,
			handlers.Route{Pattern: "GET /actuator/health", Handler: c.handleActuatorHealth},
			handlers.Route{Pattern: "GET /actuator/env", Handler: c.handleActuatorEnv},
		)
	}

	if c.cveEnabled("CVE-2023-34362") {
		routes = append(routes,
			handlers.Route{Pattern: "GET /human.aspx", Handler: c.handleMOVEitGet},
			handlers.Route{Pattern: "POST /guestaccess.aspx", Handler: c.handleMOVEitPost},
			handlers.Route{Pattern: "POST /moveitisapi/moveitisapi.dll", Handler: c.handleMOVEitPost},
		)
	}

	if c.cveEnabled("CVE-2024-3400") {
		routes = append(routes,
			handlers.Route{Pattern: "GET /global-protect/portal/css/login.css", Handler: c.handlePANOSGet},
			handlers.Route{Pattern: "POST /ssl-vpn/hipreport.esp", Handler: c.handlePANOSPost},
		)
	}

	if c.cveEnabled("CVE-2017-5638") {
		routes = append(routes,
			handlers.Route{Pattern: "GET /struts2-showcase/", Handler: c.handleStrutsGet},
			handlers.Route{Pattern: "POST /struts2-showcase/", Handler: c.handleStrutsPost},
		)
	}

	if c.cveEnabled("CVE-2023-22515") {
		routes = append(routes,
			handlers.Route{Pattern: "GET /wiki/", Handler: c.handleConfluenceGet},
			handlers.Route{Pattern: "GET /server-info.action", Handler: c.handleConfluenceInfo},
			handlers.Route{Pattern: "POST /setup/setupadministrator.action", Handler: c.handleConfluencePost},
		)
	}

	return routes
}
