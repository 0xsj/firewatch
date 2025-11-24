//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/0xsj/hexagonal-go/cmd/api/config"
	"github.com/0xsj/hexagonal-go/internal/audit"
	"github.com/0xsj/hexagonal-go/internal/identity"
	"github.com/0xsj/hexagonal-go/internal/notifications"
	"github.com/0xsj/hexagonal-go/pkg/database/postgres"
	"github.com/0xsj/hexagonal-go/pkg/email"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger/console"
	"github.com/0xsj/hexagonal-go/pkg/provider"
)

// InitializeApp wires up the entire application.
func InitializeApp(cfg *config.AppConfig) (*App, func(), error) {
	wire.Build(
		// Config extractors
		ProvidePostgresConfig,
		ProvideLoggerOptions,
		ProvideEmailConfig,

		// Infrastructure (from pkg/provider)
		provider.ProvideLogger,
		provider.ProvideDatabase,
		provider.ProvideEventBus,
		provider.ProvideEmailSender,

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
