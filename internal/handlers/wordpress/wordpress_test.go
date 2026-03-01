package wordpress

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

func newTestModule() (*WordPress, *mockStore) {
	store := &mockStore{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mod := New(config.WordPressModuleConfig{Enabled: true, FakeVersion: "6.4.2"}, store, logger)
	return mod, store
}

func TestWordPress_Name(t *testing.T) {
	mod, _ := newTestModule()
	if mod.Name() != "wordpress" {
		t.Errorf("Name() = %q, want %q", mod.Name(), "wordpress")
	}
}

func TestWordPress_Routes(t *testing.T) {
	mod, _ := newTestModule()
	routes := mod.Routes()
	if len(routes) != 8 {
		t.Fatalf("Routes() returned %d routes, want 8", len(routes))
	}

	patterns := make(map[string]bool)
	for _, r := range routes {
		patterns[r.Pattern] = true
	}

	required := []string{
		"GET /wp-login.php",
		"POST /wp-login.php",
		"GET /wp-admin/",
		"POST /xmlrpc.php",
		"GET /xmlrpc.php",
		"GET /wp-json/",
		"GET /wp-includes/",
		"GET /wp-content/",
	}
	for _, p := range required {
		if !patterns[p] {
			t.Errorf("missing route pattern: %s", p)
		}
	}
}

func TestWordPress_LoginGet(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/wp-login.php", nil)

	mod.handleLoginGet(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
	if php := rec.Header().Get("X-Powered-By"); php != "PHP/8.1.0" {
		t.Errorf("X-Powered-By = %q, want PHP/8.1.0", php)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "WordPress") {
		t.Error("response body missing WordPress")
	}
	if !strings.Contains(body, "6.4.2") {
		t.Error("response body missing fake version 6.4.2")
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "medium" {
		t.Errorf("severity = %q, want medium", store.events[0].Severity)
	}
}

func TestWordPress_LoginPost(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/wp-login.php",
		strings.NewReader("log=admin&pwd=password123"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mod.handleLoginPost(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "WordPress") {
		t.Error("response body missing WordPress login page")
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "high" {
		t.Errorf("severity = %q, want high", store.events[0].Severity)
	}
}

func TestWordPress_LoginPostEmpty(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/wp-login.php",
		strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mod.handleLoginPost(rec, req)

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

func TestWordPress_XMLRPCGet(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/xmlrpc.php", nil)

	mod.handleXMLRPC(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/xml") {
		t.Errorf("Content-Type = %q, want text/xml", ct)
	}
	if !strings.Contains(rec.Body.String(), "methodResponse") {
		t.Error("response body missing XML-RPC response")
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "high" {
		t.Errorf("severity = %q, want high", store.events[0].Severity)
	}
}

func TestWordPress_XMLRPCPost(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/xmlrpc.php",
		strings.NewReader(`<?xml version="1.0"?><methodCall><methodName>system.listMethods</methodName></methodCall>`))
	req.Header.Set("Content-Type", "text/xml")

	mod.handleXMLRPC(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	e := store.events[0]
	if e.Severity != "critical" {
		t.Errorf("severity = %q, want critical", e.Severity)
	}
	if len(e.Signatures) < 2 {
		t.Errorf("signatures = %v, want at least 2", e.Signatures)
	}
}

func TestWordPress_Admin(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/wp-admin/", nil)

	mod.handleAdmin(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	loc := rec.Header().Get("Location")
	if !strings.Contains(loc, "wp-login.php") {
		t.Errorf("Location = %q, want redirect to wp-login.php", loc)
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "medium" {
		t.Errorf("severity = %q, want medium", store.events[0].Severity)
	}
}

func TestWordPress_WPJSON(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/wp-json/", nil)

	mod.handleWPJSON(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "wp/v2") {
		t.Error("response body missing wp/v2 namespace")
	}
	if !strings.Contains(body, "wp-site-health/v1") {
		t.Error("response body missing wp-site-health namespace")
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "low" {
		t.Errorf("severity = %q, want low", store.events[0].Severity)
	}
}

func TestWordPress_Static(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/wp-includes/js/jquery.js", nil)

	mod.handleStatic(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if php := rec.Header().Get("X-Powered-By"); php != "PHP/8.1.0" {
		t.Errorf("X-Powered-By = %q, want PHP/8.1.0", php)
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "info" {
		t.Errorf("severity = %q, want info", store.events[0].Severity)
	}
}

func TestWordPress_StaticContent(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/wp-content/uploads/image.png", nil)

	mod.handleStatic(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
}
