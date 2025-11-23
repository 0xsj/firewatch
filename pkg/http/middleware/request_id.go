package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

const (
	// RequestIDKey is the context key for request ID.
	RequestIDKey ContextKey = "request_id"
)

// RequestID is a middleware that generates a unique request ID for each request.
// The ID is added to the request context and response headers.
//
// Header: X-Request-ID
//
// Usage with Chi:
//
//	r := chi.NewRouter()
//	r.Use(middleware.RequestID)
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request already has an ID (from proxy/load balancer)
		requestID := r.Header.Get("X-Request-ID")

		// Generate new ID if not present
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Add to response header
		w.Header().Set("X-Request-ID", requestID)

		// Add to request context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// generateRequestID generates a random 16-byte hex string.
// Format: 32 characters (16 bytes in hex)
// Example: "3f2504e04f8911e3a9d70800200c9a66"
func generateRequestID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to a simple counter if random fails (shouldn't happen)
		return "fallback-id"
	}
	return hex.EncodeToString(bytes)
}

// GetRequestID extracts the request ID from context.
// Returns empty string if not found.
//
// Example:
//
//	requestID := middleware.GetRequestID(r.Context())
//	log.Info("processing request", "request_id", requestID)
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// MustGetRequestID extracts the request ID and panics if not found.
// Only use this when you're certain the middleware has run.
func MustGetRequestID(ctx context.Context) string {
	id := GetRequestID(ctx)
	if id == "" {
		panic("request ID not found in context - did you forget to add RequestID middleware?")
	}
	return id
}
