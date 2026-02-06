package api

import (
	"net/http"

	"github.com/0xsj/firewatch/internal/handlers"
	"github.com/0xsj/firewatch/pkg/httputil"
)

const (
	sigRESTProbe     = "api-rest-001"
	sigRESTAuthProbe = "api-rest-002"
)

func (a *API) handleREST(w http.ResponseWriter, r *http.Request) {
	sigs := []string{sigRESTProbe}
	severity := "low"

	// Check for auth-related probing
	if httputil.HasHeader(r, "Authorization") || httputil.HasHeader(r, "X-Api-Key") {
		sigs = append(sigs, sigRESTAuthProbe)
		severity = "medium"
	}

	a.logger.Info("REST API probe",
		"path", r.URL.Path,
		"method", r.Method,
		"ip", httputil.ClientIP(r),
	)

	handlers.RecordEvent(a.store, a.logger, r, moduleName, severity, sigs)

	httputil.JSON(w, http.StatusUnauthorized, map[string]any{
		"error":   "Unauthorized",
		"message": "Valid authentication credentials are required",
		"status":  401,
	})
}
