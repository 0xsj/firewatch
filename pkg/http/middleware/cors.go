package middleware

import (
	"net/http"
	"strconv"
	"strings"
)

// CORSConfig configures CORS behavior.
type CORSConfig struct {
	// AllowedOrigins is a list of origins that are allowed to make cross-origin requests.
	// Use "*" to allow all origins (development only - not recommended for production).
	// Examples: ["https://myapp.com", "https://admin.myapp.com"]
	AllowedOrigins []string

	// AllowedMethods is a list of HTTP methods allowed for cross-origin requests.
	// Default: ["GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"]
	AllowedMethods []string

	// AllowedHeaders is a list of headers allowed in cross-origin requests.
	// Default: ["Accept", "Authorization", "Content-Type", "X-Request-ID"]
	AllowedHeaders []string

	// ExposedHeaders is a list of headers exposed to the browser.
	// Browsers can only access these headers from the response.
	// Default: ["X-Request-ID"]
	ExposedHeaders []string

	// AllowCredentials indicates whether credentials (cookies, auth headers) are allowed.
	// If true, AllowedOrigins cannot be "*"
	// Default: true
	AllowCredentials bool

	// MaxAge indicates how long (in seconds) preflight results can be cached.
	// Default: 86400 (24 hours)
	MaxAge int
}

// DefaultCORSConfig returns a secure default CORS configuration.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{}, // Must be explicitly configured
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-Request-ID",
		},
		ExposedHeaders: []string{
			"X-Request-ID",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}

// DevelopmentCORSConfig returns a permissive CORS configuration for development.
// WARNING: Do not use in production!
func DevelopmentCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"*"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"*"},
		AllowCredentials: false, // Cannot use credentials with "*"
		MaxAge:           86400,
	}
}

// CORS is a middleware that handles Cross-Origin Resource Sharing (CORS).
//
// IMPORTANT: CORS is NOT a security feature!
// It's a browser mechanism that allows/blocks JavaScript from reading responses.
// Non-browser clients (curl, Postman, backend services) ignore CORS entirely.
//
// For actual API security, use authentication and authorization middleware.
//
// Usage with Chi:
//
//	config := middleware.DefaultCORSConfig()
//	config.AllowedOrigins = []string{"https://myapp.com", "https://admin.myapp.com"}
//	r.Use(middleware.CORS(config))
func CORS(config CORSConfig) func(next http.Handler) http.Handler {
	// Validate configuration
	if len(config.AllowedOrigins) == 0 {
		config.AllowedOrigins = []string{"*"}
	}
	if len(config.AllowedMethods) == 0 {
		config.AllowedMethods = DefaultCORSConfig().AllowedMethods
	}
	if len(config.AllowedHeaders) == 0 {
		config.AllowedHeaders = DefaultCORSConfig().AllowedHeaders
	}
	if config.MaxAge == 0 {
		config.MaxAge = 86400
	}

	// Pre-compute header values for performance
	allowedMethodsStr := strings.Join(config.AllowedMethods, ", ")
	allowedHeadersStr := strings.Join(config.AllowedHeaders, ", ")
	exposedHeadersStr := strings.Join(config.ExposedHeaders, ", ")
	maxAgeStr := strconv.Itoa(config.MaxAge)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// If no origin header, this is not a cross-origin request
			// (same-origin requests don't send Origin header)
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Check if origin is allowed
			allowedOrigin := getAllowedOrigin(origin, config.AllowedOrigins)
			if allowedOrigin == "" {
				// Origin not allowed - don't set CORS headers
				// Browser will block the response
				next.ServeHTTP(w, r)
				return
			}

			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)

			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if len(config.ExposedHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", exposedHeadersStr)
			}

			// Handle preflight (OPTIONS) request
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", allowedMethodsStr)
				w.Header().Set("Access-Control-Allow-Headers", allowedHeadersStr)
				w.Header().Set("Access-Control-Max-Age", maxAgeStr)
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// Handle actual request
			next.ServeHTTP(w, r)
		})
	}
}

// getAllowedOrigin checks if the origin is allowed and returns it.
// Returns empty string if origin is not allowed.
func getAllowedOrigin(origin string, allowedOrigins []string) string {
	// Check for wildcard
	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			return "*"
		}
		if allowed == origin {
			return origin
		}
	}
	return ""
}
