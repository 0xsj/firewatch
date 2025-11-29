package v1

import (
	"github.com/go-chi/chi/v5"

	"github.com/0xsj/hexagonal-go/pkg/http/middleware"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/security/jwt"
)

// NewRouter creates a new router for feature flag routes.
// Mounted at /api/v1/flags
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

	// All flag routes require authentication
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(jwtService, log))

		// Flag CRUD
		r.Get("/", h.ListFlags)
		r.Post("/", h.CreateFlag)
		r.Get("/by-key", h.GetFlagByKey)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetFlag)
			r.Put("/", h.UpdateFlag)
			r.Delete("/", h.DeleteFlag)

			// Flag status
			r.Post("/enable", h.EnableFlag)
			r.Post("/disable", h.DisableFlag)

			// Rules
			r.Post("/rules", h.AddRule)
			r.Delete("/rules/{ruleId}", h.RemoveRule)

			// Overrides
			r.Post("/overrides", h.SetOverride)
			r.Delete("/overrides", h.RemoveOverride)
		})

		// Evaluation (by key)
		r.Post("/{key}/evaluate", h.EvaluateFlag)
	})

	return r
}

// RegisterRoutes registers flag routes on an existing router.
// Use this when you want to add routes to another module's router.
func RegisterRoutes(
	r chi.Router,
	h *Handler,
	jwtService jwt.Service,
	log logger.Logger,
) {
	r.Route("/flags", func(r chi.Router) {
		r.Use(middleware.RequireAuth(jwtService, log))

		// Flag CRUD
		r.Get("/", h.ListFlags)
		r.Post("/", h.CreateFlag)
		r.Get("/by-key", h.GetFlagByKey)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetFlag)
			r.Put("/", h.UpdateFlag)
			r.Delete("/", h.DeleteFlag)

			// Flag status
			r.Post("/enable", h.EnableFlag)
			r.Post("/disable", h.DisableFlag)

			// Rules
			r.Post("/rules", h.AddRule)
			r.Delete("/rules/{ruleId}", h.RemoveRule)

			// Overrides
			r.Post("/overrides", h.SetOverride)
			r.Delete("/overrides", h.RemoveOverride)
		})

		// Evaluation (by key)
		r.Post("/{key}/evaluate", h.EvaluateFlag)
	})
}
