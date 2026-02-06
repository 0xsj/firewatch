package timeutil

import (
	"fmt"
	"time"
)

// NowUTC returns the current time in UTC.
func NowUTC() time.Time {
	return time.Now().UTC()
}

// FormatRFC3339 formats a time as RFC3339 in UTC.
func FormatRFC3339(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// ParseRFC3339 parses an RFC3339 timestamp string.
func ParseRFC3339(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing RFC3339 timestamp: %w", err)
	}
	return t.UTC(), nil
}

// ParseDuration parses a human-friendly duration string.
// Supports Go's time.ParseDuration format: "5m", "1h30m", "24h".
func ParseDuration(s string) (time.Duration, error) {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("parsing duration %q: %w", s, err)
	}
	return d, nil
}

// Since returns the duration elapsed since t in UTC.
func Since(t time.Time) time.Duration {
	return NowUTC().Sub(t.UTC())
}

// Ago returns the time that was d duration ago from now (UTC).
func Ago(d time.Duration) time.Time {
	return NowUTC().Add(-d)
}
