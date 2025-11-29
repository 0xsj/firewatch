package ratelimit

import "context"

// Strategy defines the interface for rate limiting algorithms.
// This is a PORT in hexagonal architecture.
//
// Implementations:
//   - TokenBucket: allows bursts, smooth refill
//   - SlidingWindow: accurate, no boundary issues
//   - FixedWindow: simple, memory efficient
type Strategy interface {
	// Allow checks if a request identified by key is allowed.
	// Returns a Result indicating whether the request can proceed.
	Allow(ctx context.Context, key string) (Result, error)

	// AllowN checks if n requests identified by key are allowed.
	// Useful for batch operations or weighted requests.
	AllowN(ctx context.Context, key string, n int) (Result, error)

	// Reset clears the rate limit state for a key.
	// Used for administrative purposes (e.g., unblocking a user).
	Reset(ctx context.Context, key string) error

	// Peek returns the current state without consuming a request.
	// Useful for checking remaining quota without affecting it.
	Peek(ctx context.Context, key string) (Result, error)
}
