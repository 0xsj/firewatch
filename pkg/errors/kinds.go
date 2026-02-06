package errors

import "net/http"

// Kind classifies an error into a broad category. Unlike Code which
// identifies specific failures, Kind groups errors by their nature
// so they can be mapped uniformly to HTTP statuses, log levels, and
// alert severities.
type Kind uint8

const (
	KindUnexpected   Kind = iota // Zero value — unclassified error
	KindNotFound                 // Resource does not exist
	KindValidation               // Input failed validation
	KindUnauthorized             // Authentication required
	KindForbidden                // Insufficient permissions
	KindConflict                 // State conflict (duplicate, version mismatch)
	KindTimeout                  // Operation exceeded deadline
	KindInternal                 // Unexpected internal failure
	KindUnavailable              // Dependency or service unavailable
	KindRateLimit                // Rate limit exceeded
	KindCanceled                 // Operation was canceled
)

var kindNames = [...]string{
	KindUnexpected:   "unexpected",
	KindNotFound:     "not found",
	KindValidation:   "validation",
	KindUnauthorized: "unauthorized",
	KindForbidden:    "forbidden",
	KindConflict:     "conflict",
	KindTimeout:      "timeout",
	KindInternal:     "internal",
	KindUnavailable:  "unavailable",
	KindRateLimit:    "rate limit",
	KindCanceled:     "canceled",
}

// String returns the human-readable name of the Kind.
func (k Kind) String() string {
	if int(k) < len(kindNames) {
		return kindNames[k]
	}
	return "unknown"
}

// HTTPStatus maps a Kind to the appropriate HTTP status code.
func (k Kind) HTTPStatus() int {
	switch k {
	case KindNotFound:
		return http.StatusNotFound
	case KindValidation:
		return http.StatusBadRequest
	case KindUnauthorized:
		return http.StatusUnauthorized
	case KindForbidden:
		return http.StatusForbidden
	case KindConflict:
		return http.StatusConflict
	case KindTimeout:
		return http.StatusGatewayTimeout
	case KindUnavailable:
		return http.StatusServiceUnavailable
	case KindRateLimit:
		return http.StatusTooManyRequests
	case KindCanceled:
		return http.StatusRequestTimeout
	default:
		return http.StatusInternalServerError
	}
}
