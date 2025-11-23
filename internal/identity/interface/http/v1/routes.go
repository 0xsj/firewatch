package v1

import (
	"github.com/0xsj/hexagonal-go/pkg/http/middleware"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/go-chi/chi/v5"
)

// Routes creates and configures the Chi router for v1 Identity API.
// Returns a chi.Router with all middleware and routes configured.
func (h *Handler) Routes(log logger.Logger, corsConfig middleware.CORSConfig) chi.Router {
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
		// Public Routes (no authentication required)
		// ====================================================================

		// User Registration
		r.Post("/users/register", h.Register)

		// Authentication
		r.Post("/auth/login", h.Login)

		// Email Verification
		// Supports both GET (from email link) and POST (from API)
		r.Get("/users/verify-email", h.VerifyEmail)
		r.Post("/users/verify-email", h.VerifyEmail)

		// ====================================================================
		// Protected Routes (authentication required - TODO)
		// ====================================================================

		// r.Group(func(r chi.Router) {
		// 	// Add authentication middleware here (when we build it)
		// 	// r.Use(middleware.RequireAuth)
		//
		// 	// User queries
		// 	r.Get("/users/{id}", h.GetUser)
		// 	r.Get("/users", h.ListUsers)
		// 	r.Get("/users/me", h.GetCurrentUser)  // TODO
		//
		// 	// User updates
		// 	r.Patch("/users/me", h.UpdateProfile)       // TODO
		// 	r.Post("/users/me/change-password", h.ChangePassword)  // TODO
		//
		// 	// Admin routes (require admin role)
		// 	r.Group(func(r chi.Router) {
		// 		// r.Use(middleware.RequireRole("admin"))
		// 		r.Post("/users/{id}/suspend", h.SuspendUser)  // TODO
		// 		r.Post("/users/{id}/activate", h.ActivateUser)  // TODO
		// 		r.Post("/users/{id}/change-role", h.ChangeRole)  // TODO
		// 	})
		// })

		// For now, make these public for testing
		r.Get("/users/{id}", h.GetUser)
		r.Get("/users", h.ListUsers)
	})

	return r
}
