package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIPFilter_AllowPassthrough(t *testing.T) {
	cfg, err := ParseIPFilter([]string{"192.168.1.1"}, []string{"10.0.0.0/8"})
	if err != nil {
		t.Fatal(err)
	}

	store := &mockStore{}
	handler := IPFilter(cfg, store, testLogger())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("allowed IP: got status %d, want 200", rr.Code)
	}
	if len(store.events) != 0 {
		t.Errorf("allowed IP should not generate events, got %d", len(store.events))
	}
}

func TestIPFilter_BlockReturns403(t *testing.T) {
	cfg, err := ParseIPFilter(nil, []string{"10.0.0.0/8"})
	if err != nil {
		t.Fatal(err)
	}

	store := &mockStore{}
	handler := IPFilter(cfg, store, testLogger())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.1.2.3:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("blocked IP: got status %d, want 403", rr.Code)
	}
	if len(store.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(store.events))
	}
	e := store.events[0]
	if e.Module != "ip_filter" {
		t.Errorf("module = %q, want ip_filter", e.Module)
	}
	if e.Severity != "high" {
		t.Errorf("severity = %q, want high", e.Severity)
	}
}

func TestIPFilter_CIDRMatching(t *testing.T) {
	cfg, err := ParseIPFilter(nil, []string{"192.168.0.0/16"})
	if err != nil {
		t.Fatal(err)
	}

	store := &mockStore{}
	handler := IPFilter(cfg, store, testLogger())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// IP in range should be blocked
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.50.100:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("CIDR blocked IP: got status %d, want 403", rr.Code)
	}

	// IP outside range should pass
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "10.0.0.1:1234"
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("non-blocked IP: got status %d, want 200", rr2.Code)
	}
}

func TestIPFilter_IPv6(t *testing.T) {
	cfg, err := ParseIPFilter(nil, []string{"::1"})
	if err != nil {
		t.Fatal(err)
	}

	store := &mockStore{}
	handler := IPFilter(cfg, store, testLogger())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "[::1]:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("blocked IPv6: got status %d, want 403", rr.Code)
	}
}

func TestIPFilter_NilPassthrough(t *testing.T) {
	handler := IPFilter(nil, nil, testLogger())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("nil config: got status %d, want 200", rr.Code)
	}
}

func TestIPFilter_AllowPrecedence(t *testing.T) {
	// IP is in both allowlist and blocklist — allowlist should win
	cfg, err := ParseIPFilter([]string{"10.0.0.1"}, []string{"10.0.0.0/8"})
	if err != nil {
		t.Fatal(err)
	}

	store := &mockStore{}
	handler := IPFilter(cfg, store, testLogger())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("allowlist precedence: got status %d, want 200", rr.Code)
	}
	if len(store.events) != 0 {
		t.Errorf("allowlist should skip event recording, got %d events", len(store.events))
	}
}

func TestIPFilter_UnknownIPPasses(t *testing.T) {
	cfg, err := ParseIPFilter([]string{"192.168.1.0/24"}, []string{"10.0.0.0/8"})
	if err != nil {
		t.Fatal(err)
	}

	store := &mockStore{}
	handler := IPFilter(cfg, store, testLogger())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "172.16.0.1:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("unknown IP: got status %d, want 200", rr.Code)
	}
}

func TestParseIPFilter_InvalidIP(t *testing.T) {
	_, err := ParseIPFilter(nil, []string{"not-an-ip"})
	if err == nil {
		t.Error("expected error for invalid IP")
	}
}
