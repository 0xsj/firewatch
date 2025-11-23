package console

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// Logger is a colorized console logger for development.
// Outputs human-readable, beautifully formatted logs with ANSI colors.
//
// Features:
//   - Colorized log levels with high contrast
//   - Structured fields with clear formatting
//   - Automatic context field extraction
//   - Caller information (file:line)
//   - Pretty-printed timestamps
//
// Example output:
//
//	2024-01-15 10:30:45.123  INFO  user.Service.CreateUser  user created successfully
//	  tenant_id=acme-corp  user_id=usr_123  email=user@example.com  duration=45ms
type Logger struct {
	mu     sync.Mutex
	writer io.Writer
	level  logger.Level
	scheme ColorScheme

	// Pre-populated fields (from With())
	fields []logger.Field

	// Options
	showCaller    bool
	showTimestamp bool
	colorize      bool
}

// Options for creating a console logger.
type Options struct {
	// Writer is where logs are written (default: os.Stdout)
	Writer io.Writer

	// Level is the minimum log level (default: InfoLevel)
	Level logger.Level

	// ColorScheme defines colors (default: DefaultColorScheme)
	ColorScheme ColorScheme

	// ShowCaller includes file:line in output (default: true)
	ShowCaller bool

	// ShowTimestamp includes timestamp in output (default: true)
	ShowTimestamp bool

	// Colorize enables colors (default: true, auto-disable for non-TTY)
	Colorize bool
}

// DefaultOptions returns default console logger options.
func DefaultOptions() Options {
	return Options{
		Writer:        os.Stdout,
		Level:         logger.InfoLevel,
		ColorScheme:   DefaultColorScheme,
		ShowCaller:    true,
		ShowTimestamp: true,
		Colorize:      isTerminal(os.Stdout),
	}
}

// New creates a new colorized console logger.
func New(opts Options) logger.Logger {
	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}

	return &Logger{
		writer:        opts.Writer,
		level:         opts.Level,
		scheme:        opts.ColorScheme,
		showCaller:    opts.ShowCaller,
		showTimestamp: opts.ShowTimestamp,
		colorize:      opts.Colorize,
		fields:        make([]logger.Field, 0),
	}
}

// NewDefault creates a console logger with default options.
func NewDefault() logger.Logger {
	return New(DefaultOptions())
}

// NewWithLevel creates a console logger with a specific level.
func NewWithLevel(level logger.Level) logger.Logger {
	opts := DefaultOptions()
	opts.Level = level
	return New(opts)
}

// ============================================================================
// Logger Interface Implementation
// ============================================================================

func (l *Logger) Debug(msg string, fields ...logger.Field) {
	l.log(logger.DebugLevel, msg, fields)
}

func (l *Logger) Info(msg string, fields ...logger.Field) {
	l.log(logger.InfoLevel, msg, fields)
}

func (l *Logger) Warn(msg string, fields ...logger.Field) {
	l.log(logger.WarnLevel, msg, fields)
}

func (l *Logger) Error(msg string, fields ...logger.Field) {
	l.log(logger.ErrorLevel, msg, fields)
}

func (l *Logger) Fatal(msg string, fields ...logger.Field) {
	l.log(logger.FatalLevel, msg, fields)
	os.Exit(1)
}

func (l *Logger) With(fields ...logger.Field) logger.Logger {
	// Create a copy with additional fields
	return &Logger{
		writer:        l.writer,
		level:         l.level,
		scheme:        l.scheme,
		showCaller:    l.showCaller,
		showTimestamp: l.showTimestamp,
		colorize:      l.colorize,
		fields:        append(l.fields, fields...),
	}
}

func (l *Logger) WithContext(ctx context.Context) logger.Logger {
	// Extract context fields and add them
	contextFields := logger.ExtractContextFields(ctx)
	return l.With(contextFields...)
}

// ============================================================================
// Core Logging Logic
// ============================================================================

func (l *Logger) log(level logger.Level, msg string, fields []logger.Field) {
	// Check level
	if level < l.level {
		return
	}

	// Merge pre-populated fields with new fields
	allFields := append(l.fields, fields...)

	// Get caller info
	var caller string
	if l.showCaller {
		caller = getCaller(3) // Skip 3 frames: log -> Debug/Info/etc -> actual caller
	}

	// Format the log entry
	entry := l.formatEntry(level, msg, caller, allFields)

	// Write atomically
	l.mu.Lock()
	defer l.mu.Unlock()

	fmt.Fprintln(l.writer, entry)
}

// formatEntry formats a complete log entry with colors.
func (l *Logger) formatEntry(level logger.Level, msg string, caller string, fields []logger.Field) string {
	var b strings.Builder

	// Timestamp
	if l.showTimestamp {
		timestamp := time.Now().Format("2006-01-02 15:04:05.000")
		b.WriteString(l.colorizeComponent(l.scheme.Timestamp, timestamp))
		b.WriteString("  ")
	}

	// Level (with background color)
	levelStr := l.formatLevel(level)
	b.WriteString(levelStr)
	b.WriteString("  ")

	// Caller (file:line)
	if caller != "" {
		b.WriteString(l.colorizeComponent(l.scheme.Caller, caller))
		b.WriteString("  ")
	}

	// Message (bold white)
	b.WriteString(l.colorizeComponent(l.scheme.Message, msg))

	// Fields (INLINE - same line, not new line)
	if len(fields) > 0 {
		b.WriteString("  ") // Just spaces, not newline
		b.WriteString(l.formatFields(fields))
	}

	return b.String()
}

// formatLevel formats the log level with background color.
func (l *Logger) formatLevel(level logger.Level) string {
	var color string

	switch level {
	case logger.DebugLevel:
		color = l.scheme.DebugLevel
	case logger.InfoLevel:
		color = l.scheme.InfoLevel
	case logger.WarnLevel:
		color = l.scheme.WarnLevel
	case logger.ErrorLevel:
		color = l.scheme.ErrorLevel
	case logger.FatalLevel:
		color = l.scheme.FatalLevel
	default:
		color = ""
	}

	// Pad to 5 characters for alignment
	levelText := fmt.Sprintf(" %-5s ", level.ShortString())

	return l.colorizeComponent(color, levelText)
}

// formatFields formats all fields with proper coloring.
func (l *Logger) formatFields(fields []logger.Field) string {
	var b strings.Builder

	for i, field := range fields {
		if i > 0 {
			b.WriteString("  ") // Space between fields
		}

		// Determine color for this field
		valueColor := l.getFieldColor(field.Key)

		// Format: key=value
		b.WriteString(l.colorizeComponent(l.scheme.Key, field.Key))
		b.WriteString("=")
		b.WriteString(l.colorizeComponent(valueColor, l.formatValue(field.Value)))
	}

	return b.String()
}

// getFieldColor returns the appropriate color for a field based on its key.
func (l *Logger) getFieldColor(key string) string {
	switch key {
	case logger.FieldTenantID:
		return l.scheme.TenantID
	case logger.FieldUserID:
		return l.scheme.UserID
	case logger.FieldRequestID, logger.FieldCorrelationID:
		return l.scheme.RequestID
	case logger.FieldError:
		return l.scheme.Error
	default:
		return l.scheme.Value
	}
}

// formatValue formats a field value for display.
func (l *Logger) formatValue(value any) string {
	if value == nil {
		return "<nil>"
	}

	switch v := value.(type) {
	case string:
		// Check if string needs quoting (has spaces)
		if strings.Contains(v, " ") {
			return fmt.Sprintf("%q", v)
		}
		return v

	case time.Time:
		return v.Format(time.RFC3339)

	case time.Duration:
		return v.String()

	case error:
		return v.Error()

	case fmt.Stringer:
		return v.String()

	default:
		return fmt.Sprintf("%v", v)
	}
}

// colorizeComponent applies color if colorization is enabled.
func (l *Logger) colorizeComponent(color, text string) string {
	if !l.colorize {
		return text
	}
	return colorize(color, text)
}

// ============================================================================
// Helper Functions
// ============================================================================

// getCaller returns the caller information (file:line).
func getCaller(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "???"
	}

	// Get just the filename (not full path)
	file = filepath.Base(file)

	return fmt.Sprintf("%s:%d", file, line)
}

// isTerminal checks if the writer is a terminal (supports colors).
func isTerminal(w io.Writer) bool {
	// Check if it's stdout/stderr
	if f, ok := w.(*os.File); ok {
		// Simple check: if it's stdout or stderr, assume terminal
		// For more robust check, use golang.org/x/term
		return f == os.Stdout || f == os.Stderr
	}
	return false
}
