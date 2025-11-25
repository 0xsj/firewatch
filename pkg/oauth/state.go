package oauth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"
)

// StateManager manages OAuth state tokens for CSRF protection
type StateManager struct {
	// In-memory store for simplicity
	// In production, use Redis or encrypted JWT
	states map[string]time.Time
}

// NewStateManager creates a new state manager
func NewStateManager() *StateManager {
	return &StateManager{
		states: make(map[string]time.Time),
	}
}

// Generate generates a new state token
func (sm *StateManager) Generate() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}

	state := base64.URLEncoding.EncodeToString(b)
	sm.states[state] = time.Now().Add(10 * time.Minute) // 10 minute expiry

	return state, nil
}

// Validate validates a state token and removes it (one-time use)
func (sm *StateManager) Validate(state string) error {
	expiry, exists := sm.states[state]
	if !exists {
		return ErrInvalidState
	}

	delete(sm.states, state) // Remove after validation (one-time use)

	if time.Now().After(expiry) {
		return ErrExpiredState
	}

	return nil
}

// Cleanup removes expired state tokens (should be called periodically)
func (sm *StateManager) Cleanup() {
	now := time.Now()
	for state, expiry := range sm.states {
		if now.After(expiry) {
			delete(sm.states, state)
		}
	}
}
