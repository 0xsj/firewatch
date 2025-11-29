package ratelimit

import "errors"

// Sentinel errors for rate limiting operations.
var (
	// ErrRateLimitExceeded is returned when a request exceeds the rate limit.
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// ErrKeyNotFound is returned when a rate limit key does not exist in the store.
	ErrKeyNotFound = errors.New("key not found")

	// ErrStoreUnavailable is returned when the backing store is unavailable.
	ErrStoreUnavailable = errors.New("rate limit store unavailable")

	// ErrInvalidConfig is returned when rate limiter configuration is invalid.
	ErrInvalidConfig = errors.New("invalid rate limiter configuration")
)
