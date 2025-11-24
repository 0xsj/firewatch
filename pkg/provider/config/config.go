package config

import (
	"github.com/0xsj/hexagonal-go/pkg/database/postgres"
	"github.com/0xsj/hexagonal-go/pkg/email"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger/console"
)

// AppConfig holds all application configuration.
// Configuration is loaded from environment variables with prefixes:
//   - DB_*        for database settings
//   - LOG_*       for logger settings
//   - SERVER_*    for HTTP server settings
//   - EMAIL_*     for email/SMTP settings
//   - MESSAGING_* for messaging settings (future)
type AppConfig struct {
	Database  postgres.Config
	Logger    console.Options
	Server    ServerConfig
	Email     email.Config
	Messaging MessagingConfig
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Host string `env:"HOST"`
	Port int    `env:"PORT"`
}

// MessagingConfig holds message broker configuration.
// Currently using in-memory bus which requires no configuration.
// When adding a real broker (RabbitMQ, Kafka, etc.), add fields here:
//   - Type     string `env:"TYPE"`     // "memory", "rabbitmq", "kafka"
//   - Host     string `env:"HOST"`
//   - Port     int    `env:"PORT"`
//   - User     string `env:"USER"`
//   - Password string `env:"PASSWORD"`
//   - Vhost    string `env:"VHOST"`    // RabbitMQ virtual host
type MessagingConfig struct {
	// Placeholder for future broker configuration
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
		Email:     email.DefaultConfig(),
		Messaging: MessagingConfig{},
	}
}
