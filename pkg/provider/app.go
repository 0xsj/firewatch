package provider

import (
	"fmt"

	"github.com/0xsj/hexagonal-go/pkg/database"
	"github.com/0xsj/hexagonal-go/pkg/database/postgres"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger/console"
)

// ProvideLogger creates and configures the application logger.
func ProvideLogger(opts console.Options) logger.Logger {
	return console.New(opts)
}

// ProvideDatabase provides a PostgreSQL database connection.
func ProvideDatabase(config postgres.Config, log logger.Logger) (database.DB, func(), error) {
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
