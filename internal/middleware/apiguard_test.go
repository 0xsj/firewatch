package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIGuard_MatchingPrefix(t *testing.T) {
	apiCalled := false
	apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true
		w.WriteHeader(http.StatusOK)
	})

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := APIGuard("/api/v1/", apiHandler)(next)

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !apiCalled {
		t.Error("expected API handler to be called for /api/v1/ prefix")
	}
	if nextCalled {
		t.Error("next handler should not be called for /api/v1/ prefix")
	}
}

func TestAPIGuard_NonMatchingPrefix(t *testing.T) {
	apiCalled := false
	apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true
	})

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := APIGuard("/api/v1/", apiHandler)(next)

	req := httptest.NewRequest("GET", "/wp-login.php", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if apiCalled {
		t.Error("API handler should not be called for non-API path")
	}
	if !nextCalled {
		t.Error("expected next handler to be called for non-API path")
	}
}

func TestAPIGuard_ExactPrefix(t *testing.T) {
	apiCalled := false
	apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := APIGuard("/api/v1/", apiHandler)(http.NotFoundHandler())

	req := httptest.NewRequest("GET", "/api/v1/events", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !apiCalled {
		t.Error("expected API handler to be called for /api/v1/events")
	}
	_ = rec
}
