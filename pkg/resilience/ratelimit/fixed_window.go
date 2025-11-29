package ratelimit

import (
	"context"
	"time"
)

// FixedWindow implements the fixed window counter rate limiting algorithm.
// Simple and memory-efficient, but allows bursts at window boundaries.
// Best for scenarios where simplicity and low overhead are priorities.
type FixedWindow struct {
	config Config
	store  Store
}

// NewFixedWindow creates a new fixed window rate limiter.
func NewFixedWindow(config Config, store Store) (*FixedWindow, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &FixedWindow{
		config: config,
		store:  store,
	}, nil
}

// Allow checks if a single request is allowed.
func (fw *FixedWindow) Allow(ctx context.Context, key string) (Result, error) {
	return fw.AllowN(ctx, key, 1)
}

// AllowN checks if n requests are allowed.
func (fw *FixedWindow) AllowN(ctx context.Context, key string, n int) (Result, error) {
	prefixedKey := fw.config.KeyPrefix + key
	now := time.Now()

	// Calculate window key based on current time
	windowKey := fw.windowKey(prefixedKey, now)

	// Atomically increment counter
	count, remaining, err := fw.store.Increment(ctx, windowKey, n, fw.config.Window)
	if err != nil {
		return Result{}, err
	}

	// Calculate reset time (end of current window)
	resetAt := fw.windowEnd(now)

	// Build result
	result := Result{
		Limit:   fw.config.Limit,
		ResetAt: resetAt,
	}

	if count <= fw.config.Limit {
		result.Allowed = true
		result.Remaining = fw.config.Limit - count
	} else {
		result.Allowed = false
		result.Remaining = 0
		result.RetryAfter = remaining
	}

	return result, nil
}

// Reset clears the rate limit state for a key.
func (fw *FixedWindow) Reset(ctx context.Context, key string) error {
	prefixedKey := fw.config.KeyPrefix + key
	now := time.Now()
	windowKey := fw.windowKey(prefixedKey, now)
	return fw.store.Delete(ctx, windowKey)
}

// Peek returns the current state without consuming a request.
func (fw *FixedWindow) Peek(ctx context.Context, key string) (Result, error) {
	prefixedKey := fw.config.KeyPrefix + key
	now := time.Now()
	windowKey := fw.windowKey(prefixedKey, now)

	// Get current state
	state, err := fw.store.Get(ctx, windowKey)
	if err != nil && err != ErrKeyNotFound {
		return Result{}, err
	}

	// Calculate reset time
	resetAt := fw.windowEnd(now)

	// No state means no requests in current window
	if state == nil {
		return Result{
			Allowed:   true,
			Limit:     fw.config.Limit,
			Remaining: fw.config.Limit,
			ResetAt:   resetAt,
		}, nil
	}

	remaining := fw.config.Limit - state.Count
	if remaining < 0 {
		remaining = 0
	}

	result := Result{
		Allowed:   remaining > 0,
		Limit:     fw.config.Limit,
		Remaining: remaining,
		ResetAt:   resetAt,
	}

	if !result.Allowed {
		result.RetryAfter = resetAt.Sub(now)
	}

	return result, nil
}

// windowKey generates a unique key for the current time window.
func (fw *FixedWindow) windowKey(key string, t time.Time) string {
	// Truncate time to window boundary
	windowStart := t.Truncate(fw.config.Window)
	return key + ":" + windowStart.Format("20060102150405")
}

// windowEnd calculates when the current window ends.
func (fw *FixedWindow) windowEnd(t time.Time) time.Time {
	windowStart := t.Truncate(fw.config.Window)
	return windowStart.Add(fw.config.Window)
}
