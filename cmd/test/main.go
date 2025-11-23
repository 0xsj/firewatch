package main

import (
	"context"
	"fmt"
	"time"

	pkgcontext "github.com/0xsj/hexagonal-go/pkg/context"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger/console"
)

func main() {
	// Create a beautiful colorized console logger
	log := console.NewDefault()

	// Basic logging with different levels
	log.Info("🚀 Server starting",
		logger.String("host", "0.0.0.0"),
		logger.Int("port", 8080),
		logger.String("env", "development"),
	)

	log.Debug("Debug information",
		logger.String("config_file", "config.yaml"),
		logger.Bool("hot_reload", true),
	)

	log.Warn("Cache miss detected",
		logger.String("key", "user:123"),
		logger.Duration("latency", 45*time.Millisecond),
	)

	// With pre-populated fields
	serviceLog := log.With(
		logger.String("component", "user-service"),
		logger.String("version", "1.0.0"),
	)

	serviceLog.Info("Service initialized",
		logger.Int("workers", 10),
		logger.String("mode", "async"),
	)

	// With context (multi-tenancy + auth)
	ctx := context.Background()
	ctx = pkgcontext.WithTenantID(ctx, "acme-corp")
	ctx = pkgcontext.WithUserID(ctx, "usr_abc123")
	ctx = pkgcontext.WithRequestID(ctx, "req_xyz789")

	contextLog := log.WithContext(ctx)
	contextLog.Info("User authenticated successfully",
		logger.String("email", "john@acme-corp.com"),
		logger.String("ip_address", "192.168.1.100"),
	)

	contextLog.Info("Creating new order",
		logger.String("order_id", "ord_001"),
		logger.Float64("amount", 99.99),
		logger.String("currency", "USD"),
	)

	// Error logging
	log.Error("Database connection failed",
		logger.String("host", "localhost"),
		logger.Int("port", 5432),
		logger.Err(fmt.Errorf("connection refused")),
		logger.Int("retry_count", 3),
	)

	// Nested logger with more context
	requestLog := log.With(
		logger.String("request_id", "req_demo_001"),
		logger.String("method", "POST"),
		logger.String("path", "/api/v1/users"),
	).WithContext(ctx)

	requestLog.Info("Processing request",
		logger.Duration("duration", 127*time.Millisecond),
		logger.Int("status_code", 201),
	)

	requestLog.Warn("Rate limit approaching",
		logger.Int("current", 95),
		logger.Int("limit", 100),
		logger.String("window", "1m"),
	)

	// Simulate different log levels
	log.Debug("Detailed debugging information", logger.String("trace", "enabled"))
	log.Info("Normal operation", logger.String("status", "healthy"))
	log.Warn("Non-critical issue detected", logger.String("issue", "slow query"))
	log.Error("Critical error occurred", logger.String("severity", "high"))

	log.Info("✅ Logger test complete")
}
