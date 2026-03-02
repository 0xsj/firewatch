package exposure

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

type mockStore struct {
	events      []*models.Event
	honeyTokens []*models.HoneyToken
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
func (m *mockStore) SaveHoneyToken(_ context.Context, token *models.HoneyToken) error {
	m.honeyTokens = append(m.honeyTokens, token)
	return nil
}
func (m *mockStore) GetHoneyTokenByValue(context.Context, string) (*models.HoneyToken, error) {
	return nil, nil
}
func (m *mockStore) ListHoneyTokens(context.Context, storage.HoneyTokenFilter) ([]*models.HoneyToken, error) {
	return nil, nil
}
func (m *mockStore) Close() error { return nil }

func newTestModule() (*Exposure, *mockStore) {
	store := &mockStore{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	deception := config.DeceptionConfig{HoneyTokens: true, Breadcrumbs: true, FakeErrors: true}
	mod := New(config.ExposureModuleConfig{Enabled: true}, deception, store, logger)
	return mod, store
}

func newTestModuleWithFakeEnv(fakeEnv string) (*Exposure, *mockStore) {
	store := &mockStore{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	deception := config.DeceptionConfig{HoneyTokens: true, Breadcrumbs: true, FakeErrors: true}
	mod := New(config.ExposureModuleConfig{Enabled: true, FakeEnv: fakeEnv}, deception, store, logger)
	return mod, store
}

func TestExposure_Name(t *testing.T) {
	mod, _ := newTestModule()
	if mod.Name() != "exposure" {
		t.Errorf("Name() = %q, want %q", mod.Name(), "exposure")
	}
}

func TestExposure_Routes(t *testing.T) {
	mod, _ := newTestModule()
	routes := mod.Routes()
	if len(routes) != 14 {
		t.Fatalf("Routes() returned %d routes, want 14", len(routes))
	}

	patterns := make(map[string]bool)
	for _, r := range routes {
		patterns[r.Pattern] = true
	}

	required := []string{
		"GET /.env",
		"GET /.env.local",
		"GET /.env.production",
		"GET /.env.backup",
		"GET /.git/",
		"GET /.git/config",
		"GET /.git/HEAD",
		"GET /config.php",
		"GET /wp-config.php",
		"GET /web.config",
		"GET /.htaccess",
		"GET /.htpasswd",
		"GET /.DS_Store",
	}
	for _, p := range required {
		if !patterns[p] {
			t.Errorf("missing route pattern: %s", p)
		}
	}
}

func TestExposure_EnvDefault(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/.env", nil)

	mod.handleEnv(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/plain") {
		t.Errorf("Content-Type = %q, want text/plain", ct)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "DB_PASSWORD") {
		t.Error("response body missing DB_PASSWORD from default env")
	}
	if !strings.Contains(body, "AWS_SECRET_ACCESS_KEY") {
		t.Error("response body missing AWS_SECRET_ACCESS_KEY from default env")
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "high" {
		t.Errorf("severity = %q, want high", store.events[0].Severity)
	}
}

func TestExposure_EnvCustom(t *testing.T) {
	customEnv := "CUSTOM_KEY=custom_value\nSECRET=mysecret\n"
	mod, store := newTestModuleWithFakeEnv(customEnv)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/.env", nil)

	mod.handleEnv(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if body != customEnv {
		t.Errorf("body = %q, want %q", body, customEnv)
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
}

func TestExposure_EnvLocal(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/.env.local", nil)

	mod.handleEnv(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "high" {
		t.Errorf("severity = %q, want high", store.events[0].Severity)
	}
}

func TestExposure_GitConfig(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/.git/config", nil)

	mod.handleGit(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "[core]") {
		t.Error("response body missing [core] section")
	}
	if !strings.Contains(body, "bare = false") {
		t.Error("response body missing bare = false")
	}
	if !strings.Contains(body, "git@github.com") {
		t.Error("response body missing remote URL")
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	e := store.events[0]
	if e.Severity != "high" {
		t.Errorf("severity = %q, want high", e.Severity)
	}
	if len(e.Signatures) < 2 {
		t.Errorf("signatures = %v, want at least 2", e.Signatures)
	}
}

func TestExposure_GitHEAD(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/.git/HEAD", nil)

	mod.handleGit(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "ref: refs/heads/main") {
		t.Errorf("body = %q, want ref: refs/heads/main", body)
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if len(store.events[0].Signatures) < 2 {
		t.Errorf("signatures = %v, want at least 2", store.events[0].Signatures)
	}
}

func TestExposure_GitDirectory(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/.git/", nil)

	mod.handleGit(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "high" {
		t.Errorf("severity = %q, want high", store.events[0].Severity)
	}
}

func TestExposure_ConfigPHP(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/config.php", nil)

	mod.handleConfig(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "medium" {
		t.Errorf("severity = %q, want medium", store.events[0].Severity)
	}
}

func TestExposure_WPConfig(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/wp-config.php", nil)

	mod.handleConfig(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
}

func TestExposure_Htaccess(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/.htaccess", nil)

	mod.handleConfig(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
}

func TestExposure_DSStore(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/.DS_Store", nil)

	mod.handleConfig(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
}
