package api

import (
	"net/http"

	"github.com/0xsj/firewatch/internal/handlers"
	"github.com/0xsj/firewatch/pkg/httputil"
)

const (
	sigGraphQLProbe         = "api-graphql-001"
	sigGraphQLIntrospection = "api-graphql-002"
)

func (a *API) handleGraphQL(w http.ResponseWriter, r *http.Request) {
	sigs := []string{sigGraphQLProbe}
	severity := "medium"

	// Check for introspection query
	if r.Method == http.MethodPost {
		body, _ := httputil.ReadBody(r, 0)
		bodyStr := string(body)
		if len(bodyStr) > 0 {
			// Simple check for introspection
			if contains(bodyStr, "__schema") || contains(bodyStr, "__type") || contains(bodyStr, "IntrospectionQuery") {
				sigs = append(sigs, sigGraphQLIntrospection)
				severity = "high"
			}
		}
	}

	a.logger.Info("GraphQL probe",
		"path", r.URL.Path,
		"method", r.Method,
		"ip", httputil.ClientIP(r),
	)

	handlers.RecordEvent(a.store, a.logger, r, moduleName, severity, sigs)

	httputil.JSON(w, http.StatusOK, map[string]any{
		"errors": []map[string]any{
			{"message": "Must provide query string."},
		},
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
