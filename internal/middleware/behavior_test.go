package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/0xsj/firewatch/internal/detection"
)

func newTestBehaviorTracker() *detection.BehaviorTracker {
	return detection.NewBehaviorTracker(detection.BehaviorTrackerConfig{
		Window:          5 * time.Minute,
		SweepThreshold:  3,
		BruteThreshold:  3,
		ModuleThreshold: 3,
		CleanupInterval: 1 * time.Minute,
	})
}

func TestBehavior_NilTrackerPassthrough(t *testing.T) {
	called := false
	handler := Behavior(nil, nil, testLogger())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Error("handler was not called with nil tracker")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestBehavior_FewRequestsNoEvent(t *testing.T) {
	bt := newTestBehaviorTracker()
	defer bt.Stop()

	store := &mockStore{}
	handler := Behavior(bt, store, testLogger())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Only 2 requests — not enough for any detection
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	if len(store.events) != 0 {
		t.Errorf("expected 0 events for few requests, got %d", len(store.events))
	}
}

func TestBehavior_SweepTriggersEvent(t *testing.T) {
	bt := newTestBehaviorTracker()
	defer bt.Stop()

	store := &mockStore{}
	handler := Behavior(bt, store, testLogger())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	paths := []string{"/a", "/b", "/c", "/d", "/e"}
	for _, p := range paths {
		req := httptest.NewRequest("GET", p, nil)
		req.RemoteAddr = "1.2.3.4:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	// Should have at least one behavioral event
	found := false
	for _, e := range store.events {
		if e.Module == "behavior" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected behavioral event for scan sweep")
	}
}

func TestBehavior_BruteForceTriggersEvent(t *testing.T) {
	bt := newTestBehaviorTracker()
	defer bt.Stop()

	store := &mockStore{}
	handler := Behavior(bt, store, testLogger())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "/wp-login.php", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	found := false
	for _, e := range store.events {
		if e.Module == "behavior" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected behavioral event for brute force")
	}
}

func TestBehavior_CorrectModuleAndSignatures(t *testing.T) {
	bt := newTestBehaviorTracker()
	defer bt.Stop()

	store := &mockStore{}
	handler := Behavior(bt, store, testLogger())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Trigger brute force
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "/login", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	for _, e := range store.events {
		if e.Module == "behavior" {
			if len(e.Signatures) == 0 {
				t.Error("behavioral event has no signatures")
			}
			if e.Severity == "" {
				t.Error("behavioral event has no severity")
			}
			return
		}
	}
	t.Error("no behavioral event found")
}
