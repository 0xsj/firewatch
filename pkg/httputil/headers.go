package httputil

import "net/http"

// Common header keys.
const (
	HeaderContentType   = "Content-Type"
	HeaderAuthorization = "Authorization"
	HeaderUserAgent     = "User-Agent"
	HeaderXForwardedFor = "X-Forwarded-For"
	HeaderXRealIP       = "X-Real-Ip"
	HeaderXRequestID    = "X-Request-Id"
	HeaderAccept        = "Accept"
	HeaderReferer       = "Referer"
	HeaderOrigin        = "Origin"
	HeaderHost          = "Host"
)

// Common content type values.
const (
	ContentTypeJSON = "application/json; charset=utf-8"
	ContentTypeHTML = "text/html; charset=utf-8"
	ContentTypeText = "text/plain; charset=utf-8"
	ContentTypeXML  = "application/xml; charset=utf-8"
)

// HeaderKeys returns all header keys present in h. Note that Go's
// http.Header is a map, so iteration order is not guaranteed.
// For true wire-order fingerprinting, use the fingerprint middleware.
func HeaderKeys(h http.Header) []string {
	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	return keys
}

// HeaderMap flattens headers to map[string]string using the first
// value for each key. Useful for logging and serialization.
func HeaderMap(h http.Header) map[string]string {
	m := make(map[string]string, len(h))
	for k, v := range h {
		if len(v) > 0 {
			m[k] = v[0]
		}
	}
	return m
}
