package v1

import (
	"github.com/go-chi/chi/v5"

	"github.com/0xsj/hexagonal-go/pkg/http/middleware"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/security/jwt"
)

// NewRouter creates a new router for email template routes.
// Mounted at /api/v1/email
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

	// All template routes require authentication
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(jwtService, log))

		r.Route("/templates", func(r chi.Router) {
			r.Get("/", h.ListTemplates)
			r.Post("/", h.CreateTemplate)
			r.Get("/by-slug", h.GetTemplateBySlug)
			r.Post("/preview-by-slug", h.PreviewTemplateBySlug)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", h.GetTemplate)
				r.Put("/", h.UpdateTemplate)
				r.Delete("/", h.DeleteTemplate)
				r.Post("/activate", h.ActivateTemplate)
				r.Post("/archive", h.ArchiveTemplate)
				r.Post("/preview", h.PreviewTemplate)
			})
		})
	})

	return r
}

// RegisterRoutes registers email routes on an existing router.
// Use this when you want to add routes to another module's router.
func RegisterRoutes(
	r chi.Router,
	h *Handler,
	jwtService jwt.Service,
	log logger.Logger,
) {
	r.Route("/email/templates", func(r chi.Router) {
		r.Use(middleware.RequireAuth(jwtService, log))

		r.Get("/", h.ListTemplates)
		r.Post("/", h.CreateTemplate)
		r.Get("/by-slug", h.GetTemplateBySlug)
		r.Post("/preview-by-slug", h.PreviewTemplateBySlug)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetTemplate)
			r.Put("/", h.UpdateTemplate)
			r.Delete("/", h.DeleteTemplate)
			r.Post("/activate", h.ActivateTemplate)
			r.Post("/archive", h.ArchiveTemplate)
			r.Post("/preview", h.PreviewTemplate)
		})
	})
}
