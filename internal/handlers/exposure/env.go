package exposure

import (
	"net/http"

	"github.com/0xsj/firewatch/internal/deception"
	"github.com/0xsj/firewatch/internal/handlers"
	"github.com/0xsj/firewatch/pkg/httputil"
)

const sigEnvProbe = "exposure-env-001"

func (e *Exposure) handleEnv(w http.ResponseWriter, r *http.Request) {
	e.logger.Info("env file probe",
		"path", r.URL.Path,
		"ip", httputil.ClientIP(r),
	)

	handlers.RecordEvent(e.store, e.logger, r, moduleName, "high", []string{sigEnvProbe})

	content := e.cfg.FakeEnv
	if content == "" {
		content = deception.ExposedEnvFile()
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(content))
}
