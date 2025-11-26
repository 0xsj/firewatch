package config

import (
	"github.com/0xsj/hexagonal-go/pkg/cache"
	"github.com/0xsj/hexagonal-go/pkg/database/postgres"
	"github.com/0xsj/hexagonal-go/pkg/email"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger/console"
	"github.com/0xsj/hexagonal-go/pkg/observability/metrics"
	"github.com/0xsj/hexagonal-go/pkg/observability/tracing"
	"github.com/0xsj/hexagonal-go/pkg/security/jwt"
	"github.com/0xsj/hexagonal-go/pkg/storage"
)

type AppConfig struct {
	Database postgres.Config
	Logger   console.Options
	Server   ServerConfig
	Email    email.Config
	Metrics  metrics.Config
	Tracing  tracing.Config
	JWT      jwt.Config
	Cache    cache.Config
	Storage  storage.Config
}

type ServerConfig struct {
	Host string `env:"HOST"`
	Port int    `env:"PORT"`
}

func DefaultAppConfig() AppConfig {
	return AppConfig{
		Database: postgres.DefaultConfig(),
		Logger:   console.DefaultOptions(),
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Email:   email.DefaultConfig(),
		Metrics: metrics.DefaultConfig(),
		Tracing: tracing.DefaultConfig(),
		JWT:     jwt.DefaultConfig(),
		Cache:   cache.DefaultConfig(),
		Storage: storage.DefaultConfig(),
	}
}
