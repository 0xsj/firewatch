package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// Server wraps http.Server with graceful shutdown and configuration.
type Server struct {
	server *http.Server
	logger logger.Logger
}

// Config holds HTTP server configuration.
type Config struct {
	// Host is the server host address (e.g., "localhost", "0.0.0.0")
	// Default: "0.0.0.0" (listen on all interfaces)
	Host string

	// Port is the server port
	// Default: 8080
	Port int

	// ReadTimeout is the maximum duration for reading the entire request
	// Default: 15 seconds
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out writes of the response
	// Default: 15 seconds
	WriteTimeout time.Duration

	// IdleTimeout is the maximum duration to wait for the next request (keep-alive)
	// Default: 60 seconds
	IdleTimeout time.Duration

	// ShutdownTimeout is the maximum duration to wait for active connections to finish
	// during graceful shutdown
	// Default: 30 seconds
	ShutdownTimeout time.Duration
}

// DefaultConfig returns a production-ready server configuration.
func DefaultConfig() Config {
	return Config{
		Host:            "0.0.0.0",
		Port:            8080,
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 30 * time.Second,
	}
}

// DevelopmentConfig returns a configuration suitable for local development.
func DevelopmentConfig() Config {
	return Config{
		Host:            "localhost",
		Port:            8080,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     120 * time.Second,
		ShutdownTimeout: 5 * time.Second,
	}
}

// NewServer creates a new HTTP server with the given handler and configuration.
func NewServer(handler http.Handler, config Config, log logger.Logger) *Server {
	// Apply defaults
	if config.Host == "" {
		config.Host = "0.0.0.0"
	}
	if config.Port == 0 {
		config.Port = 8080
	}
	if config.ReadTimeout == 0 {
		config.ReadTimeout = 15 * time.Second
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 15 * time.Second
	}
	if config.IdleTimeout == 0 {
		config.IdleTimeout = 60 * time.Second
	}
	if config.ShutdownTimeout == 0 {
		config.ShutdownTimeout = 30 * time.Second
	}

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	return &Server{
		server: srv,
		logger: log,
	}
}

// Start starts the HTTP server and blocks until shutdown.
// Listens for interrupt signals (SIGINT, SIGTERM) for graceful shutdown.
//
// Returns an error if the server fails to start or encounters an error during shutdown.
func (s *Server) Start() error {
	// Channel to receive OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Channel to receive server errors
	errChan := make(chan error, 1)

	// Start server in goroutine
	go func() {
		s.logger.Info("starting http server",
			logger.String("addr", s.server.Addr),
		)

		// ListenAndServe blocks until server stops or error occurs
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	// Wait for interrupt signal or error
	select {
	case err := <-errChan:
		s.logger.Error("server error", logger.Err(err))
		return fmt.Errorf("server error: %w", err)

	case sig := <-sigChan:
		s.logger.Info("received shutdown signal",
			logger.String("signal", sig.String()),
		)
		return s.Shutdown()
	}
}

// Shutdown gracefully shuts down the server.
// It waits for active connections to finish or until ShutdownTimeout is reached.
func (s *Server) Shutdown() error {
	s.logger.Info("shutting down http server")

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("server shutdown error", logger.Err(err))
		return fmt.Errorf("shutdown error: %w", err)
	}

	s.logger.Info("http server stopped gracefully")
	return nil
}

// Addr returns the server address (host:port).
func (s *Server) Addr() string {
	return s.server.Addr
}
