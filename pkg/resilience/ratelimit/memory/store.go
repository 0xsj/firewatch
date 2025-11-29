package memory

import (
	"context"
	"sync"
	"time"

	"github.com/0xsj/hexagonal-go/pkg/resilience/ratelimit"
)

// Store is an in-memory implementation of ratelimit.Store.
// Suitable for single-instance deployments, development, and testing.
// Not suitable for distributed systems as state is not shared.
type Store struct {
	mu      sync.RWMutex
	entries map[string]*entry
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

// entry holds state with expiration metadata.
type entry struct {
	state     *ratelimit.State
	expiresAt time.Time
}

// NewStore creates a new in-memory store.
// If cleanupInterval > 0, starts a background goroutine to purge expired entries.
func NewStore(cleanupInterval time.Duration) *Store {
	s := &Store{
		entries: make(map[string]*entry),
		stopCh:  make(chan struct{}),
	}

	if cleanupInterval > 0 {
		s.wg.Add(1)
		go s.cleanupLoop(cleanupInterval)
	}

	return s
}

// Get retrieves the current state for a key.
func (s *Store) Get(ctx context.Context, key string) (*ratelimit.State, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, exists := s.entries[key]
	if !exists {
		return nil, ratelimit.ErrKeyNotFound
	}

	if time.Now().After(e.expiresAt) {
		return nil, ratelimit.ErrKeyNotFound
	}

	return e.state, nil
}

// Set stores the state for a key with expiration.
func (s *Store) Set(ctx context.Context, key string, state *ratelimit.State, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries[key] = &entry{
		state:     state,
		expiresAt: time.Now().Add(ttl),
	}

	return nil
}

// Increment atomically increments the counter for a key.
func (s *Store) Increment(ctx context.Context, key string, n int, window time.Duration) (int, time.Duration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	e, exists := s.entries[key]
	if !exists || now.After(e.expiresAt) {
		// Create new entry
		state := ratelimit.NewFixedWindowState()
		state.Count = n
		expiresAt := now.Add(window)

		s.entries[key] = &entry{
			state:     state,
			expiresAt: expiresAt,
		}

		return n, window, nil
	}

	// Increment existing
	e.state.Count += n
	e.state.LastAccess = now
	remaining := e.expiresAt.Sub(now)

	return e.state.Count, remaining, nil
}

// Delete removes a key from the store.
func (s *Store) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.entries, key)
	return nil
}

// Close stops the cleanup goroutine and releases resources.
func (s *Store) Close() error {
	close(s.stopCh)
	s.wg.Wait()
	return nil
}

// cleanupLoop periodically removes expired entries.
func (s *Store) cleanupLoop(interval time.Duration) {
	defer s.wg.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.cleanup()
		}
	}
}

// cleanup removes all expired entries.
func (s *Store) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for key, e := range s.entries {
		if now.After(e.expiresAt) {
			delete(s.entries, key)
		}
	}
}

// Size returns the number of entries in the store.
// Useful for testing and monitoring.
func (s *Store) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries)
}
