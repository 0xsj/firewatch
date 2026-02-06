package handlers

import "net/http"

// Module is the interface that all honeypot modules implement.
// Each module provides a name for identification and a set of
// routes that it handles.
type Module interface {
	// Name returns the module identifier (e.g., "nextjs", "wordpress").
	Name() string

	// Routes returns the HTTP routes this module handles.
	Routes() []Route
}

// Route maps an HTTP pattern to a handler. Patterns follow
// Go 1.22+ ServeMux syntax: "GET /path", "POST /api/{id}", "/prefix/".
type Route struct {
	Pattern string
	Handler http.HandlerFunc
}
