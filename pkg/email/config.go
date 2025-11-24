package email

import "time"

// Config holds email sender configuration.
type Config struct {
	// SMTP server settings
	Host string `env:"HOST"`
	Port int    `env:"PORT"`

	// Authentication (optional for Mailpit)
	Username string `env:"USERNAME"`
	Password string `env:"PASSWORD"`

	// Default sender address
	FromAddress string `env:"FROM_ADDRESS"`
	FromName    string `env:"FROM_NAME"`

	// TLS settings
	TLS         bool `env:"TLS"`
	InsecureTLS bool `env:"INSECURE_TLS"`

	// Timeouts
	Timeout time.Duration `env:"TIMEOUT"`
}

// DefaultConfig returns default email configuration.
// Configured for local Mailpit instance.
func DefaultConfig() Config {
	return Config{
		Host:        "localhost",
		Port:        1025,
		Username:    "",
		Password:    "",
		FromAddress: "noreply@hexagonal.local",
		FromName:    "Hexagonal App",
		TLS:         false,
		InsecureTLS: false,
		Timeout:     10 * time.Second,
	}
}
