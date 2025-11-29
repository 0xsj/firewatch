package ratelimit

import (
	"context"
	"fmt"
)

// Limiter is the main interface for rate limiting operations.
// This is a PORT in hexagonal architecture, providing a unified API
// regardless of the underlying strategy.
type Limiter interface {
	// Allow checks if a request identified by key is allowed.
	Allow(ctx context.Context, key string) (Result, error)

	// AllowN checks if n requests identified by key are allowed.
	AllowN(ctx context.Context, key string, n int) (Result, error)

	// Reset clears the rate limit state for a key.
	Reset(ctx context.Context, key string) error

	// Peek returns the current state without consuming a request.
	Peek(ctx context.Context, key string) (Result, error)
}

// NewLimiter creates a rate limiter with the specified configuration and store.
// This is a factory function that returns the appropriate strategy implementation.
func NewLimiter(config Config, store Store) (Limiter, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	switch config.Strategy {
	case StrategyTokenBucket:
		return NewTokenBucket(config, store)
	case StrategySlidingWindow:
		return NewSlidingWindow(config, store)
	case StrategyFixedWindow:
		return NewFixedWindow(config, store)
	default:
		return nil, fmt.Errorf("%w: unknown strategy %q", ErrInvalidConfig, config.Strategy)
	}
}

// KeyFunc is a function that extracts a rate limit key from context.
// Used by middleware to determine how to identify requesters.
type KeyFunc func(ctx context.Context) (string, error)

// MultiLimiter applies multiple limiters in sequence.
// All limiters must allow the request for it to proceed.
// Useful for layered rate limiting (e.g., per-IP and per-user).
type MultiLimiter struct {
	limiters []Limiter
}

// NewMultiLimiter creates a limiter that applies multiple limiters in sequence.
func NewMultiLimiter(limiters ...Limiter) *MultiLimiter {
	return &MultiLimiter{
		limiters: limiters,
	}
}

// Allow checks if a request is allowed by all limiters.
func (m *MultiLimiter) Allow(ctx context.Context, key string) (Result, error) {
	return m.AllowN(ctx, key, 1)
}

// AllowN checks if n requests are allowed by all limiters.
// Returns the most restrictive result (lowest remaining, earliest retry).
func (m *MultiLimiter) AllowN(ctx context.Context, key string, n int) (Result, error) {
	if len(m.limiters) == 0 {
		return Result{Allowed: true}, nil
	}

	var combined Result
	combined.Allowed = true
	combined.Remaining = -1 // Sentinel for first iteration

	for _, limiter := range m.limiters {
		result, err := limiter.AllowN(ctx, key, n)
		if err != nil {
			return Result{}, err
		}

		// If any limiter denies, the request is denied
		if !result.Allowed {
			combined.Allowed = false
		}

		// Track most restrictive remaining count
		if combined.Remaining == -1 || result.Remaining < combined.Remaining {
			combined.Remaining = result.Remaining
			combined.Limit = result.Limit
		}

		// Track longest retry-after
		if result.RetryAfter > combined.RetryAfter {
			combined.RetryAfter = result.RetryAfter
		}

		// Track earliest reset time
		if combined.ResetAt.IsZero() || result.ResetAt.Before(combined.ResetAt) {
			combined.ResetAt = result.ResetAt
		}
	}

	return combined, nil
}

// Reset clears the rate limit state for a key across all limiters.
func (m *MultiLimiter) Reset(ctx context.Context, key string) error {
	for _, limiter := range m.limiters {
		if err := limiter.Reset(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

// Peek returns the most restrictive state across all limiters.
func (m *MultiLimiter) Peek(ctx context.Context, key string) (Result, error) {
	if len(m.limiters) == 0 {
		return Result{Allowed: true}, nil
	}

	var combined Result
	combined.Allowed = true
	combined.Remaining = -1

	for _, limiter := range m.limiters {
		result, err := limiter.Peek(ctx, key)
		if err != nil {
			return Result{}, err
		}

		if !result.Allowed {
			combined.Allowed = false
		}

		if combined.Remaining == -1 || result.Remaining < combined.Remaining {
			combined.Remaining = result.Remaining
			combined.Limit = result.Limit
		}

		if result.RetryAfter > combined.RetryAfter {
			combined.RetryAfter = result.RetryAfter
		}

		if combined.ResetAt.IsZero() || result.ResetAt.Before(combined.ResetAt) {
			combined.ResetAt = result.ResetAt
		}
	}

	return combined, nil
}
