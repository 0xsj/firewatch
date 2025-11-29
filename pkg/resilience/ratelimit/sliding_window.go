package ratelimit

import (
	"context"
	"sync"
	"time"
)

// SlidingWindow implements the sliding window log rate limiting algorithm.
// Provides accurate rate limiting with smooth distribution across the window.
// More memory-intensive than fixed window but avoids boundary burst issues.
type SlidingWindow struct {
	config Config
	store  Store
	mu     sync.Mutex
}

// NewSlidingWindow creates a new sliding window rate limiter.
func NewSlidingWindow(config Config, store Store) (*SlidingWindow, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &SlidingWindow{
		config: config,
		store:  store,
	}, nil
}

// Allow checks if a single request is allowed.
func (sw *SlidingWindow) Allow(ctx context.Context, key string) (Result, error) {
	return sw.AllowN(ctx, key, 1)
}

// AllowN checks if n requests are allowed.
func (sw *SlidingWindow) AllowN(ctx context.Context, key string, n int) (Result, error) {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	prefixedKey := sw.config.KeyPrefix + key
	now := time.Now()
	windowStart := now.Add(-sw.config.Window)

	// Get current state
	state, err := sw.store.Get(ctx, prefixedKey)
	if err != nil && err != ErrKeyNotFound {
		return Result{}, err
	}

	// Initialize state if not found
	if state == nil {
		state = NewSlidingWindowState()
	}

	// Remove expired timestamps
	state.Timestamps = sw.filterExpired(state.Timestamps, windowStart)

	// Build result
	currentCount := len(state.Timestamps)
	result := Result{
		Limit:   sw.config.Limit,
		ResetAt: now.Add(sw.config.Window),
	}

	// Check if we have room for n requests
	if currentCount+n <= sw.config.Limit {
		// Add new timestamps
		for i := 0; i < n; i++ {
			state.Timestamps = append(state.Timestamps, now)
		}
		result.Allowed = true
		result.Remaining = sw.config.Limit - currentCount - n
	} else {
		// Over limit
		result.Allowed = false
		result.Remaining = sw.config.Limit - currentCount
		if result.Remaining < 0 {
			result.Remaining = 0
		}

		// Calculate retry after based on oldest timestamp
		if len(state.Timestamps) > 0 {
			oldest := state.Timestamps[0]
			result.RetryAfter = oldest.Add(sw.config.Window).Sub(now)
			if result.RetryAfter < 0 {
				result.RetryAfter = 0
			}
		}
	}

	// Update state
	state.LastAccess = now

	// Persist with TTL slightly longer than window
	ttl := sw.config.Window + time.Minute
	if err := sw.store.Set(ctx, prefixedKey, state, ttl); err != nil {
		return Result{}, err
	}

	return result, nil
}

// Reset clears the rate limit state for a key.
func (sw *SlidingWindow) Reset(ctx context.Context, key string) error {
	prefixedKey := sw.config.KeyPrefix + key
	return sw.store.Delete(ctx, prefixedKey)
}

// Peek returns the current state without consuming a request.
func (sw *SlidingWindow) Peek(ctx context.Context, key string) (Result, error) {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	prefixedKey := sw.config.KeyPrefix + key
	now := time.Now()
	windowStart := now.Add(-sw.config.Window)

	// Get current state
	state, err := sw.store.Get(ctx, prefixedKey)
	if err != nil && err != ErrKeyNotFound {
		return Result{}, err
	}

	// No state means full capacity available
	if state == nil {
		return Result{
			Allowed:   true,
			Limit:     sw.config.Limit,
			Remaining: sw.config.Limit,
			ResetAt:   now.Add(sw.config.Window),
		}, nil
	}

	// Count non-expired timestamps
	validTimestamps := sw.filterExpired(state.Timestamps, windowStart)
	currentCount := len(validTimestamps)
	remaining := sw.config.Limit - currentCount
	if remaining < 0 {
		remaining = 0
	}

	result := Result{
		Allowed:   remaining > 0,
		Limit:     sw.config.Limit,
		Remaining: remaining,
		ResetAt:   now.Add(sw.config.Window),
	}

	// Calculate retry after if at limit
	if !result.Allowed && len(validTimestamps) > 0 {
		oldest := validTimestamps[0]
		result.RetryAfter = oldest.Add(sw.config.Window).Sub(now)
		if result.RetryAfter < 0 {
			result.RetryAfter = 0
		}
	}

	return result, nil
}

// filterExpired removes timestamps older than windowStart.
func (sw *SlidingWindow) filterExpired(timestamps []time.Time, windowStart time.Time) []time.Time {
	if len(timestamps) == 0 {
		return timestamps
	}

	// Find first valid index using binary search approach
	// Timestamps are in chronological order
	firstValid := 0
	for i, ts := range timestamps {
		if !ts.Before(windowStart) {
			firstValid = i
			break
		}
		// If we've checked all and none are valid
		if i == len(timestamps)-1 {
			return []time.Time{}
		}
	}

	return timestamps[firstValid:]
}
