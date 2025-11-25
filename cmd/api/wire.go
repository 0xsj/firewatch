//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"github.com/google/wire"

	"github.com/0xsj/hexagonal-go/cmd/api/config"
	"github.com/0xsj/hexagonal-go/internal/audit"
	"github.com/0xsj/hexagonal-go/internal/identity"
	"github.com/0xsj/hexagonal-go/internal/notifications"
	"github.com/0xsj/hexagonal-go/pkg/cache"
	"github.com/0xsj/hexagonal-go/pkg/database/postgres"
	"github.com/0xsj/hexagonal-go/pkg/email"
	"github.com/0xsj/hexagonal-go/pkg/http/middleware"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger/console"
	"github.com/0xsj/hexagonal-go/pkg/observability/metrics"
	"github.com/0xsj/hexagonal-go/pkg/observability/tracing"
	"github.com/0xsj/hexagonal-go/pkg/provider"
	"github.com/0xsj/hexagonal-go/pkg/security/jwt"
)

// InitializeApp wires up the entire application.
func InitializeApp(ctx context.Context, cfg *config.AppConfig) (*App, func(), error) {
	wire.Build(
		// Config extractors
		ProvidePostgresConfig,
		ProvideLoggerOptions,
		ProvideEmailConfig,
		ProvideMetricsConfig,
		ProvideTracingConfig,
		ProvideJWTConfig,
		ProvideCacheConfig,

		// Infrastructure (from pkg/provider)
		provider.ProvideLogger,
		provider.ProvideDatabase,
		provider.ProvideEventBus,
		provider.ProvideEmailSender,
		provider.ProvideMetricsProvider,
		provider.ProvideTracingProvider,
		provider.ProvideJWTService,
		provider.ProvideCache,

		// HTTP Metrics
		middleware.NewHTTPMetrics,

		// Identity domain
		identity.IdentitySet,

		// Audit domain
		audit.AuditSet,

		// Notifications domain
		notifications.NotificationsSet,

		// Wire the App struct
		wire.Struct(new(App), "*"),
	)
	return &App{}, nil, nil
}

// ProvidePostgresConfig extracts database config from AppConfig.
func ProvidePostgresConfig(cfg *config.AppConfig) postgres.Config {
	return cfg.Database
}

// ProvideLoggerOptions extracts logger options from AppConfig.
func ProvideLoggerOptions(cfg *config.AppConfig) console.Options {
	return cfg.Logger
}

// ProvideEmailConfig extracts email config from AppConfig.
func ProvideEmailConfig(cfg *config.AppConfig) email.Config {
	return cfg.Email
}

// ProvideMetricsConfig extracts metrics config from AppConfig.
func ProvideMetricsConfig(cfg *config.AppConfig) metrics.Config {
	return cfg.Metrics
}

// ProvideTracingConfig extracts tracing config from AppConfig.
func ProvideTracingConfig(cfg *config.AppConfig) tracing.Config {
	return cfg.Tracing
}

// ProvideJWTConfig extracts JWT config from AppConfig.
func ProvideJWTConfig(cfg *config.AppConfig) jwt.Config {
	return cfg.JWT
}

// ProvideCacheConfig extracts cache config from AppConfig.
func ProvideCacheConfig(cfg *config.AppConfig) cache.Config {
	return cfg.Cache
}
