package ratelimit

import "time"

// Result contains the outcome of a rate limit check.
type Result struct {
	// Allowed indicates whether the request is permitted.
	Allowed bool

	// Limit is the maximum number of requests allowed in the window.
	Limit int

	// Remaining is the number of requests remaining in the current window.
	Remaining int

	// RetryAfter indicates when the client can retry if rate limited.
	// Zero value means retry information is not available.
	RetryAfter time.Duration

	// ResetAt indicates when the current window resets.
	ResetAt time.Time
}

// Denied returns true if the request was rate limited.
func (r Result) Denied() bool {
	return !r.Allowed
}

// Headers returns standard rate limit headers for HTTP responses.
// Returns a map suitable for setting response headers.
func (r Result) Headers() map[string]string {
	headers := map[string]string{
		"X-RateLimit-Limit":     itoa(r.Limit),
		"X-RateLimit-Remaining": itoa(r.Remaining),
	}

	if !r.ResetAt.IsZero() {
		headers["X-RateLimit-Reset"] = itoa64(r.ResetAt.Unix())
	}

	if r.RetryAfter > 0 {
		headers["Retry-After"] = itoa(int(r.RetryAfter.Seconds()))
	}

	return headers
}

// itoa converts an int to string without importing strconv.
func itoa(n int) string {
	return itoa64(int64(n))
}

// itoa64 converts an int64 to string without importing strconv.
func itoa64(n int64) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	var buf [20]byte
	i := len(buf)

	for n > 0 {
		i--
		buf[i] = byte(n%10) + '0'
		n /= 10
	}

	if negative {
		i--
		buf[i] = '-'
	}

	return string(buf[i:])
}
