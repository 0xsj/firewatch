package cache

import "time"

// Config holds cache configuration.
type Config struct {
	// Enabled determines if caching is enabled.
	Enabled bool `env:"ENABLED"`

	// Redis connection settings
	Host     string `env:"HOST"`
	Port     int    `env:"PORT"`
	Password string `env:"PASSWORD"`
	DB       int    `env:"DB"`

	// Connection pool settings
	PoolSize     int           `env:"POOL_SIZE"`
	MinIdleConns int           `env:"MIN_IDLE_CONNS"`
	MaxRetries   int           `env:"MAX_RETRIES"`
	DialTimeout  time.Duration `env:"DIAL_TIMEOUT"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT"`
}

// DefaultConfig returns default cache configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:      true,
		Host:         "localhost",
		Port:         6379,
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}
