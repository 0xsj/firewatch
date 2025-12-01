// internal/audit/interface/http/v1/routes.go
package v1

import (
	"github.com/go-chi/chi/v5"

	"github.com/0xsj/hexagonal-go/pkg/http/middleware"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/security/jwt"
)

// Routes creates and configures the Chi router for v1 Audit API.
// This router should be mounted at /api/v1/audit.
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
		// Audit entry queries (any authenticated user)
		// Note: In production, you may want to restrict access to admins
		// or add tenant-scoping middleware to filter results
		// ================================================================
		r.Get("/entries", h.ListEntries)
		r.Get("/entries/{id}", h.GetEntry)

		// ================================================================
		// Resource audit trail
		// ================================================================
		r.Get("/resources/{type}/{id}", h.GetResourceTrail)

		// ================================================================
		// Actor activity
		// ================================================================
		r.Get("/actors/{userID}", h.GetActorActivity)
	})

	return r
}

// AdminRoutes creates a router with admin-only access to audit logs.
// Use this if you want to restrict audit access to administrators only.
func (h *Handler) AdminRoutes(log logger.Logger, corsConfig middleware.CORSConfig, jwtService jwt.Service) chi.Router {
	r := chi.NewRouter()

	// ========================================================================
	// Global Middleware
	// ========================================================================
	r.Use(middleware.RequestID)
	r.Use(middleware.CORS(corsConfig))
	r.Use(middleware.Logger(log))
	r.Use(middleware.Recovery(log))

	// ========================================================================
	// Public Routes
	// ========================================================================
	r.Get("/health", h.Health)

	// ========================================================================
	// Admin-Only Routes
	// ========================================================================
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(jwtService, log))
		r.Use(middleware.RequireAdmin(log))

		r.Get("/entries", h.ListEntries)
		r.Get("/entries/{id}", h.GetEntry)
		r.Get("/resources/{type}/{id}", h.GetResourceTrail)
		r.Get("/actors/{userID}", h.GetActorActivity)
	})

	return r
}
