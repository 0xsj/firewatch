package session

import (
	"context"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Repository is the port for session persistence.
type Repository interface {
	// Save persists a session (insert or update).
	Save(ctx context.Context, session *Session) error

	// FindByID retrieves a session by ID.
	FindByID(ctx context.Context, id types.ID) (*Session, error)

	// FindByRefreshToken retrieves a session by refresh token.
	FindByRefreshToken(ctx context.Context, token string) (*Session, error)

	// FindActiveByUserID retrieves all active sessions for a user.
	FindActiveByUserID(ctx context.Context, userID types.ID) ([]*Session, error)

	// FindActiveByUserIDAndTenant retrieves active sessions for a user in a tenant.
	FindActiveByUserIDAndTenant(ctx context.Context, userID types.ID, tenantID string) ([]*Session, error)

	// CountActiveByUserID returns the count of active sessions for a user.
	CountActiveByUserID(ctx context.Context, userID types.ID) (int, error)

	// DeleteExpired removes expired sessions older than the given duration.
	DeleteExpired(ctx context.Context) (int, error)

	// RevokeAllByUserID revokes all sessions for a user.
	RevokeAllByUserID(ctx context.Context, userID types.ID, reason string) error

	// RevokeAllByUserIDExcept revokes all sessions for a user except the given session.
	RevokeAllByUserIDExcept(ctx context.Context, userID types.ID, exceptSessionID types.ID, reason string) error
}

// CacheRepository is an optional interface for session caching.
// Implementations can use Redis or other caching systems.
type CacheRepository interface {
	// Get retrieves a session from cache by ID.
	Get(ctx context.Context, id types.ID) (*Session, error)

	// GetByRefreshToken retrieves a session from cache by refresh token.
	GetByRefreshToken(ctx context.Context, token string) (*Session, error)

	// Set stores a session in cache with TTL.
	Set(ctx context.Context, session *Session) error

	// Delete removes a session from cache.
	Delete(ctx context.Context, id types.ID) error

	// DeleteByRefreshToken removes a session from cache by refresh token.
	DeleteByRefreshToken(ctx context.Context, token string) error
}
