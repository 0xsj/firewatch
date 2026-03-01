package timeutil

import (
	"strings"
	"testing"
	"time"
)

func TestNowUTC(t *testing.T) {
	now := NowUTC()
	if now.Location() != time.UTC {
		t.Errorf("NowUTC() location = %v, want UTC", now.Location())
	}
	if time.Since(now) > time.Second {
		t.Error("NowUTC() returned a time too far in the past")
	}
}

func TestFormatRFC3339(t *testing.T) {
	ts := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	got := FormatRFC3339(ts)
	want := "2024-06-15T10:30:00Z"
	if got != want {
		t.Errorf("FormatRFC3339() = %q, want %q", got, want)
	}
}

func TestFormatRFC3339_NonUTC(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	ts := time.Date(2024, 6, 15, 10, 30, 0, 0, loc)
	got := FormatRFC3339(ts)
	if !strings.HasSuffix(got, "Z") {
		t.Errorf("FormatRFC3339() = %q, expected UTC suffix Z", got)
	}
}

func TestParseRFC3339(t *testing.T) {
	ts, err := ParseRFC3339("2024-06-15T10:30:00Z")
	if err != nil {
		t.Fatalf("ParseRFC3339() error: %v", err)
	}
	if ts.Year() != 2024 || ts.Month() != 6 || ts.Day() != 15 {
		t.Errorf("parsed time = %v, want 2024-06-15", ts)
	}
	if ts.Location() != time.UTC {
		t.Errorf("location = %v, want UTC", ts.Location())
	}
}

func TestParseRFC3339_Invalid(t *testing.T) {
	_, err := ParseRFC3339("not-a-timestamp")
	if err == nil {
		t.Error("expected error for invalid timestamp")
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input string
		want  time.Duration
	}{
		{"5m", 5 * time.Minute},
		{"1h30m", 90 * time.Minute},
		{"24h", 24 * time.Hour},
		{"500ms", 500 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseDuration(tt.input)
			if err != nil {
				t.Fatalf("ParseDuration(%q) error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseDuration(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseDuration_Invalid(t *testing.T) {
	_, err := ParseDuration("invalid")
	if err == nil {
		t.Error("expected error for invalid duration")
	}
}

func TestSince(t *testing.T) {
	past := NowUTC().Add(-5 * time.Second)
	d := Since(past)
	if d < 4*time.Second || d > 6*time.Second {
		t.Errorf("Since() = %v, expected ~5s", d)
	}
}

func TestAgo(t *testing.T) {
	ago := Ago(1 * time.Hour)
	expected := NowUTC().Add(-1 * time.Hour)
	diff := ago.Sub(expected)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("Ago(1h) = %v, expected ~%v", ago, expected)
	}
}
