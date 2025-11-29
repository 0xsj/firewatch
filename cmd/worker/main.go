package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/0xsj/hexagonal-go/cmd/worker/config"
	notificationjobs "github.com/0xsj/hexagonal-go/internal/notifications/application/jobs"
	"github.com/0xsj/hexagonal-go/pkg/database/postgres"
	"github.com/0xsj/hexagonal-go/pkg/email"
	"github.com/0xsj/hexagonal-go/pkg/email/smtp"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger/console"
	"github.com/0xsj/hexagonal-go/pkg/worker"
	"github.com/0xsj/hexagonal-go/pkg/worker/memory"
	postgresqueue "github.com/0xsj/hexagonal-go/pkg/worker/postgres"
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
	// Initialize Database (for PostgreSQL queue)
	// ========================================================================
	var db *postgres.PostgresDB
	if cfg.Worker.QueueType == "postgres" {
		dbConfig := postgres.Config{
			Host:     cfg.Database.Host,
			Port:     cfg.Database.Port,
			User:     cfg.Database.User,
			Password: cfg.Database.Password,
			Database: cfg.Database.Database,
			SSLMode:  cfg.Database.SSLMode,
		}

		db, err = postgres.Connect(dbConfig)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer db.Close()

		log.Info("connected to database",
			logger.String("host", cfg.Database.Host),
			logger.String("database", cfg.Database.Database),
		)
	}

	// ========================================================================
	// Initialize Queue
	// ========================================================================
	queue, err := initQueue(cfg, db)
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
	// Register Job Handlers
	// ========================================================================
	registerHandlers(w, emailSender, log)

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
	// Print Stats and Start Worker
	// ========================================================================
	printStats(w, log)
	log.Info("worker started, press Ctrl+C to stop")

	if err := w.Start(ctx); err != nil {
		return fmt.Errorf("worker error: %w", err)
	}

	log.Info("worker stopped")
	return nil
}

// initQueue initializes the job queue based on configuration.
func initQueue(cfg *config.Config, db *postgres.PostgresDB) (worker.Queue, error) {
	switch cfg.Worker.QueueType {
	case "memory":
		return memory.NewQueue(), nil
	case "postgres":
		if db == nil {
			return nil, fmt.Errorf("database connection required for postgres queue")
		}
		return postgresqueue.NewQueue(db), nil
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

// registerHandlers registers all job handlers with the worker.
func registerHandlers(w *worker.Worker, emailSender email.Sender, log logger.Logger) {
	// Notification: Send Email
	sendEmailHandler := notificationjobs.NewSendEmailHandler(emailSender, log)
	w.Register(notificationjobs.JobTypeSendEmail, sendEmailHandler)

	log.Info("job handlers registered",
		logger.String("handler", notificationjobs.JobTypeSendEmail),
	)
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
		logger.Int64("retrying", stats.Retrying),
		logger.Int64("total", stats.Total),
	)
}
