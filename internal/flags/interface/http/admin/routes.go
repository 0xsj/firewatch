package admin

import (
	"github.com/go-chi/chi/v5"

	"github.com/0xsj/hexagonal-go/pkg/http/middleware"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/security/jwt"
)

// NewRouter creates a new router for the flags admin dashboard.
// Mounted at /admin/flags
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

	// Admin routes (require authentication)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(jwtService, log))

		// Pages
		r.Get("/", h.Index)
		r.Get("/new", h.New)
		r.Post("/", h.Create)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.Show)
			r.Get("/edit", h.Edit)
			r.Put("/", h.Update)
			r.Post("/", h.Update) // For form submissions
			r.Delete("/", h.Delete)

			// HTMX partials
			r.Post("/toggle", h.Toggle)
			r.Get("/row", h.FlagRow)
		})
	})

	return r
}

// NewPublicRouter creates a router without authentication (for development).
// Mounted at /admin/flags
func NewPublicRouter(h *Handler, log logger.Logger) chi.Router {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger(log))
	r.Use(middleware.Recovery(log))

	// Pages
	r.Get("/", h.Index)
	r.Get("/new", h.New)
	r.Post("/", h.Create)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.Show)
		r.Get("/edit", h.Edit)
		r.Put("/", h.Update)
		r.Post("/", h.Update)
		r.Delete("/", h.Delete)

		// HTMX partials
		r.Post("/toggle", h.Toggle)
		r.Get("/row", h.FlagRow)
	})

	return r
}
