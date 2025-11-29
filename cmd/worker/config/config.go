package config

import "time"

// Config holds all worker configuration.
type Config struct {
	Worker   WorkerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Email    EmailConfig
	Logger   LoggerConfig
}

// DefaultConfig returns configuration with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Worker:   DefaultWorkerConfig(),
		Database: DefaultDatabaseConfig(),
		Redis:    DefaultRedisConfig(),
		Email:    DefaultEmailConfig(),
		Logger:   DefaultLoggerConfig(),
	}
}

// WorkerConfig holds worker-specific settings.
type WorkerConfig struct {
	Concurrency     int           `config:"CONCURRENCY"`
	PollInterval    time.Duration `config:"POLL_INTERVAL"`
	MaxRetries      int           `config:"MAX_RETRIES"`
	RetryBackoff    time.Duration `config:"RETRY_BACKOFF"`
	MaxBackoff      time.Duration `config:"MAX_BACKOFF"`
	JobTimeout      time.Duration `config:"JOB_TIMEOUT"`
	ShutdownTimeout time.Duration `config:"SHUTDOWN_TIMEOUT"`
	BatchSize       int           `config:"BATCH_SIZE"`
	QueueType       string        `config:"QUEUE_TYPE"`
}

// DefaultWorkerConfig returns default worker settings.
func DefaultWorkerConfig() WorkerConfig {
	return WorkerConfig{
		Concurrency:     5,
		PollInterval:    1 * time.Second,
		MaxRetries:      3,
		RetryBackoff:    5 * time.Second,
		MaxBackoff:      5 * time.Minute,
		JobTimeout:      5 * time.Minute,
		ShutdownTimeout: 30 * time.Second,
		BatchSize:       10,
		QueueType:       "postgres",
	}
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	Host     string `config:"HOST"`
	Port     int    `config:"PORT"`
	User     string `config:"USER"`
	Password string `config:"PASSWORD"`
	Database string `config:"DATABASE"`
	SSLMode  string `config:"SSL_MODE"`
}

// DefaultDatabaseConfig returns default database settings.
func DefaultDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     "localhost",
		Port:     5436,
		User:     "hexagonal",
		Password: "hexagonal_dev_pass",
		Database: "hexagonal_go",
		SSLMode:  "disable",
	}
}

// DSN returns the database connection string.
func (c DatabaseConfig) DSN() string {
	return "host=" + c.Host +
		" port=" + itoa(c.Port) +
		" user=" + c.User +
		" password=" + c.Password +
		" dbname=" + c.Database +
		" sslmode=" + c.SSLMode
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Host     string `config:"HOST"`
	Port     int    `config:"PORT"`
	Password string `config:"PASSWORD"`
	DB       int    `config:"DB"`
}

// DefaultRedisConfig returns default Redis settings.
func DefaultRedisConfig() RedisConfig {
	return RedisConfig{
		Host:     "localhost",
		Port:     6383,
		Password: "",
		DB:       0,
	}
}

// Addr returns the Redis address.
func (c RedisConfig) Addr() string {
	return c.Host + ":" + itoa(c.Port)
}

// EmailConfig holds email/SMTP settings.
type EmailConfig struct {
	Host        string `config:"HOST"`
	Port        int    `config:"PORT"`
	User        string `config:"USER"`
	Password    string `config:"PASSWORD"`
	FromAddress string `config:"FROM_ADDRESS"`
	FromName    string `config:"FROM_NAME"`
}

// DefaultEmailConfig returns default email settings.
func DefaultEmailConfig() EmailConfig {
	return EmailConfig{
		Host:        "localhost",
		Port:        1027,
		User:        "",
		Password:    "",
		FromAddress: "noreply@hexagonal.app",
		FromName:    "Hexagonal App",
	}
}

// LoggerConfig holds logging settings.
type LoggerConfig struct {
	Level  string `config:"LEVEL"`
	Format string `config:"FORMAT"`
}

// DefaultLoggerConfig returns default logger settings.
func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Level:  "info",
		Format: "console",
	}
}

// itoa converts int to string.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}

	var b [20]byte
	pos := len(b)
	neg := i < 0
	if neg {
		i = -i
	}

	for i > 0 {
		pos--
		b[pos] = byte('0' + i%10)
		i /= 10
	}

	if neg {
		pos--
		b[pos] = '-'
	}

	return string(b[pos:])
}
