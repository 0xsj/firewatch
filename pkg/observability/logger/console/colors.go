package console

// ANSI color codes for terminal output
const (
	// Reset
	Reset = "\033[0m"

	// Text styles
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"

	// Foreground colors (normal)
	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	// Foreground colors (bright)
	BrightBlack   = "\033[90m"
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"

	// Background colors
	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"
)

// ColorScheme defines colors for different log elements.
type ColorScheme struct {
	// Level colors
	DebugLevel string
	InfoLevel  string
	WarnLevel  string
	ErrorLevel string
	FatalLevel string

	// Component colors
	Timestamp string
	Caller    string
	Message   string

	// Field colors
	Key   string
	Value string

	// Special field colors
	TenantID  string
	UserID    string
	RequestID string
	Error     string
}

// DefaultColorScheme returns a color scheme with high contrast.
var DefaultColorScheme = ColorScheme{
	// Level colors (background + white text for visibility)
	DebugLevel: BgBlue + BrightWhite + Bold,
	InfoLevel:  BgGreen + BrightWhite + Bold,
	WarnLevel:  BgYellow + Black + Bold,
	ErrorLevel: BgRed + BrightWhite + Bold,
	FatalLevel: BgMagenta + BrightWhite + Bold,

	// Component colors
	Timestamp: BrightBlack,        // Dim gray
	Caller:    BrightBlue,         // Bright blue
	Message:   BrightWhite + Bold, // Bold white

	// Field colors
	Key:   Cyan,        // Cyan for keys
	Value: BrightWhite, // Bright white for values

	// Special field colors
	TenantID:  BrightMagenta,    // Magenta for tenant
	UserID:    BrightYellow,     // Yellow for user
	RequestID: BrightCyan,       // Cyan for request
	Error:     BrightRed + Bold, // Bold red for errors
}

// DarkColorScheme is optimized for dark terminals.
var DarkColorScheme = ColorScheme{
	DebugLevel: BgBlue + BrightWhite + Bold,
	InfoLevel:  BgGreen + Black + Bold,
	WarnLevel:  BgYellow + Black + Bold,
	ErrorLevel: BgRed + BrightWhite + Bold,
	FatalLevel: BgMagenta + BrightWhite + Bold,

	Timestamp: BrightBlack,
	Caller:    BrightBlue,
	Message:   BrightWhite,

	Key:   BrightCyan,
	Value: White,

	TenantID:  BrightMagenta,
	UserID:    BrightYellow,
	RequestID: BrightCyan,
	Error:     BrightRed,
}

// LightColorScheme is optimized for light terminals.
var LightColorScheme = ColorScheme{
	DebugLevel: BgBlue + White + Bold,
	InfoLevel:  BgGreen + White + Bold,
	WarnLevel:  BgYellow + Black + Bold,
	ErrorLevel: BgRed + White + Bold,
	FatalLevel: BgMagenta + White + Bold,

	Timestamp: Black,
	Caller:    Blue,
	Message:   Black + Bold,

	Key:   Blue,
	Value: Black,

	TenantID:  Magenta,
	UserID:    Yellow,
	RequestID: Cyan,
	Error:     Red + Bold,
}

// colorize wraps text with color codes.
func colorize(color, text string) string {
	if color == "" {
		return text
	}
	return color + text + Reset
}
