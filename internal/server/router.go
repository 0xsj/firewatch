package server

import (
	"log/slog"
	"net/http"
)

// Router manages route registration for honeypot modules
// and provides a catch-all for unmatched requests.
type Router struct {
	mux      *http.ServeMux
	fallback http.Handler
}

// NewRouter creates a Router with a default fallback handler.
func NewRouter() *Router {
	return &Router{
		mux:      http.NewServeMux(),
		fallback: http.NotFoundHandler(),
	}
}

// Handle registers a handler for the given pattern.
// Patterns follow Go 1.22+ ServeMux syntax: "GET /path", "/prefix/".
func (r *Router) Handle(pattern string, handler http.Handler) {
	r.mux.Handle(pattern, handler)
}

// HandleFunc registers a handler function for the given pattern.
func (r *Router) HandleFunc(pattern string, handler http.HandlerFunc) {
	r.mux.HandleFunc(pattern, handler)
}

// SetFallback sets the handler for requests that don't match any
// registered pattern. The fallback still goes through middleware
// and can be used to capture broad scanning behavior.
func (r *Router) SetFallback(handler http.Handler) {
	r.fallback = handler
}

// Mount registers all routes for a honeypot module under a prefix.
// The module handler receives requests with the prefix stripped.
func (r *Router) Mount(prefix string, handler http.Handler) {
	r.mux.Handle(prefix, http.StripPrefix(prefix, handler))
}

// ServeHTTP dispatches the request to the matching handler.
// If no route matches, the fallback handler is used.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Check if the mux has a match. The ServeMux returns its
	// built-in 404 handler for unmatched routes, so we use
	// the Handler method to detect misses.
	handler, pattern := r.mux.Handler(req)
	if pattern == "" {
		r.fallback.ServeHTTP(w, req)
		return
	}
	handler.ServeHTTP(w, req)
}

// LogRoutes logs all registered routes at debug level.
func (r *Router) LogRoutes(logger *slog.Logger) {
	logger.Debug("registered routes — use specific module patterns to list")
}
