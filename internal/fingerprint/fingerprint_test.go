package fingerprint

import (
	"net/http/httptest"
	"testing"
)

func TestEngine_AnalyzeWithoutJA3(t *testing.T) {
	engine := NewEngine(nil)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("User-Agent", "curl/8.0")
	req.Header.Set("Accept", "*/*")

	result := engine.Analyze(req)

	if result.KnownClient != "curl" {
		t.Errorf("KnownClient = %q, want curl", result.KnownClient)
	}
	if result.HeaderOrderHash == "" {
		t.Error("HeaderOrderHash is empty")
	}
	if result.UserAgent != "curl/8.0" {
		t.Errorf("UserAgent = %q, want curl/8.0", result.UserAgent)
	}
	// JA3 should be empty without TLS.
	if result.JA3Hash != "" {
		t.Errorf("JA3Hash = %q, want empty (no TLS)", result.JA3Hash)
	}
}

func TestAnalyzeHeaders_KnownClients(t *testing.T) {
	tests := []struct {
		ua   string
		want string
	}{
		{"python-requests/2.31.0", "python-requests"},
		{"Go-http-client/1.1", "go-http-client"},
		{"curl/8.4.0", "curl"},
		{"Nuclei - Open-source project", "nuclei"},
		{"sqlmap/1.7", "sqlmap"},
		{"Mozilla/5.0 (compatible; Nmap Scripting Engine)", "nmap"},
		{"Mozilla/5.0 (Windows NT 10.0; Win64) Chrome/120", ""},
	}

	for _, tt := range tests {
		t.Run(tt.ua, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("User-Agent", tt.ua)

			fp := AnalyzeHeaders(req)
			if fp.KnownClient != tt.want {
				t.Errorf("KnownClient = %q, want %q", fp.KnownClient, tt.want)
			}
		})
	}
}

func TestAnalyzeHeaders_Anomalies(t *testing.T) {
	// Request with no standard browser headers.
	req := httptest.NewRequest("GET", "/", nil)
	// Remove default headers by creating a bare request.
	req.Header = map[string][]string{}

	fp := AnalyzeHeaders(req)

	wantAnomalies := map[string]bool{
		"missing_accept":          true,
		"missing_accept_language": true,
		"missing_accept_encoding": true,
		"missing_user_agent":      true,
	}

	for _, a := range fp.Anomalies {
		delete(wantAnomalies, a)
	}
	for missing := range wantAnomalies {
		t.Errorf("missing expected anomaly: %s", missing)
	}
}

func TestAnalyzeHeaders_NoAnomaliesForBrowser(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Accept-Language", "en-US")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	fp := AnalyzeHeaders(req)

	if len(fp.Anomalies) != 0 {
		t.Errorf("expected no anomalies, got %v", fp.Anomalies)
	}
}

func TestContextRoundtrip(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	result := Result{
		JA3Hash:     "abc123",
		KnownClient: "curl",
	}

	ctx := WithResult(req.Context(), result)
	got := GetResult(ctx)

	if got.JA3Hash != "abc123" {
		t.Errorf("JA3Hash = %q, want abc123", got.JA3Hash)
	}
	if got.KnownClient != "curl" {
		t.Errorf("KnownClient = %q, want curl", got.KnownClient)
	}
}

func TestGetResult_EmptyContext(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	got := GetResult(req.Context())

	if got.JA3Hash != "" || got.KnownClient != "" {
		t.Errorf("expected empty result, got %+v", got)
	}
}
