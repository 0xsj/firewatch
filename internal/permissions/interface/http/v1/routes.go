package v1

import (
	"github.com/go-chi/chi/v5"

	"github.com/0xsj/hexagonal-go/pkg/http/middleware"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/security/jwt"
)

// Routes creates and configures the Chi router for v1 Permissions API.
// This router should be mounted at /api/v1/permissions.
func (h *Handler) Routes(log logger.Logger, corsConfig middleware.CORSConfig, jwtService jwt.Service) chi.Router {
	r := chi.NewRouter()

	// ========================================================================
	// Global Middleware (applies to all routes)
	// ========================================================================

	r.Use(middleware.RequestID)
	r.Use(middleware.CORS(corsConfig))
	r.Use(middleware.Logger(log))
	r.Use(middleware.Recovery(log))

	// ========================================================================
	// Public Routes (no authentication required)
	// ========================================================================
	r.Get("/health", h.Health)

	// ========================================================================
	// Protected Routes (authentication required)
	// ========================================================================
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(jwtService, log))

		// ================================================================
		// Permission queries (any authenticated user)
		// ================================================================
		r.Get("/me", h.GetMyPermissions)
		r.Get("/check", h.CheckPermission)

		// ================================================================
		// Role queries (any authenticated user can view)
		// ================================================================
		r.Get("/roles", h.ListRoles)
		r.Get("/roles/{id}", h.GetRole)

		// ================================================================
		// Admin routes (require admin role)
		// ================================================================
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAdmin(log))

			// Role management
			r.Post("/roles", h.CreateRole)
			r.Put("/roles/{id}", h.UpdateRole)
			r.Delete("/roles/{id}", h.DeleteRole)

			// Assignment management
			r.Post("/assignments", h.AssignRole)
			r.Post("/assignments/revoke", h.RevokeRole)
		})
	})

	return r
}
