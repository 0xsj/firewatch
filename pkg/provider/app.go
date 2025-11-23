package provider

import (
	"github.com/0xsj/hexagonal-go/pkg/database"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger/console"
)

// ProvideLogger creates and configures the application logger.
func ProvideLogger() logger.Logger {
	return console.New(console.Options{
		Level:         logger.InfoLevel,
		ShowTimestamp: true,
		ShowCaller:    true,
		Colorize:      true,
		ColorScheme:   console.DefaultColorScheme,
	})
}

// ProvideDatabase provides a database connection.
// For now, returns nil (mock repository doesn't need it).
func ProvideDatabase(log logger.Logger) (database.DB, func(), error) {
	log.Info("database provider (mock mode - no real connection)")

	// No cleanup needed for mock
	cleanup := func() {}

	return nil, cleanup, nil
}

// Future infrastructure providers (no internal/ imports):
// func ProvideCache(log logger.Logger) (cache.Cache, func(), error)
// func ProvideJWTService() jwt.Service
// func ProvideMessageBus(log logger.Logger) (messaging.Publisher, func(), error)
