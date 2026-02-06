package httputil

import (
	"io"
	"net"
	"net/http"
	"strings"
)

// DefaultMaxBodySize is the default limit for reading request bodies (1MB).
const DefaultMaxBodySize = 1 << 20

// ReadBody reads the request body up to maxSize bytes.
// If maxSize is 0, DefaultMaxBodySize is used.
// Returns nil with no error if the body is nil.
func ReadBody(r *http.Request, maxSize int64) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	if maxSize <= 0 {
		maxSize = DefaultMaxBodySize
	}
	defer r.Body.Close()
	return io.ReadAll(io.LimitReader(r.Body, maxSize))
}

// ClientIP extracts the client IP from the request, checking
// X-Forwarded-For and X-Real-IP before falling back to RemoteAddr.
func ClientIP(r *http.Request) string {
	if xff := r.Header.Get(HeaderXForwardedFor); xff != "" {
		if i := strings.IndexByte(xff, ','); i != -1 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}

	if xri := r.Header.Get(HeaderXRealIP); xri != "" {
		return strings.TrimSpace(xri)
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// HasHeader reports whether the request contains the named header.
func HasHeader(r *http.Request, key string) bool {
	_, ok := r.Header[http.CanonicalHeaderKey(key)]
	return ok
}
