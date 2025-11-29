package ratelimit

import (
	"context"
	"net/http"

	"github.com/0xsj/hexagonal-go/pkg/errors"
	"github.com/0xsj/hexagonal-go/pkg/http/response"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// Middleware provides HTTP middleware for rate limiting.
type Middleware struct {
	limiter Limiter
	keyFunc KeyFunc
	logger  logger.Logger
}

// NewMiddleware creates a new rate limiting middleware.
func NewMiddleware(limiter Limiter, keyFunc KeyFunc, logger logger.Logger) *Middleware {
	return &Middleware{
		limiter: limiter,
		keyFunc: keyFunc,
		logger:  logger,
	}
}

// Handler returns an HTTP middleware handler.
func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract rate limit key
		key, err := m.keyFunc(ctx)
		if err != nil {
			m.logger.Error("failed to extract rate limit key",
				logger.String("path", r.URL.Path),
				logger.Err(err),
			)
			// Allow request on key extraction failure to avoid blocking legitimate traffic
			next.ServeHTTP(w, r)
			return
		}

		// Check rate limit
		result, err := m.limiter.Allow(ctx, key)
		if err != nil {
			m.logger.Error("rate limit check failed",
				logger.String("key", key),
				logger.String("path", r.URL.Path),
				logger.Err(err),
			)
			// Allow request on limiter failure to avoid blocking legitimate traffic
			next.ServeHTTP(w, r)
			return
		}

		// Set rate limit headers
		for name, value := range result.Headers() {
			w.Header().Set(name, value)
		}

		// Deny if rate limited
		if result.Denied() {
			m.logger.Warn("rate limit exceeded",
				logger.String("key", key),
				logger.String("path", r.URL.Path),
				logger.Int("limit", result.Limit),
			)
			err := errors.RateLimit("ratelimit.Middleware", "rate limit exceeded")
			response.Error(w, err)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ============================================================================
// Common Key Functions
// ============================================================================

// IPKeyFunc extracts the client IP address as the rate limit key.
func IPKeyFunc(r *http.Request) KeyFunc {
	return func(ctx context.Context) (string, error) {
		return extractIP(r), nil
	}
}

// IPKeyFuncFromContext returns a KeyFunc that extracts IP from the request in context.
// Requires the request to be stored in context.
func IPKeyFuncFromContext() KeyFunc {
	return func(ctx context.Context) (string, error) {
		r, ok := ctx.Value(requestContextKey).(*http.Request)
		if !ok {
			return "unknown", nil
		}
		return extractIP(r), nil
	}
}

// UserKeyFunc returns a KeyFunc that extracts a user ID from context.
// Uses the provided function to get the user ID.
func UserKeyFunc(getUserID func(ctx context.Context) string) KeyFunc {
	return func(ctx context.Context) (string, error) {
		userID := getUserID(ctx)
		if userID == "" {
			return "anonymous", nil
		}
		return "user:" + userID, nil
	}
}

// TenantKeyFunc returns a KeyFunc that extracts a tenant ID from context.
// Uses the provided function to get the tenant ID.
func TenantKeyFunc(getTenantID func(ctx context.Context) string) KeyFunc {
	return func(ctx context.Context) (string, error) {
		tenantID := getTenantID(ctx)
		if tenantID == "" {
			return "default", nil
		}
		return "tenant:" + tenantID, nil
	}
}

// CompositeKeyFunc combines multiple key functions into a single key.
// Keys are joined with a colon separator.
func CompositeKeyFunc(funcs ...KeyFunc) KeyFunc {
	return func(ctx context.Context) (string, error) {
		var key string
		for i, fn := range funcs {
			part, err := fn(ctx)
			if err != nil {
				return "", err
			}
			if i == 0 {
				key = part
			} else {
				key = key + ":" + part
			}
		}
		return key, nil
	}
}

// EndpointKeyFunc returns a KeyFunc that includes the endpoint in the key.
// Useful for per-endpoint rate limiting.
func EndpointKeyFunc(r *http.Request, base KeyFunc) KeyFunc {
	return func(ctx context.Context) (string, error) {
		baseKey, err := base(ctx)
		if err != nil {
			return "", err
		}
		return baseKey + ":" + r.Method + ":" + r.URL.Path, nil
	}
}

// ============================================================================
// Helpers
// ============================================================================

// contextKey is a custom type for context keys.
type contextKey string

const requestContextKey contextKey = "http_request"

// WithRequest stores the HTTP request in context.
// Required for IPKeyFuncFromContext.
func WithRequest(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, requestContextKey, r)
}

// extractIP extracts the client IP from an HTTP request.
// Checks common proxy headers before falling back to RemoteAddr.
func extractIP(r *http.Request) string {
	// Check X-Forwarded-For header (may contain multiple IPs)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP (original client)
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	addr := r.RemoteAddr
	// Strip port if present
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i]
		}
		if addr[i] == ']' {
			// IPv6 address with brackets
			break
		}
	}

	return addr
}
