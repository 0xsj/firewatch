package middleware

import (
	"context"
	"net/http"

	"github.com/0xsj/hexagonal-go/pkg/http/response"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// PermissionChecker is an interface for checking permissions.
// This allows the middleware to be decoupled from the permissions domain.
type PermissionChecker interface {
	HasPermission(ctx context.Context, userID, tenantID, identityRole, action, resource string) (bool, error)
}

// RequirePermission is middleware that requires a specific permission.
// Must be used after RequireAuth middleware.
func RequirePermission(checker PermissionChecker, action, resource string, log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r.Context())
			tenantID := GetTenantID(r.Context())
			identityRole := GetRole(r.Context())

			if userID == "" {
				log.Warn("missing user_id in context",
					logger.String("path", r.URL.Path),
					logger.String("method", r.Method),
				)
				response.Forbidden(w, "insufficient permissions")
				return
			}

			allowed, err := checker.HasPermission(r.Context(), userID, tenantID, identityRole, action, resource)
			if err != nil {
				log.Error("permission check failed",
					logger.String("path", r.URL.Path),
					logger.String("method", r.Method),
					logger.String("user_id", userID),
					logger.String("permission", action+":"+resource),
					logger.Err(err),
				)
				response.InternalServerError(w, "permission check failed")
				return
			}

			if !allowed {
				log.Warn("permission denied",
					logger.String("path", r.URL.Path),
					logger.String("method", r.Method),
					logger.String("user_id", userID),
					logger.String("permission", action+":"+resource),
				)
				response.Forbidden(w, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission is middleware that requires any of the specified permissions.
// Must be used after RequireAuth middleware.
func RequireAnyPermission(checker PermissionChecker, permissions []string, log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r.Context())
			tenantID := GetTenantID(r.Context())
			identityRole := GetRole(r.Context())

			if userID == "" {
				log.Warn("missing user_id in context",
					logger.String("path", r.URL.Path),
					logger.String("method", r.Method),
				)
				response.Forbidden(w, "insufficient permissions")
				return
			}

			for _, perm := range permissions {
				action, resource := parsePermission(perm)
				if action == "" || resource == "" {
					continue
				}

				allowed, err := checker.HasPermission(r.Context(), userID, tenantID, identityRole, action, resource)
				if err != nil {
					log.Error("permission check failed",
						logger.String("path", r.URL.Path),
						logger.String("permission", perm),
						logger.Err(err),
					)
					continue
				}

				if allowed {
					next.ServeHTTP(w, r)
					return
				}
			}

			log.Warn("permission denied - none of required permissions",
				logger.String("path", r.URL.Path),
				logger.String("method", r.Method),
				logger.String("user_id", userID),
			)
			response.Forbidden(w, "insufficient permissions")
		})
	}
}

// RequireAllPermissions is middleware that requires all of the specified permissions.
// Must be used after RequireAuth middleware.
func RequireAllPermissions(checker PermissionChecker, permissions []string, log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r.Context())
			tenantID := GetTenantID(r.Context())
			identityRole := GetRole(r.Context())

			if userID == "" {
				log.Warn("missing user_id in context",
					logger.String("path", r.URL.Path),
					logger.String("method", r.Method),
				)
				response.Forbidden(w, "insufficient permissions")
				return
			}

			for _, perm := range permissions {
				action, resource := parsePermission(perm)
				if action == "" || resource == "" {
					log.Warn("invalid permission format",
						logger.String("permission", perm),
					)
					response.Forbidden(w, "insufficient permissions")
					return
				}

				allowed, err := checker.HasPermission(r.Context(), userID, tenantID, identityRole, action, resource)
				if err != nil {
					log.Error("permission check failed",
						logger.String("path", r.URL.Path),
						logger.String("permission", perm),
						logger.Err(err),
					)
					response.InternalServerError(w, "permission check failed")
					return
				}

				if !allowed {
					log.Warn("permission denied",
						logger.String("path", r.URL.Path),
						logger.String("method", r.Method),
						logger.String("user_id", userID),
						logger.String("missing_permission", perm),
					)
					response.Forbidden(w, "insufficient permissions")
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// parsePermission parses "action:resource" format.
func parsePermission(perm string) (action, resource string) {
	for i := 0; i < len(perm); i++ {
		if perm[i] == ':' {
			return perm[:i], perm[i+1:]
		}
	}
	return "", ""
}
