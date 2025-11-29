package v1

import (
	"github.com/go-chi/chi/v5"

	"github.com/0xsj/hexagonal-go/pkg/http/middleware"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/security/jwt"
)

// NewRouter creates a new router for tenant routes.
// Mounted at /api/v1/tenants
func NewRouter(
	h *Handler,
	jwtService jwt.Service,
	log logger.Logger,
	corsConfig middleware.CORSConfig,
) chi.Router {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.CORS(corsConfig))
	r.Use(middleware.Logger(log))
	r.Use(middleware.Recovery(log))

	// All tenant routes require authentication
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(jwtService, log))

		// Tenant queries (any authenticated user)
		r.Get("/", h.ListTenants)
		r.Get("/{id}", h.GetTenant)
		r.Get("/slug/{slug}", h.GetTenantBySlug)

		// Admin-only routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAdmin(log))

			// Tenant management
			r.Post("/", h.CreateTenant)
			r.Patch("/{id}", h.UpdateTenant)
			r.Delete("/{id}", h.DeleteTenant)

			// Settings management
			r.Put("/{id}/settings", h.UpdateSettings)

			// Plan management
			r.Post("/{id}/plan", h.ChangePlan)

			// Status management
			r.Post("/{id}/suspend", h.SuspendTenant)
			r.Post("/{id}/reactivate", h.ReactivateTenant)
		})
	})

	return r
}

// Routes creates and configures the Chi router for v1 Tenant API.
// Returns a chi.Router with all middleware and routes configured.
// Deprecated: Use NewRouter instead for consistent mounting pattern.
func (h *Handler) Routes(log logger.Logger, corsConfig middleware.CORSConfig, jwtService jwt.Service) chi.Router {
	r := chi.NewRouter()

	// ========================================================================
	// Global Middleware (applies to all routes)
	// ========================================================================

	// Request ID generation (must be first for tracing)
	r.Use(middleware.RequestID)

	// CORS handling (for browser clients)
	r.Use(middleware.CORS(corsConfig))

	// Request/response logging
	r.Use(middleware.Logger(log))

	// Panic recovery (catch panics, return 500)
	r.Use(middleware.Recovery(log))

	// ========================================================================
	// Health Check (no auth required)
	// ========================================================================
	r.Get("/health", h.Health)

	// ========================================================================
	// API v1 Routes
	// ========================================================================
	r.Route("/api/v1", func(r chi.Router) {
		// ====================================================================
		// Protected Routes (authentication required)
		// ====================================================================
		r.Group(func(r chi.Router) {
			// Add authentication middleware
			r.Use(middleware.RequireAuth(jwtService, log))

			// Tenant queries
			r.Get("/tenants", h.ListTenants)
			r.Get("/tenants/{id}", h.GetTenant)
			r.Get("/tenants/slug/{slug}", h.GetTenantBySlug)

			// ================================================================
			// Admin routes (require admin role)
			// ================================================================
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireAdmin(log))

				// Tenant management
				r.Post("/tenants", h.CreateTenant)
				r.Patch("/tenants/{id}", h.UpdateTenant)
				r.Delete("/tenants/{id}", h.DeleteTenant)

				// Settings management
				r.Put("/tenants/{id}/settings", h.UpdateSettings)

				// Plan management
				r.Post("/tenants/{id}/plan", h.ChangePlan)

				// Status management
				r.Post("/tenants/{id}/suspend", h.SuspendTenant)
				r.Post("/tenants/{id}/reactivate", h.ReactivateTenant)
			})
		})
	})

	return r
}
