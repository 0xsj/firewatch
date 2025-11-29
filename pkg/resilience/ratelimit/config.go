package ratelimit

import "time"

// StrategyType identifies the rate limiting algorithm.
type StrategyType string

const (
	// StrategyTokenBucket uses the token bucket algorithm.
	// Allows bursts up to bucket capacity, refills at a steady rate.
	StrategyTokenBucket StrategyType = "token_bucket"

	// StrategySlidingWindow uses the sliding window log algorithm.
	// Provides accurate rate limiting with smooth distribution.
	StrategySlidingWindow StrategyType = "sliding_window"

	// StrategyFixedWindow uses the fixed window counter algorithm.
	// Simple and memory-efficient, but allows bursts at window boundaries.
	StrategyFixedWindow StrategyType = "fixed_window"
)

// Config holds rate limiter configuration.
type Config struct {
	// Strategy specifies which algorithm to use.
	Strategy StrategyType

	// Limit is the maximum number of requests allowed within the window.
	Limit int

	// Window is the time window for the limit.
	Window time.Duration

	// Burst is the maximum burst size (only applicable to token bucket).
	// If zero, defaults to Limit.
	Burst int

	// KeyPrefix is prepended to all keys in the store.
	// Useful for namespacing different limiters.
	KeyPrefix string

	// CleanupInterval is how often expired entries are cleaned up.
	// Only applicable to in-memory stores. If zero, defaults to Window.
	CleanupInterval time.Duration
}

// DefaultConfig returns a sensible default configuration.
func DefaultConfig() Config {
	return Config{
		Strategy:        StrategyTokenBucket,
		Limit:           100,
		Window:          time.Minute,
		Burst:           0,
		KeyPrefix:       "rl:",
		CleanupInterval: 0,
	}
}

// Validate checks if the configuration is valid.
func (c Config) Validate() error {
	if c.Limit <= 0 {
		return ErrInvalidConfig
	}
	if c.Window <= 0 {
		return ErrInvalidConfig
	}
	if c.Burst < 0 {
		return ErrInvalidConfig
	}

	switch c.Strategy {
	case StrategyTokenBucket, StrategySlidingWindow, StrategyFixedWindow:
		// Valid
	default:
		return ErrInvalidConfig
	}

	return nil
}

// BurstSize returns the effective burst size.
// Returns Burst if set, otherwise returns Limit.
func (c Config) BurstSize() int {
	if c.Burst > 0 {
		return c.Burst
	}
	return c.Limit
}

// RefillRate returns tokens per second for token bucket.
func (c Config) RefillRate() float64 {
	return float64(c.Limit) / c.Window.Seconds()
}
