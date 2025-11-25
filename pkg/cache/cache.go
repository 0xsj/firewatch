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

	// Close closes the cache connection.
	Close() error
}

// Errors
var (
	ErrCacheMiss = &Error{Code: "cache_miss", Message: "key not found in cache"}
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
