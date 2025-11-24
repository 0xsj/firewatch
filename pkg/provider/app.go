package provider

import (
	"fmt"

	"github.com/0xsj/hexagonal-go/pkg/database"
	"github.com/0xsj/hexagonal-go/pkg/database/postgres"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger/console"
)

// ProvideLogger creates and configures the application logger.
func ProvideLogger() logger.Logger {
	return console.New(console.Options{
		Level:         logger.DebugLevel,
		ShowTimestamp: true,
		ShowCaller:    true,
		Colorize:      true,
		ColorScheme:   console.DefaultColorScheme,
	})
}

// ProvideDatabase provides a PostgreSQL database connection.
func ProvideDatabase(log logger.Logger) (database.DB, func(), error) {
	// Database configuration
	config := postgres.Config{
		Host:     "localhost",
		Port:     5436, // From docker-compose
		Database: "hexagonal_identity",
		User:     "hexagonal",
		Password: "hexagonal_dev_pass",
		SSLMode:  "disable",

		// Connection pool settings
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * 60, // 5 minutes
	}

	log.Info("connecting to database",
		logger.String("host", config.Host),
		logger.Int("port", config.Port),
		logger.String("database", config.Database),
	)

	// Connect to PostgreSQL
	db, err := postgres.Connect(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Info("database connected successfully")

	// Cleanup function
	cleanup := func() {
		log.Info("closing database connection")
		if err := db.Close(); err != nil {
			log.Error("failed to close database", logger.Err(err))
		}
	}

	return db, cleanup, nil
}
