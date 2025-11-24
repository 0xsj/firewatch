package config

import (
	"github.com/0xsj/hexagonal-go/pkg/database/postgres"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger/console"
)

// AppConfig holds all application configuration.
// Configuration is loaded from environment variables with prefixes:
//   - DB_*     for database settings
//   - LOG_*    for logger settings
//   - SERVER_* for HTTP server settings
type AppConfig struct {
	Database postgres.Config
	Logger   console.Options
	Server   ServerConfig
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Host string `env:"HOST"`
	Port int    `env:"PORT"`
}

// DefaultAppConfig returns the default application configuration.
func DefaultAppConfig() AppConfig {
	return AppConfig{
		Database: postgres.DefaultConfig(),
		Logger:   console.DefaultOptions(),
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
	}
}
