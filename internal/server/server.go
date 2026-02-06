package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/0xsj/firewatch/internal/config"
	"github.com/0xsj/firewatch/internal/middleware"
	"github.com/0xsj/firewatch/internal/storage"
	"github.com/0xsj/firewatch/pkg/errors"
)

const op = errors.Op("server")

// Server is the core Firewatch HTTP server. It serves honeypot
// modules through a middleware pipeline and records captured
// interactions to the store.
type Server struct {
	cfg    *config.Config
	store  storage.Store
	router *Router
	logger *slog.Logger
	http   *http.Server
}

// New creates a Server with the given config, store, and logger.
func New(cfg *config.Config, store storage.Store, logger *slog.Logger) *Server {
	s := &Server{
		cfg:    cfg,
		store:  store,
		logger: logger,
		router: NewRouter(),
	}

	// Build the middleware chain.
	chain := middleware.Chain(
		middleware.Correlation,
		middleware.Logging(logger),
	)

	s.http = &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: chain(s.router),
	}

	if cfg.Server.TLS.Enabled {
		s.http.TLSConfig = NewTLSConfig()
	}

	return s
}

// Start begins listening for requests. It blocks until the server
// is shut down or encounters a fatal error.
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.cfg.Server.Port)

	s.logger.Info("starting server",
		"addr", addr,
		"domain", s.cfg.Server.Domain,
		"tls", s.cfg.Server.TLS.Enabled,
		"modules", s.cfg.EnabledModules(),
	)

	var err error
	if s.cfg.Server.TLS.Enabled {
		err = s.http.ListenAndServeTLS(
			s.cfg.Server.TLS.Cert,
			s.cfg.Server.TLS.Key,
		)
	} else {
		err = s.http.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		return errors.E(op, errors.KindInternal, errors.CodeServerBind, err)
	}
	return nil
}

// Shutdown gracefully stops the server, waiting for active
// connections to finish within the context deadline.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down server")
	if err := s.http.Shutdown(ctx); err != nil {
		return errors.E(op, errors.KindInternal, errors.CodeServerShutdown, err)
	}
	return nil
}

// Router returns the server's router for registering handlers.
func (s *Server) Router() *Router {
	return s.router
}

// Store returns the server's storage backend.
func (s *Server) Store() storage.Store {
	return s.store
}
