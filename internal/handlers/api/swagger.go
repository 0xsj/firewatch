package api

import (
	"net/http"

	"github.com/0xsj/firewatch/internal/handlers"
	"github.com/0xsj/firewatch/pkg/httputil"
)

const sigSwaggerProbe = "api-swagger-001"

func (a *API) handleSwagger(w http.ResponseWriter, r *http.Request) {
	a.logger.Info("swagger/openapi probe",
		"path", r.URL.Path,
		"ip", httputil.ClientIP(r),
	)

	handlers.RecordEvent(a.store, a.logger, r, moduleName, "medium", []string{sigSwaggerProbe})

	// Return a minimal OpenAPI spec
	paths := map[string]any{
		"/api/v1/users": map[string]any{
			"get": map[string]any{
				"summary":  "List users",
				"security": []map[string]any{{"bearerAuth": []string{}}},
			},
		},
		"/api/v1/admin": map[string]any{
			"get": map[string]any{
				"summary":  "Admin panel",
				"security": []map[string]any{{"bearerAuth": []string{}}},
			},
		},
	}

	if a.deception.Breadcrumbs {
		paths["/internal/health"] = map[string]any{
			"get": map[string]any{
				"summary":  "Health check (internal)",
				"security": []map[string]any{{"bearerAuth": []string{}}},
			},
		}
		paths["/admin/backup"] = map[string]any{
			"get": map[string]any{
				"summary":  "Database backup endpoint",
				"security": []map[string]any{{"bearerAuth": []string{}}},
			},
		}
	}

	httputil.JSON(w, http.StatusOK, map[string]any{
		"openapi": "3.0.0",
		"info": map[string]any{
			"title":   "Internal API",
			"version": "1.0.0",
		},
		"paths": paths,
	})
}
