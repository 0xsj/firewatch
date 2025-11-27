package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/0xsj/hexagonal-go/cmd/worker/config"
	"github.com/0xsj/hexagonal-go/cmd/worker/handlers"
	"github.com/0xsj/hexagonal-go/pkg/email"
	"github.com/0xsj/hexagonal-go/pkg/email/smtp"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger/console"
	"github.com/0xsj/hexagonal-go/pkg/worker"
	"github.com/0xsj/hexagonal-go/pkg/worker/memory"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ========================================================================
	// Load Configuration
	// ========================================================================
	cfg, err := config.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// ========================================================================
	// Initialize Logger
	// ========================================================================
	log := console.NewWithLevel(logger.ParseLevel(cfg.Logger.Level))
	log.Info("starting worker",
		logger.Int("concurrency", cfg.Worker.Concurrency),
		logger.String("queue_type", cfg.Worker.QueueType),
	)

	// ========================================================================
	// Initialize Queue
	// ========================================================================
	queue, err := initQueue(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize queue: %w", err)
	}

	// ========================================================================
	// Initialize Email Sender
	// ========================================================================
	emailSender := initEmailSender(cfg)

	// ========================================================================
	// Initialize Worker
	// ========================================================================
	w := worker.New(
		queue,
		log,
		worker.WithConcurrency(cfg.Worker.Concurrency),
		worker.WithPollInterval(cfg.Worker.PollInterval),
		worker.WithMaxRetries(cfg.Worker.MaxRetries),
		worker.WithRetryBackoff(cfg.Worker.RetryBackoff),
		worker.WithMaxBackoff(cfg.Worker.MaxBackoff),
		worker.WithJobTimeout(cfg.Worker.JobTimeout),
		worker.WithShutdownTimeout(cfg.Worker.ShutdownTimeout),
		worker.WithBatchSize(cfg.Worker.BatchSize),
	)

	// ========================================================================
	// Register Handlers
	// ========================================================================
	registry := handlers.SetupHandlers(emailSender, log)
	registry.RegisterAll(w)

	log.Info("job handlers registered")

	// ========================================================================
	// Handle Shutdown Signals
	// ========================================================================
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Info("received shutdown signal", logger.String("signal", sig.String()))
		cancel()
	}()

	// ========================================================================
	// Start Worker
	// ========================================================================
	log.Info("worker started, press Ctrl+C to stop")
	printStats(w, log)

	if err := w.Start(ctx); err != nil {
		return fmt.Errorf("worker error: %w", err)
	}

	log.Info("worker stopped")
	return nil
}

// initQueue initializes the job queue based on configuration.
func initQueue(cfg *config.Config) (worker.Queue, error) {
	switch cfg.Worker.QueueType {
	case "memory":
		return memory.NewQueue(), nil
	case "postgres":
		// TODO: Implement postgres queue
		return nil, fmt.Errorf("postgres queue not yet implemented")
	default:
		return nil, fmt.Errorf("unknown queue type: %s", cfg.Worker.QueueType)
	}
}

// initEmailSender initializes the email sender.
func initEmailSender(cfg *config.Config) email.Sender {
	emailConfig := email.Config{
		Host:        cfg.Email.Host,
		Port:        cfg.Email.Port,
		Username:    cfg.Email.User,
		Password:    cfg.Email.Password,
		FromAddress: cfg.Email.FromAddress,
		FromName:    cfg.Email.FromName,
	}

	return smtp.New(emailConfig)
}

// printStats prints worker startup information.
func printStats(w *worker.Worker, log logger.Logger) {
	stats, err := w.Stats(context.Background())
	if err != nil {
		log.Warn("failed to get queue stats", logger.Err(err))
		return
	}

	log.Info("queue stats",
		logger.Int64("pending", stats.Pending),
		logger.Int64("running", stats.Running),
		logger.Int64("completed", stats.Completed),
		logger.Int64("failed", stats.Failed),
		logger.Int64("total", stats.Total),
	)
}
