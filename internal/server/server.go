package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/0xsj/firewatch/internal/config"
	"github.com/0xsj/firewatch/internal/detection"
	"github.com/0xsj/firewatch/internal/fingerprint"
	"github.com/0xsj/firewatch/internal/geoip"
	"github.com/0xsj/firewatch/internal/middleware"
	"github.com/0xsj/firewatch/internal/storage"
	"github.com/0xsj/firewatch/pkg/errors"
)

const op = errors.Op("server")

// Server is the core Firewatch HTTP server. It serves honeypot
// modules through a middleware pipeline and records captured
// interactions to the store.
type Server struct {
	cfg         *config.Config
	store       storage.Store
	router      *Router
	logger      *slog.Logger
	http        *http.Server
	rateLimiter *middleware.RateLimiter
}

// New creates a Server with the given config, store, and logger.
// Optional components (fpEngine, detector, geoReader) are included
// in the middleware chain when non-nil.
func New(cfg *config.Config, store storage.Store, fpEngine *fingerprint.Engine, detector *detection.Detector, geoReader *geoip.Reader, logger *slog.Logger) *Server {
	s := &Server{
		cfg:    cfg,
		store:  store,
		logger: logger,
		router: NewRouter(),
	}

	// Build the middleware chain.
	mws := []middleware.Middleware{
		middleware.Correlation,
	}

	// Rate limiting (if enabled)
	if cfg.RateLimit.Enabled {
		rlCfg := middleware.RateLimiterConfig{
			RequestsPerSecond: float64(cfg.RateLimit.RequestsPerSecond),
			Burst:             cfg.RateLimit.Burst,
			CleanupInterval:   time.Duration(cfg.RateLimit.CleanupMinutes) * time.Minute,
		}
		s.rateLimiter = middleware.NewRateLimiter(rlCfg, store, logger)
		mws = append(mws, middleware.RateLimit(s.rateLimiter))
	}

	mws = append(mws, middleware.Logging(logger))
	if geoReader != nil {
		mws = append(mws, middleware.GeoIP(geoReader, logger))
	}
	if fpEngine != nil {
		mws = append(mws, middleware.Fingerprint(fpEngine, logger))
	}
	if detector != nil {
		mws = append(mws, middleware.Detection(detector, store, logger))
	}
	chain := middleware.Chain(mws...)

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

	// Stop rate limiter cleanup goroutine
	if s.rateLimiter != nil {
		s.rateLimiter.Stop()
	}

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

// HTTPServer returns the underlying http.Server for advanced
// configuration (e.g., wiring JA3 capture into TLS).
func (s *Server) HTTPServer() *http.Server {
	return s.http
}
