package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIKeyAuth_MissingKey(t *testing.T) {
	handler := newTestHandler(&mockStore{})

	req := httptest.NewRequest("GET", "/api/v1/events", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["error"] != "missing X-Api-Key header" {
		t.Errorf("error = %q, want 'missing X-Api-Key header'", body["error"])
	}
}

func TestAPIKeyAuth_WrongKey(t *testing.T) {
	handler := newTestHandler(&mockStore{})

	req := httptest.NewRequest("GET", "/api/v1/events", nil)
	req.Header.Set("X-Api-Key", "wrong-key")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["error"] != "invalid API key" {
		t.Errorf("error = %q, want 'invalid API key'", body["error"])
	}
}

func TestAPIKeyAuth_ValidKey(t *testing.T) {
	handler := newTestHandler(&mockStore{})

	req := httptest.NewRequest("GET", "/api/v1/events", nil)
	req.Header.Set("X-Api-Key", "test-key")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestAPIKeyAuth_HealthExempt(t *testing.T) {
	handler := newTestHandler(&mockStore{})

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	// No X-Api-Key header.
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 (health should be exempt)", rec.Code)
	}
}
