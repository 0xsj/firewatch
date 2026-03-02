package nextjs

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/0xsj/firewatch/internal/config"
	"github.com/0xsj/firewatch/internal/storage"
	"github.com/0xsj/firewatch/internal/storage/models"
)

// mockStore implements storage.Store with just enough to capture events.
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
func (m *mockStore) SaveHoneyToken(context.Context, *models.HoneyToken) error       { return nil }
func (m *mockStore) GetHoneyTokenByValue(context.Context, string) (*models.HoneyToken, error) {
	return nil, nil
}
func (m *mockStore) ListHoneyTokens(context.Context, storage.HoneyTokenFilter) ([]*models.HoneyToken, error) {
	return nil, nil
}
func (m *mockStore) Close() error { return nil }

func newTestModule() (*NextJS, *mockStore) {
	store := &mockStore{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	deception := config.DeceptionConfig{HoneyTokens: true, Breadcrumbs: true, FakeErrors: true}
	mod := New(config.NextJSModuleConfig{Enabled: true}, deception, store, logger)
	return mod, store
}

func TestNextJS_Name(t *testing.T) {
	mod, _ := newTestModule()
	if mod.Name() != "nextjs" {
		t.Errorf("Name() = %q, want %q", mod.Name(), "nextjs")
	}
}

func TestNextJS_Routes(t *testing.T) {
	mod, _ := newTestModule()
	routes := mod.Routes()
	if len(routes) == 0 {
		t.Fatal("Routes() returned empty slice")
	}

	patterns := make(map[string]bool)
	for _, r := range routes {
		patterns[r.Pattern] = true
	}

	required := []string{"POST /", "GET /_next/static/", "GET /_rsc", "GET /"}
	for _, p := range required {
		if !patterns[p] {
			t.Errorf("missing route pattern: %s", p)
		}
	}
}

func TestNextJS_HandlePage(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	mod.handlePage(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
	if rec.Header().Get("X-Powered-By") != "Next.js" {
		t.Errorf("X-Powered-By = %q, want Next.js", rec.Header().Get("X-Powered-By"))
	}
	if !strings.Contains(rec.Body.String(), "__next") {
		t.Error("response body missing __next div")
	}
	if len(store.events) != 1 {
		t.Errorf("events recorded = %d, want 1", len(store.events))
	}
}

func TestNextJS_HandleServerAction(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"action":"test"}`))
	req.Header.Set("Next-Action", "abc123")

	mod.handleServerAction(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/x-component" {
		t.Errorf("Content-Type = %q, want text/x-component", ct)
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	e := store.events[0]
	if e.Severity != "high" {
		t.Errorf("severity = %q, want high", e.Severity)
	}
	if len(e.Signatures) < 2 {
		t.Errorf("signatures = %v, want at least 2 (probe + payload)", e.Signatures)
	}
}

func TestNextJS_HandleRSC(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/_rsc", nil)
	req.Header.Set("Rsc", "1")

	mod.handleRSC(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "medium" {
		t.Errorf("severity = %q, want medium", store.events[0].Severity)
	}
}

func TestNextJS_HandleDebugEndpoint(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/__nextjs_original-stack-frame", nil)

	mod.handleRSC(rec, req)

	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "high" {
		t.Errorf("severity = %q, want high", store.events[0].Severity)
	}
}

func TestNextJS_HandleStatic_BuildManifest(t *testing.T) {
	mod, _ := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/_next/static/chunks/buildManifest.js", nil)

	mod.handleStatic(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "__BUILD_MANIFEST") {
		t.Error("response missing build manifest content")
	}
}
