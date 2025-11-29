package ratelimit

import (
	"context"
	"sync"
	"time"
)

// TokenBucket implements the token bucket rate limiting algorithm.
// Allows bursts up to the bucket capacity, with tokens refilling at a steady rate.
type TokenBucket struct {
	config Config
	store  Store
	mu     sync.Mutex
}

// NewTokenBucket creates a new token bucket rate limiter.
func NewTokenBucket(config Config, store Store) (*TokenBucket, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &TokenBucket{
		config: config,
		store:  store,
	}, nil
}

// Allow checks if a single request is allowed.
func (tb *TokenBucket) Allow(ctx context.Context, key string) (Result, error) {
	return tb.AllowN(ctx, key, 1)
}

// AllowN checks if n requests are allowed.
func (tb *TokenBucket) AllowN(ctx context.Context, key string, n int) (Result, error) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	prefixedKey := tb.config.KeyPrefix + key
	now := time.Now()
	burst := float64(tb.config.BurstSize())
	refillRate := tb.config.RefillRate()

	// Get current state
	state, err := tb.store.Get(ctx, prefixedKey)
	if err != nil && err != ErrKeyNotFound {
		return Result{}, err
	}

	// Initialize state if not found
	if state == nil {
		state = NewTokenBucketState(burst)
	}

	// Calculate tokens to add based on elapsed time
	elapsed := now.Sub(state.LastAccess).Seconds()
	tokens := state.Tokens + (elapsed * refillRate)

	// Cap at bucket capacity
	if tokens > burst {
		tokens = burst
	}

	// Build result
	result := Result{
		Limit:   tb.config.Limit,
		ResetAt: now.Add(tb.config.Window),
	}

	// Check if we have enough tokens
	requested := float64(n)
	if tokens >= requested {
		// Consume tokens
		tokens -= requested
		result.Allowed = true
		result.Remaining = int(tokens)
	} else {
		// Not enough tokens
		result.Allowed = false
		result.Remaining = int(tokens)

		// Calculate retry after based on tokens needed
		tokensNeeded := requested - tokens
		secondsNeeded := tokensNeeded / refillRate
		result.RetryAfter = time.Duration(secondsNeeded * float64(time.Second))
	}

	// Update state
	state.Tokens = tokens
	state.LastAccess = now

	// Persist with TTL slightly longer than window to handle edge cases
	ttl := tb.config.Window + time.Minute
	if err := tb.store.Set(ctx, prefixedKey, state, ttl); err != nil {
		return Result{}, err
	}

	return result, nil
}

// Reset clears the rate limit state for a key.
func (tb *TokenBucket) Reset(ctx context.Context, key string) error {
	prefixedKey := tb.config.KeyPrefix + key
	return tb.store.Delete(ctx, prefixedKey)
}

// Peek returns the current state without consuming a request.
func (tb *TokenBucket) Peek(ctx context.Context, key string) (Result, error) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	prefixedKey := tb.config.KeyPrefix + key
	now := time.Now()
	burst := float64(tb.config.BurstSize())
	refillRate := tb.config.RefillRate()

	// Get current state
	state, err := tb.store.Get(ctx, prefixedKey)
	if err != nil && err != ErrKeyNotFound {
		return Result{}, err
	}

	// Initialize state if not found
	if state == nil {
		return Result{
			Allowed:   true,
			Limit:     tb.config.Limit,
			Remaining: tb.config.BurstSize(),
			ResetAt:   now.Add(tb.config.Window),
		}, nil
	}

	// Calculate current tokens
	elapsed := now.Sub(state.LastAccess).Seconds()
	tokens := state.Tokens + (elapsed * refillRate)
	if tokens > burst {
		tokens = burst
	}

	return Result{
		Allowed:   tokens >= 1,
		Limit:     tb.config.Limit,
		Remaining: int(tokens),
		ResetAt:   now.Add(tb.config.Window),
	}, nil
}
