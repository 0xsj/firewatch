package ratelimit

import (
	"context"
	"time"
)

// Store defines the interface for rate limit state persistence.
// This is a PORT in hexagonal architecture.
//
// Implementations:
//   - memory.Store: in-memory (single instance, development/testing)
//   - redis.Store: Redis-backed (distributed, production)
type Store interface {
	// Get retrieves the current state for a key.
	// Returns ErrKeyNotFound if the key does not exist.
	Get(ctx context.Context, key string) (*State, error)

	// Set stores the state for a key with expiration.
	// If TTL is zero, the implementation should use a default.
	Set(ctx context.Context, key string, state *State, ttl time.Duration) error

	// Increment atomically increments the counter for a key.
	// Creates the key with count=n if it doesn't exist.
	// Returns the new count and the time until expiration.
	Increment(ctx context.Context, key string, n int, window time.Duration) (count int, remaining time.Duration, err error)

	// Delete removes a key from the store.
	Delete(ctx context.Context, key string) error

	// Close releases any resources held by the store.
	Close() error
}

// State represents the stored rate limit state for a key.
type State struct {
	// Tokens is the current token count (token bucket).
	Tokens float64

	// Count is the current request count (fixed/sliding window).
	Count int

	// LastAccess is when the key was last accessed.
	LastAccess time.Time

	// WindowStart is when the current window started (fixed window).
	WindowStart time.Time

	// Timestamps holds request timestamps (sliding window log).
	Timestamps []time.Time
}

// NewState creates a new empty state.
func NewState() *State {
	return &State{
		LastAccess: time.Now(),
	}
}

// NewTokenBucketState creates initial state for token bucket.
func NewTokenBucketState(tokens float64) *State {
	return &State{
		Tokens:     tokens,
		LastAccess: time.Now(),
	}
}

// NewFixedWindowState creates initial state for fixed window.
func NewFixedWindowState() *State {
	now := time.Now()
	return &State{
		Count:       0,
		WindowStart: now,
		LastAccess:  now,
	}
}

// NewSlidingWindowState creates initial state for sliding window.
func NewSlidingWindowState() *State {
	return &State{
		Timestamps: make([]time.Time, 0),
		LastAccess: time.Now(),
	}
}
