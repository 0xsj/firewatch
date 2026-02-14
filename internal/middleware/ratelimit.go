package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/0xsj/firewatch/internal/storage"
	"github.com/0xsj/firewatch/internal/storage/models"
	"github.com/0xsj/firewatch/pkg/crypto"
	"github.com/0xsj/firewatch/pkg/httputil"
	"github.com/0xsj/firewatch/pkg/timeutil"
)

// RateLimiterConfig holds rate limiting parameters.
type RateLimiterConfig struct {
	RequestsPerSecond float64       // Rate limit (e.g., 10 = 10 req/sec)
	Burst             int           // Burst allowance (e.g., 20 = allow bursts up to 20)
	CleanupInterval   time.Duration // How often to clean up stale limiters
}

// RateLimiter manages per-IP rate limiters using token bucket algorithm.
type RateLimiter struct {
	cfg      RateLimiterConfig
	store    storage.Store
	logger   *slog.Logger
	mu       sync.RWMutex
	limiters map[string]*limiterEntry
	stopCh   chan struct{}
}

// limiterEntry holds a rate limiter and its last access time.
type limiterEntry struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

// NewRateLimiter creates a new rate limiter with background cleanup.
func NewRateLimiter(cfg RateLimiterConfig, store storage.Store, logger *slog.Logger) *RateLimiter {
	rl := &RateLimiter{
		cfg:      cfg,
		store:    store,
		logger:   logger,
		limiters: make(map[string]*limiterEntry),
		stopCh:   make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Stop terminates the cleanup goroutine.
func (rl *RateLimiter) Stop() {
	close(rl.stopCh)
}

// getLimiter retrieves or creates a rate limiter for the given IP.
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.RLock()
	entry, exists := rl.limiters[ip]
	rl.mu.RUnlock()

	if exists {
		// Update last access time
		rl.mu.Lock()
		entry.lastAccess = time.Now()
		rl.mu.Unlock()
		return entry.limiter
	}

	// Create new limiter
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double-check after acquiring write lock
	if entry, exists := rl.limiters[ip]; exists {
		entry.lastAccess = time.Now()
		return entry.limiter
	}

	limiter := rate.NewLimiter(rate.Limit(rl.cfg.RequestsPerSecond), rl.cfg.Burst)
	rl.limiters[ip] = &limiterEntry{
		limiter:    limiter,
		lastAccess: time.Now(),
	}

	return limiter
}

// cleanup periodically removes stale limiters to prevent memory leaks.
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cfg.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			staleThreshold := rl.cfg.CleanupInterval * 2

			for ip, entry := range rl.limiters {
				if now.Sub(entry.lastAccess) > staleThreshold {
					delete(rl.limiters, ip)
				}
			}

			rl.mu.Unlock()

		case <-rl.stopCh:
			return
		}
	}
}

// recordRateLimitEvent creates a detection event when rate limit is exceeded.
func (rl *RateLimiter) recordRateLimitEvent(ctx context.Context, r *http.Request) {
	event := &models.Event{
		ID:         crypto.UUID4(),
		Timestamp:  timeutil.FormatRFC3339(timeutil.NowUTC()),
		RequestID:  RequestID(ctx),
		SourceIP:   httputil.ClientIP(r),
		Module:     "rate_limit",
		Method:     r.Method,
		Path:       r.URL.Path,
		Query:      r.URL.RawQuery,
		Headers:    httputil.HeaderMap(r.Header),
		UserAgent:  r.UserAgent(),
		Severity:   "medium",
		Signatures: []string{"rate-limit-exceeded"},
	}

	if err := rl.store.SaveEvent(ctx, event); err != nil {
		rl.logger.Error("failed to save rate limit event",
			"error", err,
			"event_id", event.ID,
			"source_ip", event.SourceIP,
		)
	}
}

// RateLimit returns middleware that enforces per-IP rate limiting.
func RateLimit(rl *RateLimiter) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if rl == nil {
				next.ServeHTTP(w, r)
				return
			}

			ip := httputil.ClientIP(r)
			limiter := rl.getLimiter(ip)

			if !limiter.Allow() {
				// Rate limit exceeded
				rl.logger.Warn("rate limit exceeded",
					"ip", ip,
					"path", r.URL.Path,
					"request_id", RequestID(r.Context()),
				)

				// Record event
				rl.recordRateLimitEvent(r.Context(), r)

				// Calculate retry-after based on rate (in seconds)
				retryAfter := int(1.0 / rl.cfg.RequestsPerSecond)
				if retryAfter < 1 {
					retryAfter = 1
				}

				w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.cfg.Burst))
				w.Header().Set("X-RateLimit-Remaining", "0")
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}

			// Request allowed
			next.ServeHTTP(w, r)
		})
	}
}
