package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/0xsj/hexagonal-go/pkg/cache"
	"github.com/0xsj/hexagonal-go/pkg/resilience/ratelimit"
)

// Store is a Redis-backed implementation of ratelimit.Store.
// Suitable for distributed systems where rate limit state must be shared.
type Store struct {
	cache     cache.Cache
	keyPrefix string
}

// NewStore creates a new Redis-backed store.
func NewStore(cache cache.Cache, keyPrefix string) *Store {
	return &Store{
		cache:     cache,
		keyPrefix: keyPrefix,
	}
}

// Get retrieves the current state for a key.
func (s *Store) Get(ctx context.Context, key string) (*ratelimit.State, error) {
	fullKey := s.prefixedKey(key)

	data, err := s.cache.Get(ctx, fullKey)
	if err != nil {
		if cache.IsCacheMiss(err) {
			return nil, ratelimit.ErrKeyNotFound
		}
		return nil, fmt.Errorf("redis get: %w", err)
	}

	var state ratelimit.State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("unmarshal state: %w", err)
	}

	return &state, nil
}

// Set stores the state for a key with expiration.
func (s *Store) Set(ctx context.Context, key string, state *ratelimit.State, ttl time.Duration) error {
	fullKey := s.prefixedKey(key)

	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	if err := s.cache.Set(ctx, fullKey, data, ttl); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}

	return nil
}

// Increment atomically increments the counter for a key.
// Uses Redis INCR with expiration for atomic counter operations.
func (s *Store) Increment(ctx context.Context, key string, n int, window time.Duration) (int, time.Duration, error) {
	fullKey := s.prefixedKey(key)

	// Atomically increment
	count, err := s.cache.Increment(ctx, fullKey, int64(n))
	if err != nil {
		return 0, 0, fmt.Errorf("redis increment: %w", err)
	}

	// If this is a new key (count equals n), set expiration
	if count == int64(n) {
		if err := s.cache.Expire(ctx, fullKey, window); err != nil {
			// Key might have been deleted between increment and expire
			if !cache.IsKeyNotFound(err) {
				return 0, 0, fmt.Errorf("redis expire: %w", err)
			}
		}
		return int(count), window, nil
	}

	// Get remaining TTL for existing key
	ttl, err := s.cache.TTL(ctx, fullKey)
	if err != nil {
		if cache.IsKeyNotFound(err) {
			return int(count), window, nil
		}
		// If we can't get TTL, estimate based on window
		return int(count), window, nil
	}

	return int(count), ttl, nil
}

// Delete removes a key from the store.
func (s *Store) Delete(ctx context.Context, key string) error {
	fullKey := s.prefixedKey(key)

	if err := s.cache.Delete(ctx, fullKey); err != nil {
		return fmt.Errorf("redis delete: %w", err)
	}

	return nil
}

// Close releases any resources held by the store.
// The underlying cache connection is managed externally.
func (s *Store) Close() error {
	return nil
}

// prefixedKey returns the key with the configured prefix.
func (s *Store) prefixedKey(key string) string {
	if s.keyPrefix == "" {
		return key
	}
	return s.keyPrefix + key
}
