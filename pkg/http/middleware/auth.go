package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/0xsj/hexagonal-go/pkg/http/response"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/security/jwt"
)

// AuthContextKey is the context key for authentication claims.
type AuthContextKey string

const (
	// ContextKeyClaims is the context key for JWT claims.
	ContextKeyClaims AuthContextKey = "auth_claims"
)

// TenancyConfig holds tenancy configuration for middleware.
type TenancyConfig struct {
	Enabled         bool
	DefaultTenantID string
}

// DefaultTenancyConfig returns default tenancy configuration (enabled).
func DefaultTenancyConfig() TenancyConfig {
	return TenancyConfig{
		Enabled:         true,
		DefaultTenantID: "default",
	}
}

// RequireAuth is middleware that requires a valid JWT token.
// Extracts claims and adds them to request context.
// TenancyConfig is optional - defaults to tenancy enabled if not provided.
func RequireAuth(jwtService jwt.Service, log logger.Logger, tenancy ...TenancyConfig) func(http.Handler) http.Handler {
	// Use default if not provided
	cfg := DefaultTenancyConfig()
	if len(tenancy) > 0 {
		cfg = tenancy[0]
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			token := extractBearerToken(r)
			if token == "" {
				log.Warn("missing authorization token",
					logger.String("path", r.URL.Path),
					logger.String("method", r.Method),
				)
				response.Unauthorized(w, "missing or invalid authorization token")
				return
			}

			// Validate token
			claims, err := jwtService.Validate(r.Context(), token)
			if err != nil {
				log.Warn("invalid JWT token",
					logger.String("path", r.URL.Path),
					logger.String("method", r.Method),
					logger.Err(err),
				)
				response.Unauthorized(w, "invalid or expired token")
				return
			}

			// Add claims to context
			ctx := context.WithValue(r.Context(), ContextKeyClaims, claims)

			// Add individual claim values for convenience
			ctx = addClaimsToContext(ctx, claims, cfg)

			// Continue with claims in context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractBearerToken extracts the JWT token from Authorization header.
// Expected format: "Authorization: Bearer <token>"
func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Split "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

// addClaimsToContext adds individual claim values to context for easy access.
func addClaimsToContext(ctx context.Context, claims jwt.Claims, tenancy TenancyConfig) context.Context {
	// Add standard claims
	if subject := claims.Subject(); subject != "" {
		ctx = context.WithValue(ctx, "user_id", subject)
	}

	// Add custom claims
	if userID := claims.GetString("user_id"); userID != "" {
		ctx = context.WithValue(ctx, "user_id", userID)
	}

	// Handle tenant ID based on tenancy configuration
	if tenancy.Enabled {
		if tenantID := claims.GetString("tenant_id"); tenantID != "" {
			ctx = context.WithValue(ctx, "tenant_id", tenantID)
		}
	} else {
		// Tenancy disabled - use default tenant
		ctx = context.WithValue(ctx, "tenant_id", tenancy.DefaultTenantID)
	}

	// Store tenancy enabled flag in context
	ctx = context.WithValue(ctx, "tenancy_enabled", tenancy.Enabled)

	if sessionID := claims.GetString("session_id"); sessionID != "" {
		ctx = context.WithValue(ctx, "session_id", sessionID)
	}

	if email := claims.GetString("email"); email != "" {
		ctx = context.WithValue(ctx, "email", email)
	}

	if role := claims.GetString("role"); role != "" {
		ctx = context.WithValue(ctx, "role", role)
	}

	return ctx
}

// GetClaims extracts JWT claims from request context.
func GetClaims(ctx context.Context) (jwt.Claims, bool) {
	claims, ok := ctx.Value(ContextKeyClaims).(jwt.Claims)
	return claims, ok
}

// GetUserID extracts user ID from request context.
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}

// GetTenantID extracts tenant ID from request context.
func GetTenantID(ctx context.Context) string {
	if tenantID, ok := ctx.Value("tenant_id").(string); ok {
		return tenantID
	}
	return ""
}

// GetSessionID extracts session ID from request context.
func GetSessionID(ctx context.Context) string {
	if sessionID, ok := ctx.Value("session_id").(string); ok {
		return sessionID
	}
	return ""
}

// GetEmail extracts email from request context.
func GetEmail(ctx context.Context) string {
	if email, ok := ctx.Value("email").(string); ok {
		return email
	}
	return ""
}

// GetRole extracts role from request context.
func GetRole(ctx context.Context) string {
	if role, ok := ctx.Value("role").(string); ok {
		return role
	}
	return ""
}

// IsTenancyEnabled extracts tenancy enabled flag from request context.
func IsTenancyEnabled(ctx context.Context) bool {
	if enabled, ok := ctx.Value("tenancy_enabled").(bool); ok {
		return enabled
	}
	return true // default to enabled
}
