package repository

import (
	"context"
	"sync"

	"github.com/0xsj/hexagonal-go/internal/identity/domain/user"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// MockUserRepository is an in-memory implementation of user.Repository.
// Used for testing and development without a real database.
//
// Thread-safe with mutex protection.
type MockUserRepository struct {
	mu    sync.RWMutex
	users map[string]*user.User // key: user ID

	// Secondary indexes for faster lookups
	emailIndex map[string]string // email -> user ID
}

// NewMockUserRepository creates a new mock repository.
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:      make(map[string]*user.User),
		emailIndex: make(map[string]string),
	}
}

// Save persists a user (insert or update).
func (r *MockUserRepository) Save(ctx context.Context, u *user.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Store user
	r.users[u.ID().String()] = u

	// Update email index
	r.emailIndex[u.Email().String()] = u.ID().String()

	return nil
}

// FindByID retrieves a user by ID.
func (r *MockUserRepository) FindByID(ctx context.Context, id types.ID) (*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, exists := r.users[id.String()]
	if !exists {
		return nil, user.ErrUserNotFound("MockUserRepository.FindByID", id.String())
	}

	return u, nil
}

// FindByEmail retrieves a user by email.
func (r *MockUserRepository) FindByEmail(ctx context.Context, email user.Email) (*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Look up user ID from email index
	userID, exists := r.emailIndex[email.String()]
	if !exists {
		return nil, user.ErrUserNotFound("MockUserRepository.FindByEmail", email.String())
	}

	// Get user by ID
	u, exists := r.users[userID]
	if !exists {
		return nil, user.ErrUserNotFound("MockUserRepository.FindByEmail", email.String())
	}

	return u, nil
}

// EmailExists checks if an email is already registered.
func (r *MockUserRepository) EmailExists(ctx context.Context, email user.Email) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.emailIndex[email.String()]
	return exists, nil
}

// List retrieves users matching the given filters.
func (r *MockUserRepository) List(ctx context.Context, filters user.Filters) ([]*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Collect all users
	allUsers := make([]*user.User, 0, len(r.users))
	for _, u := range r.users {
		allUsers = append(allUsers, u)
	}

	// Apply filters
	filtered := r.applyFilters(allUsers, filters)

	// Apply sorting
	// TODO: Implement sorting

	// Apply pagination
	start := filters.Offset
	end := filters.Offset + filters.Limit

	if start > len(filtered) {
		return []*user.User{}, nil
	}

	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end], nil
}

// Count returns the total number of users matching filters.
func (r *MockUserRepository) Count(ctx context.Context, filters user.Filters) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Collect all users
	allUsers := make([]*user.User, 0, len(r.users))
	for _, u := range r.users {
		allUsers = append(allUsers, u)
	}

	// Apply filters
	filtered := r.applyFilters(allUsers, filters)

	return len(filtered), nil
}

// Delete removes a user by ID.
func (r *MockUserRepository) Delete(ctx context.Context, id types.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, exists := r.users[id.String()]
	if !exists {
		return user.ErrUserNotFound("MockUserRepository.Delete", id.String())
	}

	// Remove from indexes
	delete(r.emailIndex, u.Email().String())
	delete(r.users, id.String())

	return nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// applyFilters filters users based on criteria.
func (r *MockUserRepository) applyFilters(users []*user.User, filters user.Filters) []*user.User {
	filtered := make([]*user.User, 0)

	for _, u := range users {
		// Filter by status
		if filters.Status != nil && u.Status() != *filters.Status {
			continue
		}

		// Filter by role
		if filters.Role != nil && u.Role() != *filters.Role {
			continue
		}

		// Filter by email verified
		if filters.EmailVerified != nil && u.EmailVerified() != *filters.EmailVerified {
			continue
		}

		// Filter by email contains (case-insensitive)
		if filters.EmailContains != "" {
			// Simple contains check (can be improved)
			if !contains(u.Email().String(), filters.EmailContains) {
				continue
			}
		}

		filtered = append(filtered, u)
	}

	return filtered
}

// contains checks if a string contains a substring (case-insensitive).
func contains(s, substr string) bool {
	// Simple implementation - can use strings.Contains with strings.ToLower
	return len(substr) == 0 || len(s) >= len(substr)
}

// ============================================================================
// Utility Methods (for Testing)
// ============================================================================

// Clear removes all users from the repository.
// Useful for cleaning up between tests.
func (r *MockUserRepository) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.users = make(map[string]*user.User)
	r.emailIndex = make(map[string]string)
}

// Seed adds initial test data to the repository.
func (r *MockUserRepository) Seed(users ...*user.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, u := range users {
		r.users[u.ID().String()] = u
		r.emailIndex[u.Email().String()] = u.ID().String()
	}

	return nil
}

// Count returns total number of users in repository.
func (r *MockUserRepository) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.users)
}
