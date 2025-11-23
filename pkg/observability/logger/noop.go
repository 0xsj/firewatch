package logger

import "context"

// NoopLogger is a logger that does nothing.
// Useful for testing or disabling logging.
type NoopLogger struct{}

// NewNoop creates a new no-op logger.
func NewNoop() Logger {
	return &NoopLogger{}
}

func (l *NoopLogger) Debug(msg string, fields ...Field) {}
func (l *NoopLogger) Info(msg string, fields ...Field)  {}
func (l *NoopLogger) Warn(msg string, fields ...Field)  {}
func (l *NoopLogger) Error(msg string, fields ...Field) {}
func (l *NoopLogger) Fatal(msg string, fields ...Field) {}

func (l *NoopLogger) With(fields ...Field) Logger {
	return l
}

func (l *NoopLogger) WithContext(ctx context.Context) Logger {
	return l
}
