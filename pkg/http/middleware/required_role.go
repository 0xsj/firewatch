package middleware

import (
	"net/http"

	"github.com/0xsj/hexagonal-go/pkg/http/response"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// RequireRole is middleware that requires a specific role.
// Must be used after RequireAuth middleware.
func RequireRole(requiredRole string, log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get role from context (set by RequireAuth)
			role := GetRole(r.Context())
			if role == "" {
				log.Warn("missing role in context",
					logger.String("path", r.URL.Path),
					logger.String("method", r.Method),
				)
				response.Forbidden(w, "insufficient permissions")
				return
			}

			// Check if user has required role
			if role != requiredRole {
				log.Warn("insufficient permissions",
					logger.String("path", r.URL.Path),
					logger.String("method", r.Method),
					logger.String("required_role", requiredRole),
					logger.String("user_role", role),
				)
				response.Forbidden(w, "insufficient permissions")
				return
			}

			// User has required role, continue
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyRole is middleware that requires any of the specified roles.
// Must be used after RequireAuth middleware.
func RequireAnyRole(requiredRoles []string, log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get role from context
			role := GetRole(r.Context())
			if role == "" {
				log.Warn("missing role in context",
					logger.String("path", r.URL.Path),
					logger.String("method", r.Method),
				)
				response.Forbidden(w, "insufficient permissions")
				return
			}

			// Check if user has any of the required roles
			hasRole := false
			for _, requiredRole := range requiredRoles {
				if role == requiredRole {
					hasRole = true
					break
				}
			}

			if !hasRole {
				log.Warn("insufficient permissions",
					logger.String("path", r.URL.Path),
					logger.String("method", r.Method),
					logger.String("user_role", role),
				)
				response.Forbidden(w, "insufficient permissions")
				return
			}

			// User has required role, continue
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin is middleware that requires admin role.
// Convenience wrapper for RequireRole("admin").
func RequireAdmin(log logger.Logger) func(http.Handler) http.Handler {
	return RequireRole("admin", log)
}
