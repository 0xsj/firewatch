package cache

import (
	"context"
	"time"
)

// Cache is a generic key-value cache interface.
type Cache interface {
	// Get retrieves a value by key.
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value with TTL.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete removes a value by key.
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists.
	Exists(ctx context.Context, key string) (bool, error)

	// Increment atomically increments a key by delta and returns the new value.
	// Creates the key with value delta if it doesn't exist.
	Increment(ctx context.Context, key string, delta int64) (int64, error)

	// Expire sets a TTL on an existing key.
	// Returns ErrKeyNotFound if the key does not exist.
	Expire(ctx context.Context, key string, ttl time.Duration) error

	// TTL returns the remaining time-to-live for a key.
	// Returns ErrKeyNotFound if the key does not exist.
	// Returns zero duration if the key exists but has no expiration.
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Close closes the cache connection.
	Close() error
}

// Errors
var (
	ErrCacheMiss   = &Error{Code: "cache_miss", Message: "key not found in cache"}
	ErrKeyNotFound = &Error{Code: "key_not_found", Message: "key not found"}
)

// Error represents a cache error.
type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

// IsCacheMiss returns true if the error is a cache miss.
func IsCacheMiss(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == "cache_miss"
	}
	return false
}

// IsKeyNotFound returns true if the error is a key not found error.
func IsKeyNotFound(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == "key_not_found"
	}
	return false
}
