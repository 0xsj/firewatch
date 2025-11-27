package config

import (
	"context"
	"fmt"

	pkgconfig "github.com/0xsj/hexagonal-go/pkg/config"
)

// Load loads worker configuration from environment variables.
//
// Environment variable prefixes:
//   - WORKER_* for worker settings
//   - DB_*     for database settings
//   - REDIS_*  for Redis settings
//   - SMTP_*   for email/SMTP settings
//   - EMAIL_*  for email sender settings
//   - LOG_*    for logger settings
func Load(ctx context.Context) (*Config, error) {
	cfg := DefaultConfig()

	// Load worker config from WORKER_* env vars
	workerSource := pkgconfig.NewEnvSource("WORKER")
	if err := workerSource.Load(ctx, &cfg.Worker); err != nil {
		return nil, fmt.Errorf("failed to load worker config: %w", err)
	}

	// Load database config from DB_* env vars
	dbSource := pkgconfig.NewEnvSource("DB")
	if err := dbSource.Load(ctx, &cfg.Database); err != nil {
		return nil, fmt.Errorf("failed to load database config: %w", err)
	}

	// Load redis config from REDIS_* env vars
	redisSource := pkgconfig.NewEnvSource("REDIS")
	if err := redisSource.Load(ctx, &cfg.Redis); err != nil {
		return nil, fmt.Errorf("failed to load redis config: %w", err)
	}

	// Load SMTP config from SMTP_* env vars
	smtpSource := pkgconfig.NewEnvSource("SMTP")
	if err := smtpSource.Load(ctx, &cfg.Email); err != nil {
		return nil, fmt.Errorf("failed to load email config: %w", err)
	}

	// Load email sender config from EMAIL_* env vars
	emailSource := pkgconfig.NewEnvSource("EMAIL")
	if err := emailSource.Load(ctx, &cfg.Email); err != nil {
		return nil, fmt.Errorf("failed to load email sender config: %w", err)
	}

	// Load logger config from LOG_* env vars
	logSource := pkgconfig.NewEnvSource("LOG")
	if err := logSource.Load(ctx, &cfg.Logger); err != nil {
		return nil, fmt.Errorf("failed to load logger config: %w", err)
	}

	return &cfg, nil
}

// MustLoad loads configuration and panics on error.
// Use only in main() or initialization code.
func MustLoad(ctx context.Context) *Config {
	cfg, err := Load(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}
