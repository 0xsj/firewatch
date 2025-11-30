package demo

import (
	"github.com/go-chi/chi/v5"

	"github.com/0xsj/hexagonal-go/pkg/http/middleware"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// NewRouter creates a new router for demo routes.
// Mounted at /demo
func NewRouter(h *Handler, log logger.Logger) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger(log))
	r.Use(middleware.Recovery(log))

	// Demo endpoints (no auth required for testing)
	r.Get("/users", h.GetUsers)
	r.Get("/variant", h.GetVariant)
	r.Get("/check", h.CheckFlag)

	return r
}
