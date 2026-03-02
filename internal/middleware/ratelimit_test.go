package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/0xsj/firewatch/internal/storage"
	"github.com/0xsj/firewatch/internal/storage/models"
)

// mockStore implements storage.Store for testing
type mockRateLimitStore struct {
	mu     sync.Mutex
	events []*models.Event
}

func (m *mockRateLimitStore) SaveEvent(_ context.Context, event *models.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
	return nil
}

func (m *mockRateLimitStore) GetEvent(_ context.Context, _ string) (*models.Event, error) {
	return nil, nil
}

func (m *mockRateLimitStore) ListEvents(_ context.Context, _ storage.EventFilter) ([]*models.Event, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.events, nil
}

func (m *mockRateLimitStore) CountEvents(_ context.Context, _ storage.EventFilter) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return int64(len(m.events)), nil
}

func (m *mockRateLimitStore) SaveAttacker(_ context.Context, _ *models.Attacker) error {
	return nil
}

func (m *mockRateLimitStore) GetAttacker(_ context.Context, _ string) (*models.Attacker, error) {
	return nil, nil
}

func (m *mockRateLimitStore) GetAttackerByIP(_ context.Context, _ string) (*models.Attacker, error) {
	return nil, nil
}

func (m *mockRateLimitStore) ListAttackers(_ context.Context, _ storage.AttackerFilter) ([]*models.Attacker, error) {
	return nil, nil
}

func (m *mockRateLimitStore) SaveCampaign(_ context.Context, _ *models.Campaign) error {
	return nil
}

func (m *mockRateLimitStore) GetCampaign(_ context.Context, _ string) (*models.Campaign, error) {
	return nil, nil
}

func (m *mockRateLimitStore) ListCampaigns(_ context.Context, _ storage.CampaignFilter) ([]*models.Campaign, error) {
	return nil, nil
}

func (m *mockRateLimitStore) SaveIOC(_ context.Context, _ *models.IOC) error {
	return nil
}

func (m *mockRateLimitStore) ListIOCs(_ context.Context, _ storage.IOCFilter) ([]*models.IOC, error) {
	return nil, nil
}

func (m *mockRateLimitStore) UpdateEventLinks(_ context.Context, _, _, _ string) error {
	return nil
}

func (m *mockRateLimitStore) SaveHoneyToken(context.Context, *models.HoneyToken) error { return nil }
func (m *mockRateLimitStore) GetHoneyTokenByValue(context.Context, string) (*models.HoneyToken, error) {
	return nil, nil
}
func (m *mockRateLimitStore) ListHoneyTokens(context.Context, storage.HoneyTokenFilter) ([]*models.HoneyToken, error) {
	return nil, nil
}

func (m *mockRateLimitStore) Close() error {
	return nil
}

// Helper to add request ID to context for testing
func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func TestRateLimit_AllowsUnderLimit(t *testing.T) {
	store := &mockRateLimitStore{}
	cfg := RateLimiterConfig{
		RequestsPerSecond: 10,
		Burst:             20,
		CleanupInterval:   1 * time.Minute,
	}
	limiter := NewRateLimiter(cfg, store, testLogger())
	defer limiter.Stop()

	handler := RateLimit(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Send 5 requests (well under limit)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		req = req.WithContext(withRequestID(req.Context(), fmt.Sprintf("req-%d", i)))

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Request %d: got status %d, want %d", i, rr.Code, http.StatusOK)
		}
	}

	// Should have no rate limit events
	if len(store.events) != 0 {
		t.Errorf("Expected 0 rate limit events, got %d", len(store.events))
	}
}

func TestRateLimit_BlocksOverLimit(t *testing.T) {
	store := &mockRateLimitStore{}
	cfg := RateLimiterConfig{
		RequestsPerSecond: 1, // Very low rate
		Burst:             2, // Small burst
		CleanupInterval:   1 * time.Minute,
	}
	limiter := NewRateLimiter(cfg, store, testLogger())
	defer limiter.Stop()

	handler := RateLimit(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First 2 requests should succeed (burst allowance)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		req = req.WithContext(withRequestID(req.Context(), fmt.Sprintf("req-%d", i)))

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Request %d: got status %d, want %d", i, rr.Code, http.StatusOK)
		}
	}

	// 3rd request should be blocked
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	req = req.WithContext(withRequestID(req.Context(), "req-3"))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", rr.Code)
	}

	// Should have recorded 1 rate limit event
	if len(store.events) != 1 {
		t.Fatalf("Expected 1 rate limit event, got %d", len(store.events))
	}

	event := store.events[0]
	if event.Module != "rate_limit" {
		t.Errorf("Expected module 'rate_limit', got %q", event.Module)
	}
	if event.Severity != "medium" {
		t.Errorf("Expected severity 'medium', got %q", event.Severity)
	}
}

func TestRateLimit_PerIPIsolation(t *testing.T) {
	store := &mockRateLimitStore{}
	cfg := RateLimiterConfig{
		RequestsPerSecond: 1,
		Burst:             1,
		CleanupInterval:   1 * time.Minute,
	}
	limiter := NewRateLimiter(cfg, store, testLogger())
	defer limiter.Stop()

	handler := RateLimit(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// IP 1: Send 1 request (uses burst)
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:1234"
	req1 = req1.WithContext(withRequestID(req1.Context(), "req-ip1-1"))
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("IP1 request 1: got status %d, want 200", rr1.Code)
	}

	// IP 2: Send 1 request (uses burst, different IP)
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.2:5678"
	req2 = req2.WithContext(withRequestID(req2.Context(), "req-ip2-1"))
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("IP2 request 1: got status %d, want 200", rr2.Code)
	}

	// IP 1: Send 2nd request (should be blocked)
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "192.168.1.1:1234"
	req3 = req3.WithContext(withRequestID(req3.Context(), "req-ip1-2"))
	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, req3)

	if rr3.Code != http.StatusTooManyRequests {
		t.Errorf("IP1 request 2: got status %d, want 429", rr3.Code)
	}

	// IP 2: Send 2nd request (should also be blocked, separate limit)
	req4 := httptest.NewRequest("GET", "/test", nil)
	req4.RemoteAddr = "192.168.1.2:5678"
	req4 = req4.WithContext(withRequestID(req4.Context(), "req-ip2-2"))
	rr4 := httptest.NewRecorder()
	handler.ServeHTTP(rr4, req4)

	if rr4.Code != http.StatusTooManyRequests {
		t.Errorf("IP2 request 2: got status %d, want 429", rr4.Code)
	}

	// Should have 2 rate limit events (one per IP)
	if len(store.events) != 2 {
		t.Errorf("Expected 2 rate limit events, got %d", len(store.events))
	}
}

func TestRateLimit_Refill(t *testing.T) {
	store := &mockRateLimitStore{}
	cfg := RateLimiterConfig{
		RequestsPerSecond: 10, // 10 per second
		Burst:             5,
		CleanupInterval:   1 * time.Minute,
	}
	limiter := NewRateLimiter(cfg, store, testLogger())
	defer limiter.Stop()

	handler := RateLimit(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Use up burst
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		req = req.WithContext(withRequestID(req.Context(), fmt.Sprintf("req-burst-%d", i)))

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Burst request %d: got status %d, want 200", i, rr.Code)
		}
	}

	// Next request should be blocked
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	req = req.WithContext(withRequestID(req.Context(), "req-blocked"))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", rr.Code)
	}

	// Wait for token refill (200ms = 2 tokens at 10/sec)
	time.Sleep(200 * time.Millisecond)

	// Should be able to make 2 more requests
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		req = req.WithContext(withRequestID(req.Context(), fmt.Sprintf("req-refill-%d", i)))

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Refill request %d: got status %d, want 200", i, rr.Code)
		}
	}
}

func TestRateLimit_CleanupStale(t *testing.T) {
	store := &mockRateLimitStore{}
	cfg := RateLimiterConfig{
		RequestsPerSecond: 10,
		Burst:             20,
		CleanupInterval:   100 * time.Millisecond, // Fast cleanup for testing
	}
	limiter := NewRateLimiter(cfg, store, testLogger())
	defer limiter.Stop()

	handler := RateLimit(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Create limiters for multiple IPs
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = fmt.Sprintf("192.168.1.%d:1234", i)
		req = req.WithContext(withRequestID(req.Context(), fmt.Sprintf("req-%d", i)))

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	// Check limiter count
	limiter.mu.RLock()
	initialCount := len(limiter.limiters)
	limiter.mu.RUnlock()

	if initialCount != 5 {
		t.Errorf("Expected 5 limiters, got %d", initialCount)
	}

	// Wait for cleanup (stale threshold = 2 * cleanup interval = 200ms)
	time.Sleep(250 * time.Millisecond)

	// Limiters should be cleaned up
	limiter.mu.RLock()
	finalCount := len(limiter.limiters)
	limiter.mu.RUnlock()

	if finalCount != 0 {
		t.Errorf("Expected 0 limiters after cleanup, got %d", finalCount)
	}
}

func TestRateLimit_NilLimiter(t *testing.T) {
	// Nil limiter should pass through without rate limiting
	handler := RateLimit(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestRateLimit_Headers(t *testing.T) {
	store := &mockRateLimitStore{}
	cfg := RateLimiterConfig{
		RequestsPerSecond: 1,
		Burst:             1,
		CleanupInterval:   1 * time.Minute,
	}
	limiter := NewRateLimiter(cfg, store, testLogger())
	defer limiter.Stop()

	handler := RateLimit(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Use up burst
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:1234"
	req1 = req1.WithContext(withRequestID(req1.Context(), "req-1"))
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)

	// Next request should be blocked with headers
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.1:1234"
	req2 = req2.WithContext(withRequestID(req2.Context(), "req-2"))
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", rr2.Code)
	}

	// Check headers
	if rr2.Header().Get("Retry-After") == "" {
		t.Error("Missing Retry-After header")
	}
	if rr2.Header().Get("X-RateLimit-Limit") == "" {
		t.Error("Missing X-RateLimit-Limit header")
	}
	if rr2.Header().Get("X-RateLimit-Remaining") != "0" {
		t.Errorf("Expected X-RateLimit-Remaining: 0, got %q", rr2.Header().Get("X-RateLimit-Remaining"))
	}
}
