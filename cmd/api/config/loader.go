package config

import (
	"context"
	"fmt"

	pkgconfig "github.com/0xsj/hexagonal-go/pkg/config"
)

// Load loads application configuration from environment variables.
// It starts with defaults and overrides with environment variables.
//
// Environment variable prefixes:
//   - DB_*       for database settings
//   - LOG_*      for logger settings
//   - SERVER_*   for HTTP server settings
//   - TENANCY_*  for multi-tenancy settings
func Load(ctx context.Context) (*AppConfig, error) {
	// Start with defaults
	cfg := DefaultAppConfig()

	// Load database config from DB_* env vars
	dbSource := pkgconfig.NewEnvSource("DB")
	if err := dbSource.Load(ctx, &cfg.Database); err != nil {
		return nil, fmt.Errorf("failed to load database config: %w", err)
	}

	// Load logger config from LOG_* env vars
	logSource := pkgconfig.NewEnvSource("LOG")
	if err := logSource.Load(ctx, &cfg.Logger); err != nil {
		return nil, fmt.Errorf("failed to load logger config: %w", err)
	}

	// Load server config from SERVER_* env vars
	serverSource := pkgconfig.NewEnvSource("SERVER")
	if err := serverSource.Load(ctx, &cfg.Server); err != nil {
		return nil, fmt.Errorf("failed to load server config: %w", err)
	}

	// Load tenancy config from TENANCY_* env vars
	tenancySource := pkgconfig.NewEnvSource("TENANCY")
	if err := tenancySource.Load(ctx, &cfg.Tenancy); err != nil {
		return nil, fmt.Errorf("failed to load tenancy config: %w", err)
	}

	return &cfg, nil
}

// MustLoad loads configuration and panics on error.
// Use only in main() or initialization code.
func MustLoad(ctx context.Context) *AppConfig {
	cfg, err := Load(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}
