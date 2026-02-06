package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// DefaultShutdownTimeout is the maximum time to wait for active
// connections to finish during graceful shutdown.
const DefaultShutdownTimeout = 30 * time.Second

// WaitForShutdown blocks until an interrupt or termination signal
// is received, then gracefully shuts down the server.
func (s *Server) WaitForShutdown() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	sig := <-quit
	s.logger.Info("received signal, shutting down", "signal", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
	defer cancel()

	return s.Shutdown(ctx)
}

// ListenAndShutdown starts the server in a goroutine and blocks
// until a shutdown signal is received. Returns the first error
// from either starting or shutting down.
func (s *Server) ListenAndShutdown() error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- s.Start()
	}()

	// Wait for either a signal or a start error.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		// Server failed to start.
		return err
	case sig := <-quit:
		s.logger.Info("received signal, shutting down", "signal", sig.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
	defer cancel()

	return s.Shutdown(ctx)
}
