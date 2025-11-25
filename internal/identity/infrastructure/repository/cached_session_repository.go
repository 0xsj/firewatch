package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/0xsj/hexagonal-go/internal/identity/domain/session"
	"github.com/0xsj/hexagonal-go/pkg/cache"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// CachedSessionRepository decorates a session repository with Redis caching.
type CachedSessionRepository struct {
	repo  session.Repository
	cache cache.Cache
}

// NewCachedSessionRepository creates a new cached session repository.
func NewCachedSessionRepository(repo session.Repository, cache cache.Cache) session.Repository {
	return &CachedSessionRepository{
		repo:  repo,
		cache: cache,
	}
}

// Cache key patterns
const (
	sessionByIDKey           = "session:id:%s"
	sessionByRefreshTokenKey = "session:token:%s"
	sessionCacheTTL          = 15 * time.Minute
)

// Save persists a session and updates cache.
func (r *CachedSessionRepository) Save(ctx context.Context, s *session.Session) error {
	// Save to database first (source of truth)
	if err := r.repo.Save(ctx, s); err != nil {
		return err
	}

	// Update cache
	r.cacheSession(ctx, s)

	return nil
}

// FindByID retrieves a session by ID (cache-first).
func (r *CachedSessionRepository) FindByID(ctx context.Context, id types.ID) (*session.Session, error) {
	cacheKey := fmt.Sprintf(sessionByIDKey, id.String())

	// Try cache first
	if data, err := r.cache.Get(ctx, cacheKey); err == nil {
		var snap session.Snapshot
		if err := json.Unmarshal(data, &snap); err == nil {
			if s, err := session.FromSnapshot(snap); err == nil {
				return s, nil
			}
		}
	}

	// Cache miss - load from database
	s, err := r.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache the result
	r.cacheSession(ctx, s)

	return s, nil
}

// FindByRefreshToken retrieves a session by refresh token (cache-first).
func (r *CachedSessionRepository) FindByRefreshToken(ctx context.Context, token string) (*session.Session, error) {
	cacheKey := fmt.Sprintf(sessionByRefreshTokenKey, token)

	// Try cache first
	if data, err := r.cache.Get(ctx, cacheKey); err == nil {
		var snap session.Snapshot
		if err := json.Unmarshal(data, &snap); err == nil {
			if s, err := session.FromSnapshot(snap); err == nil {
				return s, nil
			}
		}
	}

	// Cache miss - load from database
	s, err := r.repo.FindByRefreshToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// Cache the result
	r.cacheSession(ctx, s)

	return s, nil
}

// FindActiveByUserID retrieves all active sessions for a user (database only).
func (r *CachedSessionRepository) FindActiveByUserID(ctx context.Context, userID types.ID) ([]*session.Session, error) {
	// No caching for list operations - too complex to invalidate
	return r.repo.FindActiveByUserID(ctx, userID)
}

// FindActiveByUserIDAndTenant retrieves active sessions for a user in a tenant (database only).
func (r *CachedSessionRepository) FindActiveByUserIDAndTenant(ctx context.Context, userID types.ID, tenantID string) ([]*session.Session, error) {
	// No caching for list operations
	return r.repo.FindActiveByUserIDAndTenant(ctx, userID, tenantID)
}

// CountActiveByUserID returns the count of active sessions for a user (database only).
func (r *CachedSessionRepository) CountActiveByUserID(ctx context.Context, userID types.ID) (int, error) {
	// No caching for counts - can be stale
	return r.repo.CountActiveByUserID(ctx, userID)
}

// DeleteExpired removes expired sessions and invalidates cache.
func (r *CachedSessionRepository) DeleteExpired(ctx context.Context) (int, error) {
	// Database operation only - cache entries will naturally expire via TTL
	return r.repo.DeleteExpired(ctx)
}

// RevokeAllByUserID revokes all sessions for a user and invalidates cache.
func (r *CachedSessionRepository) RevokeAllByUserID(ctx context.Context, userID types.ID, reason string) error {
	// Load sessions to get their IDs for cache invalidation
	sessions, _ := r.repo.FindActiveByUserID(ctx, userID)

	// Revoke in database
	if err := r.repo.RevokeAllByUserID(ctx, userID, reason); err != nil {
		return err
	}

	// Invalidate cache entries
	for _, s := range sessions {
		r.invalidateSession(ctx, s)
	}

	return nil
}

// RevokeAllByUserIDExcept revokes all sessions for a user except the given session.
func (r *CachedSessionRepository) RevokeAllByUserIDExcept(ctx context.Context, userID types.ID, exceptSessionID types.ID, reason string) error {
	// Load sessions to get their IDs for cache invalidation
	sessions, _ := r.repo.FindActiveByUserID(ctx, userID)

	// Revoke in database
	if err := r.repo.RevokeAllByUserIDExcept(ctx, userID, exceptSessionID, reason); err != nil {
		return err
	}

	// Invalidate cache entries (except the one we're keeping)
	for _, s := range sessions {
		if s.ID() != exceptSessionID {
			r.invalidateSession(ctx, s)
		}
	}

	return nil
}

// ============================================================================
// Cache Helpers
// ============================================================================

// cacheSession stores a session in cache.
func (r *CachedSessionRepository) cacheSession(ctx context.Context, s *session.Session) {
	snap := s.ToSnapshot()
	data, err := json.Marshal(snap)
	if err != nil {
		return // Silent failure - cache is best-effort
	}

	// Cache by ID
	idKey := fmt.Sprintf(sessionByIDKey, s.ID().String())
	r.cache.Set(ctx, idKey, data, sessionCacheTTL)

	// Cache by refresh token
	tokenKey := fmt.Sprintf(sessionByRefreshTokenKey, s.RefreshToken())
	r.cache.Set(ctx, tokenKey, data, sessionCacheTTL)
}

// invalidateSession removes a session from cache.
func (r *CachedSessionRepository) invalidateSession(ctx context.Context, s *session.Session) {
	idKey := fmt.Sprintf(sessionByIDKey, s.ID().String())
	r.cache.Delete(ctx, idKey)

	tokenKey := fmt.Sprintf(sessionByRefreshTokenKey, s.RefreshToken())
	r.cache.Delete(ctx, tokenKey)
}
