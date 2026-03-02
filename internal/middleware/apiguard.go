package middleware

import (
	"net/http"
	"strings"
)

// APIGuard returns a middleware that intercepts requests matching the
// given path prefix and routes them to the API handler. All other
// requests continue through the honeypot middleware chain.
func APIGuard(prefix string, apiHandler http.Handler) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, prefix) {
				apiHandler.ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
