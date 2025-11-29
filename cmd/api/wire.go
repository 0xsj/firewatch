// //go:build wireinject
// // +build wireinject

// package main

// import (
// 	"context"

// 	"github.com/google/wire"

// 	"github.com/0xsj/hexagonal-go/cmd/api/config"
// 	"github.com/0xsj/hexagonal-go/internal/audit"
// 	"github.com/0xsj/hexagonal-go/internal/email"
// 	"github.com/0xsj/hexagonal-go/internal/identity"
// 	"github.com/0xsj/hexagonal-go/internal/notifications"
// 	"github.com/0xsj/hexagonal-go/internal/tenant"
// 	"github.com/0xsj/hexagonal-go/pkg/cache"
// 	"github.com/0xsj/hexagonal-go/pkg/database"
// 	"github.com/0xsj/hexagonal-go/pkg/database/postgres"
// 	pkgemail "github.com/0xsj/hexagonal-go/pkg/email"
// 	"github.com/0xsj/hexagonal-go/pkg/http/middleware"
// 	"github.com/0xsj/hexagonal-go/pkg/observability/logger/console"
// 	"github.com/0xsj/hexagonal-go/pkg/observability/metrics"
// 	"github.com/0xsj/hexagonal-go/pkg/observability/tracing"
// 	"github.com/0xsj/hexagonal-go/pkg/provider"
// 	"github.com/0xsj/hexagonal-go/pkg/security/jwt"
// 	"github.com/0xsj/hexagonal-go/pkg/storage"
// 	"github.com/0xsj/hexagonal-go/pkg/worker"
// 	postgresqueue "github.com/0xsj/hexagonal-go/pkg/worker/postgres"
// )

// // InitializeApp wires up the entire application.
// func InitializeApp(ctx context.Context, cfg *config.AppConfig) (*App, func(), error) {
// 	wire.Build(
// 		// Config extractors
// 		ProvidePostgresConfig,
// 		ProvideLoggerOptions,
// 		ProvideEmailConfig,
// 		ProvideMetricsConfig,
// 		ProvideTracingConfig,
// 		ProvideJWTConfig,
// 		ProvideCacheConfig,
// 		ProvideStorageConfig,

// 		// Infrastructure (from pkg/provider)
// 		provider.ProvideLogger,
// 		provider.ProvideDatabase,
// 		provider.ProvideEventBus,
// 		provider.ProvideDomainEventPublisher,
// 		provider.ProvideEmailSender,
// 		provider.ProvideMetricsProvider,
// 		provider.ProvideTracingProvider,
// 		provider.ProvideJWTService,
// 		provider.ProvideCache,
// 		provider.ProvideStorage,

// 		// Job Queue
// 		ProvideJobQueue,
// 		wire.Bind(new(worker.Queue), new(*postgresqueue.Queue)),

// 		// HTTP Metrics
// 		middleware.NewHTTPMetrics,

// 		// Identity domain
// 		identity.IdentitySet,

// 		// Tenant domain
// 		tenant.TenantSet,

// 		// Audit domain
// 		audit.AuditSet,

// 		// Notifications domain
// 		notifications.NotificationsSet,

// 		// Email domain
// 		email.EmailSet,

// 		// Wire the App struct
// 		wire.Struct(new(App), "*"),
// 	)
// 	return &App{}, nil, nil
// }

// // ProvideJobQueue creates a PostgreSQL-backed job queue.
// func ProvideJobQueue(db database.DB) *postgresqueue.Queue {
// 	return postgresqueue.NewQueue(db)
// }

// // ProvidePostgresConfig extracts database config from AppConfig.
// func ProvidePostgresConfig(cfg *config.AppConfig) postgres.Config {
// 	return cfg.Database
// }

// // ProvideLoggerOptions extracts logger options from AppConfig.
// func ProvideLoggerOptions(cfg *config.AppConfig) console.Options {
// 	return cfg.Logger
// }

// // ProvideEmailConfig extracts email config from AppConfig.
// func ProvideEmailConfig(cfg *config.AppConfig) pkgemail.Config {
// 	return cfg.Email
// }

// // ProvideMetricsConfig extracts metrics config from AppConfig.
// func ProvideMetricsConfig(cfg *config.AppConfig) metrics.Config {
// 	return cfg.Metrics
// }

// // ProvideTracingConfig extracts tracing config from AppConfig.
// func ProvideTracingConfig(cfg *config.AppConfig) tracing.Config {
// 	return cfg.Tracing
// }

// // ProvideJWTConfig extracts JWT config from AppConfig.
// func ProvideJWTConfig(cfg *config.AppConfig) jwt.Config {
// 	return cfg.JWT
// }

// // ProvideCacheConfig extracts cache config from AppConfig.
// func ProvideCacheConfig(cfg *config.AppConfig) cache.Config {
// 	return cfg.Cache
// }

// // ProvideStorageConfig extracts storage config from AppConfig.
// func ProvideStorageConfig(cfg *config.AppConfig) storage.Config {
// 	return cfg.Storage
// }

//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"github.com/google/wire"

	"github.com/0xsj/hexagonal-go/cmd/api/config"
	"github.com/0xsj/hexagonal-go/internal/audit"
	"github.com/0xsj/hexagonal-go/internal/email"
	"github.com/0xsj/hexagonal-go/internal/flags"
	"github.com/0xsj/hexagonal-go/internal/identity"
	"github.com/0xsj/hexagonal-go/internal/notifications"
	"github.com/0xsj/hexagonal-go/internal/tenant"
	"github.com/0xsj/hexagonal-go/pkg/cache"
	"github.com/0xsj/hexagonal-go/pkg/database"
	"github.com/0xsj/hexagonal-go/pkg/database/postgres"
	pkgemail "github.com/0xsj/hexagonal-go/pkg/email"
	"github.com/0xsj/hexagonal-go/pkg/http/middleware"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger/console"
	"github.com/0xsj/hexagonal-go/pkg/observability/metrics"
	"github.com/0xsj/hexagonal-go/pkg/observability/tracing"
	"github.com/0xsj/hexagonal-go/pkg/provider"
	"github.com/0xsj/hexagonal-go/pkg/security/jwt"
	"github.com/0xsj/hexagonal-go/pkg/storage"
	"github.com/0xsj/hexagonal-go/pkg/worker"
	postgresqueue "github.com/0xsj/hexagonal-go/pkg/worker/postgres"
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
		ProvideStorageConfig,

		// Infrastructure (from pkg/provider)
		provider.ProvideLogger,
		provider.ProvideDatabase,
		provider.ProvideEventBus,
		provider.ProvideDomainEventPublisher,
		provider.ProvideEmailSender,
		provider.ProvideMetricsProvider,
		provider.ProvideTracingProvider,
		provider.ProvideJWTService,
		provider.ProvideCache,
		provider.ProvideStorage,

		// Job Queue
		ProvideJobQueue,
		wire.Bind(new(worker.Queue), new(*postgresqueue.Queue)),

		// HTTP Metrics
		middleware.NewHTTPMetrics,

		// Identity domain
		identity.IdentitySet,

		// Tenant domain
		tenant.TenantSet,

		// Audit domain
		audit.AuditSet,

		// Notifications domain
		notifications.NotificationsSet,

		// Email domain
		email.EmailSet,

		// Flags domain
		flags.FlagsSet,

		// Wire the App struct
		wire.Struct(new(App), "*"),
	)
	return &App{}, nil, nil
}

// ProvideJobQueue creates a PostgreSQL-backed job queue.
func ProvideJobQueue(db database.DB) *postgresqueue.Queue {
	return postgresqueue.NewQueue(db)
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
func ProvideEmailConfig(cfg *config.AppConfig) pkgemail.Config {
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

// ProvideStorageConfig extracts storage config from AppConfig.
func ProvideStorageConfig(cfg *config.AppConfig) storage.Config {
	return cfg.Storage
}
