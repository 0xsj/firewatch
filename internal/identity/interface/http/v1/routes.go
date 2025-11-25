package v1

import (
	"github.com/0xsj/hexagonal-go/pkg/http/middleware"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/security/jwt"
	"github.com/go-chi/chi/v5"
)

// Routes creates and configures the Chi router for v1 Identity API.
// Returns a chi.Router with all middleware and routes configured.
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
		// Public Routes (no authentication required)
		// ====================================================================

		// User Registration
		r.Post("/users/register", h.Register)

		// Authentication
		r.Post("/auth/login", h.Login)
		r.Post("/auth/refresh", h.RefreshToken)

		// Email Verification
		// Supports both GET (from email link) and POST (from API)
		r.Get("/users/verify-email", h.VerifyEmail)
		r.Post("/users/verify-email", h.VerifyEmail)

		// Password Reset
		r.Post("/auth/password/forgot", h.RequestPasswordReset)
		r.Post("/auth/password/reset", h.ResetPassword)

		// ====================================================================
		// Protected Routes (authentication required)
		// ====================================================================

		r.Group(func(r chi.Router) {
			// Add authentication middleware
			r.Use(middleware.RequireAuth(jwtService, log))

			// Auth
			r.Post("/auth/logout", h.Logout)

			// User queries
			r.Get("/users/me", h.GetCurrentUser)
			r.Get("/users/{id}", h.GetUser)
			r.Get("/users", h.ListUsers)

			// User management
			r.Post("/users/me/password", h.ChangePassword)

			// Session management
			r.Get("/sessions", h.ListSessions)

			// User updates (TODO)
			// r.Patch("/users/me", h.UpdateProfile)
			// r.Post("/users/me/change-password", h.ChangePassword)

			// ================================================================
			// Admin routes (require admin role)
			// ================================================================

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireAdmin(log))

				r.Post("/users/{id}/suspend", h.SuspendUser)
				r.Post("/users/{id}/reactivate", h.ReactivateUser)
				r.Post("/users/{id}/role", h.ChangeUserRole)
				r.Delete("/users/{id}", h.DeleteUser)
			})
		})
	})

	return r
}
