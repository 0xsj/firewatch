package api

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/0xsj/firewatch/internal/storage"
	"github.com/0xsj/firewatch/internal/storage/models"
)

// mockStore implements storage.Store with canned data.
type mockStore struct {
	events   []*models.Event
	attackers []*models.Attacker
	campaigns []*models.Campaign
	iocs     []*models.IOC
	tokens   []*models.HoneyToken
}

func (m *mockStore) SaveEvent(ctx context.Context, event *models.Event) error { return nil }
func (m *mockStore) GetEvent(ctx context.Context, id string) (*models.Event, error) { return nil, nil }
func (m *mockStore) ListEvents(ctx context.Context, f storage.EventFilter) ([]*models.Event, error) {
	return m.events, nil
}
func (m *mockStore) CountEvents(ctx context.Context, f storage.EventFilter) (int64, error) {
	return int64(len(m.events)), nil
}
func (m *mockStore) UpdateEventLinks(ctx context.Context, eventID, attackerID, campaignID string) error {
	return nil
}
func (m *mockStore) SaveAttacker(ctx context.Context, attacker *models.Attacker) error { return nil }
func (m *mockStore) GetAttacker(ctx context.Context, id string) (*models.Attacker, error) {
	return nil, nil
}
func (m *mockStore) GetAttackerByIP(ctx context.Context, ip string) (*models.Attacker, error) {
	return nil, nil
}
func (m *mockStore) ListAttackers(ctx context.Context, f storage.AttackerFilter) ([]*models.Attacker, error) {
	return m.attackers, nil
}
func (m *mockStore) SaveCampaign(ctx context.Context, campaign *models.Campaign) error { return nil }
func (m *mockStore) GetCampaign(ctx context.Context, id string) (*models.Campaign, error) {
	return nil, nil
}
func (m *mockStore) ListCampaigns(ctx context.Context, f storage.CampaignFilter) ([]*models.Campaign, error) {
	return m.campaigns, nil
}
func (m *mockStore) SaveIOC(ctx context.Context, ioc *models.IOC) error { return nil }
func (m *mockStore) ListIOCs(ctx context.Context, f storage.IOCFilter) ([]*models.IOC, error) {
	return m.iocs, nil
}
func (m *mockStore) SaveHoneyToken(ctx context.Context, token *models.HoneyToken) error { return nil }
func (m *mockStore) GetHoneyTokenByValue(ctx context.Context, value string) (*models.HoneyToken, error) {
	return nil, nil
}
func (m *mockStore) ListHoneyTokens(ctx context.Context, f storage.HoneyTokenFilter) ([]*models.HoneyToken, error) {
	return m.tokens, nil
}
func (m *mockStore) Close() error { return nil }

func newTestHandler(store storage.Store) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := New(store, logger)
	return APIKeyAuth("test-key")(h)
}

func TestHealthNoAuth(t *testing.T) {
	handler := newTestHandler(&mockStore{})

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status = %q, want ok", body["status"])
	}
}

func TestEventsWithAuth(t *testing.T) {
	store := &mockStore{
		events: []*models.Event{
			{ID: "e1", Module: "nextjs", Severity: "medium", SourceIP: "1.2.3.4",
				Timestamp: time.Now().UTC().Format(time.RFC3339)},
		},
	}
	handler := newTestHandler(store)

	req := httptest.NewRequest("GET", "/api/v1/events", nil)
	req.Header.Set("X-Api-Key", "test-key")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}

	var events []models.Event
	if err := json.NewDecoder(rec.Body).Decode(&events); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("got %d events, want 1", len(events))
	}
}

func TestEventsNilSliceReturnsEmptyArray(t *testing.T) {
	store := &mockStore{events: nil}
	handler := newTestHandler(store)

	req := httptest.NewRequest("GET", "/api/v1/events", nil)
	req.Header.Set("X-Api-Key", "test-key")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}

	body, _ := io.ReadAll(rec.Body)
	// Should be "[]" not "null"
	trimmed := string(body)
	if trimmed == "null\n" || trimmed == "null" {
		t.Error("nil slice should serialize as [], not null")
	}
}

func TestEventsBadSince(t *testing.T) {
	handler := newTestHandler(&mockStore{})

	req := httptest.NewRequest("GET", "/api/v1/events?since=not-a-time", nil)
	req.Header.Set("X-Api-Key", "test-key")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestAttackers(t *testing.T) {
	store := &mockStore{
		attackers: []*models.Attacker{
			{ID: "a1", IP: "10.0.0.1", Severity: "high"},
		},
	}
	handler := newTestHandler(store)

	req := httptest.NewRequest("GET", "/api/v1/attackers", nil)
	req.Header.Set("X-Api-Key", "test-key")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestCampaigns(t *testing.T) {
	handler := newTestHandler(&mockStore{campaigns: nil})

	req := httptest.NewRequest("GET", "/api/v1/campaigns?active=true", nil)
	req.Header.Set("X-Api-Key", "test-key")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestIOCs(t *testing.T) {
	store := &mockStore{
		iocs: []*models.IOC{
			{ID: "ioc1", Type: models.IOCTypeIP, Value: "1.2.3.4"},
		},
	}
	handler := newTestHandler(store)

	req := httptest.NewRequest("GET", "/api/v1/iocs?type=ip", nil)
	req.Header.Set("X-Api-Key", "test-key")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestTokens(t *testing.T) {
	handler := newTestHandler(&mockStore{tokens: nil})

	req := httptest.NewRequest("GET", "/api/v1/tokens?kind=aws_access_key", nil)
	req.Header.Set("X-Api-Key", "test-key")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestStats(t *testing.T) {
	store := &mockStore{
		events: []*models.Event{
			{ID: "e1", Module: "wordpress", Severity: "high", SourceIP: "1.2.3.4",
				Signatures: []string{"wp-login-001"}},
			{ID: "e2", Module: "nextjs", Severity: "medium", SourceIP: "5.6.7.8"},
		},
	}
	handler := newTestHandler(store)

	req := httptest.NewRequest("GET", "/api/v1/stats", nil)
	req.Header.Set("X-Api-Key", "test-key")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}

	var stats statsResponse
	if err := json.NewDecoder(rec.Body).Decode(&stats); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if stats.TotalEvents != 2 {
		t.Errorf("total_events = %d, want 2", stats.TotalEvents)
	}
	if stats.UniqueIPs != 2 {
		t.Errorf("unique_ips = %d, want 2", stats.UniqueIPs)
	}
}

func TestParseSince(t *testing.T) {
	tests := []struct {
		input string
		ok    bool
	}{
		{"1h", true},
		{"24h", true},
		{"7d", true},
		{"2025-01-01", true},
		{"2025-01-01T00:00:00Z", true},
		{"", true},
		{"not-a-time", false},
	}

	for _, tt := range tests {
		_, err := parseSince(tt.input)
		if (err == nil) != tt.ok {
			t.Errorf("parseSince(%q): err=%v, wantOK=%v", tt.input, err, tt.ok)
		}
	}
}
