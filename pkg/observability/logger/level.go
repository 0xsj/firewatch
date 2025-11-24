package logger

import "strings"

// Level represents the severity of a log message.
type Level int8

const (
	// DebugLevel is for detailed debugging information.
	DebugLevel Level = iota - 1

	// InfoLevel is for general informational messages.
	InfoLevel

	// WarnLevel is for warning messages.
	WarnLevel

	// ErrorLevel is for error messages.
	ErrorLevel

	// FatalLevel is for fatal errors (program exits).
	FatalLevel
)

// String returns the string representation of the level.
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// ShortString returns a short (4-char) representation.
func (l Level) ShortString() string {
	switch l {
	case DebugLevel:
		return "DEBG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERRO"
	case FatalLevel:
		return "FATL"
	default:
		return "UNKN"
	}
}

// ParseLevel parses a level from string.
func ParseLevel(s string) Level {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "DEBUG", "DEBG":
		return DebugLevel
	case "INFO":
		return InfoLevel
	case "WARN", "WARNING":
		return WarnLevel
	case "ERROR", "ERR", "ERRO":
		return ErrorLevel
	case "FATAL", "FATL":
		return FatalLevel
	default:
		return InfoLevel
	}
}

func (l *Level) UnmarshalText(text []byte) error {
	*l = ParseLevel(string(text))
	return nil
}
