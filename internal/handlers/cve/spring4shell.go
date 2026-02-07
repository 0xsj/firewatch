package cve

import (
	"net/http"

	"github.com/0xsj/firewatch/internal/deception"
	"github.com/0xsj/firewatch/internal/handlers"
	"github.com/0xsj/firewatch/pkg/httputil"
)

const (
	sigSpring4ShellProbe = "cve-spring4shell-001"
	sigSpring4ShellEnv   = "cve-spring4shell-002"
)

func (c *CVE) handleActuatorHealth(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("spring actuator health probe",
		"ip", httputil.ClientIP(r),
	)

	handlers.RecordEvent(c.store, c.logger, r, moduleName, "medium", []string{sigSpring4ShellProbe})

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(deception.SpringBootHealthJSON()))
}

func (c *CVE) handleActuatorEnv(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("spring actuator env probe",
		"ip", httputil.ClientIP(r),
	)

	handlers.RecordEvent(c.store, c.logger, r, moduleName, "high", []string{sigSpring4ShellProbe, sigSpring4ShellEnv})

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(deception.SpringBootEnvJSON()))
}
