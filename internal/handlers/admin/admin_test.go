package admin

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

func newTestModule() (*Admin, *mockStore) {
	store := &mockStore{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	deception := config.DeceptionConfig{HoneyTokens: true, Breadcrumbs: true, FakeErrors: true}
	mod := New(config.AdminModuleConfig{Enabled: true}, deception, store, logger)
	return mod, store
}

func newTestModuleNoDeception() (*Admin, *mockStore) {
	store := &mockStore{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	deception := config.DeceptionConfig{HoneyTokens: false, Breadcrumbs: false, FakeErrors: false}
	mod := New(config.AdminModuleConfig{Enabled: true}, deception, store, logger)
	return mod, store
}

func TestAdmin_Name(t *testing.T) {
	mod, _ := newTestModule()
	if mod.Name() != "admin" {
		t.Errorf("Name() = %q, want %q", mod.Name(), "admin")
	}
}

func TestAdmin_Routes(t *testing.T) {
	mod, _ := newTestModule()
	routes := mod.Routes()
	if len(routes) != 16 {
		t.Fatalf("Routes() returned %d routes, want 16", len(routes))
	}

	patterns := make(map[string]bool)
	for _, r := range routes {
		patterns[r.Pattern] = true
	}

	required := []string{
		"GET /phpmyadmin/",
		"POST /phpmyadmin/",
		"GET /adminer.php",
		"POST /adminer.php",
		"GET /cpanel",
		"POST /cpanel",
		"GET /admin",
		"POST /admin/login",
	}
	for _, p := range required {
		if !patterns[p] {
			t.Errorf("missing route pattern: %s", p)
		}
	}
}

func TestAdmin_PhpMyAdminGet(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/phpmyadmin/", nil)

	mod.handlePhpMyAdminGet(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
	if !strings.Contains(rec.Body.String(), "phpMyAdmin") {
		t.Error("response body missing phpMyAdmin")
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "medium" {
		t.Errorf("severity = %q, want medium", store.events[0].Severity)
	}
}

func TestAdmin_PhpMyAdminPost(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/phpmyadmin/",
		strings.NewReader("pma_username=root&pma_password=toor"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mod.handlePhpMyAdminPost(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
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

func TestAdmin_AdminerGet(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/adminer.php", nil)

	mod.handleAdminerGet(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "Adminer") {
		t.Error("response body missing Adminer")
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "medium" {
		t.Errorf("severity = %q, want medium", store.events[0].Severity)
	}
}

func TestAdmin_AdminerPost(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/adminer.php",
		strings.NewReader("auth%5Busername%5D=admin&auth%5Bpassword%5D=secret&auth%5Bserver%5D=localhost"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mod.handleAdminerPost(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
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

func TestAdmin_CPanelGet(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/cpanel", nil)

	mod.handleCPanelGet(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "cPanel") {
		t.Error("response body missing cPanel")
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "medium" {
		t.Errorf("severity = %q, want medium", store.events[0].Severity)
	}
}

func TestAdmin_CPanelPost(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/cpanel",
		strings.NewReader("user=admin&pass=password123"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mod.handleCPanelPost(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
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

func TestAdmin_GenericGet(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)

	mod.handleGenericGet(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "Admin Panel") {
		t.Error("response body missing Admin Panel")
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "medium" {
		t.Errorf("severity = %q, want medium", store.events[0].Severity)
	}
}

func TestAdmin_GenericPost(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/login",
		strings.NewReader("username=admin&password=admin123"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mod.handleGenericPost(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
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

func TestAdmin_GenericGet_BreadcrumbsInjected(t *testing.T) {
	mod, _ := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)

	mod.handleGenericGet(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "style=\"display:none\"") && !strings.Contains(body, "<!-- ") {
		t.Error("expected breadcrumb content in HTML when Breadcrumbs enabled")
	}
	// Breadcrumb headers should be set.
	if xpb := rec.Header().Get("X-Powered-By"); xpb == "" {
		t.Error("expected X-Powered-By header from breadcrumbs")
	}
}

func TestAdmin_GenericGet_NoBreadcrumbsWhenDisabled(t *testing.T) {
	mod, _ := newTestModuleNoDeception()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)

	mod.handleGenericGet(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "Admin Panel") {
		t.Error("response body missing Admin Panel")
	}
	if strings.Contains(body, "/solr/admin/cores") {
		t.Error("unexpected breadcrumb content when Breadcrumbs disabled")
	}
}

func TestAdmin_PhpMyAdminGet_BreadcrumbsInjected(t *testing.T) {
	mod, _ := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/phpmyadmin/", nil)

	mod.handlePhpMyAdminGet(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "style=\"display:none\"") && !strings.Contains(body, "<!-- ") {
		t.Error("expected breadcrumb content in phpMyAdmin HTML when Breadcrumbs enabled")
	}
}
