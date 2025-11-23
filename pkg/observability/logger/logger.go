package logger

import (
	"context"
)

// Logger is the core logging interface.
// All loggers must implement this interface for dependency injection.
type Logger interface {
	// Debug logs a debug message with structured fields.
	Debug(msg string, fields ...Field)

	// Info logs an informational message with structured fields.
	Info(msg string, fields ...Field)

	// Warn logs a warning message with structured fields.
	Warn(msg string, fields ...Field)

	// Error logs an error message with structured fields.
	Error(msg string, fields ...Field)

	// Fatal logs a fatal error and exits the application.
	Fatal(msg string, fields ...Field)

	// With returns a new logger with additional fields pre-populated.
	// Fields are added to all subsequent log calls.
	With(fields ...Field) Logger

	// WithContext returns a logger that extracts fields from context.
	// Automatically includes: tenant_id, user_id, request_id, correlation_id
	WithContext(ctx context.Context) Logger
}

// ContextLogger extends Logger with context-aware methods.
// Automatically extracts tenant_id, user_id, request_id from context.
type ContextLogger interface {
	Logger

	// DebugContext logs with context fields automatically included.
	DebugContext(ctx context.Context, msg string, fields ...Field)

	// InfoContext logs with context fields automatically included.
	InfoContext(ctx context.Context, msg string, fields ...Field)

	// WarnContext logs with context fields automatically included.
	WarnContext(ctx context.Context, msg string, fields ...Field)

	// ErrorContext logs with context fields automatically included.
	ErrorContext(ctx context.Context, msg string, fields ...Field)
}
