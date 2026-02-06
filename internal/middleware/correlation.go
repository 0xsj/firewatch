package middleware

import (
	"context"
	"net/http"

	"github.com/0xsj/firewatch/pkg/crypto"
)

type contextKey string

const requestIDKey contextKey = "request_id"

// Correlation assigns a unique request ID to every incoming request.
// If the request already has an X-Request-Id header, it is reused.
// The ID is stored in the request context and echoed in the response.
func Correlation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = crypto.UUID4()
		}

		ctx := context.WithValue(r.Context(), requestIDKey, id)
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestID extracts the request ID from the context.
// Returns empty string if no ID is present.
func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}
