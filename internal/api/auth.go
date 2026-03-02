package api

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

// APIKeyAuth returns middleware that requires a valid X-Api-Key header.
// The /health endpoint is exempt from authentication.
func APIKeyAuth(apiKey string) func(http.Handler) http.Handler {
	keyBytes := []byte(apiKey)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// /health is exempt from auth.
			if strings.HasSuffix(r.URL.Path, "/health") {
				next.ServeHTTP(w, r)
				return
			}

			provided := r.Header.Get("X-Api-Key")
			if provided == "" {
				writeError(w, http.StatusUnauthorized, "missing X-Api-Key header")
				return
			}

			if subtle.ConstantTimeCompare([]byte(provided), keyBytes) != 1 {
				writeError(w, http.StatusUnauthorized, "invalid API key")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
