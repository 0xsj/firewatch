package api

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

func newTestModule() (*API, *mockStore) {
	store := &mockStore{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mod := New(config.APIModuleConfig{Enabled: true}, store, logger)
	return mod, store
}

func TestAPI_Name(t *testing.T) {
	mod, _ := newTestModule()
	if mod.Name() != "api" {
		t.Errorf("Name() = %q, want %q", mod.Name(), "api")
	}
}

func TestAPI_Routes(t *testing.T) {
	mod, _ := newTestModule()
	routes := mod.Routes()
	if len(routes) != 9 {
		t.Fatalf("Routes() returned %d routes, want 9", len(routes))
	}

	patterns := make(map[string]bool)
	for _, r := range routes {
		patterns[r.Pattern] = true
	}

	required := []string{
		"GET /api/",
		"POST /api/",
		"GET /graphql",
		"POST /graphql",
		"GET /graphiql",
		"GET /swagger/",
		"GET /swagger.json",
		"GET /openapi.json",
	}
	for _, p := range required {
		if !patterns[p] {
			t.Errorf("missing route pattern: %s", p)
		}
	}
}

func TestAPI_RESTProbe(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)

	mod.handleREST(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	if !strings.Contains(rec.Body.String(), "Unauthorized") {
		t.Error("response body missing Unauthorized")
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "low" {
		t.Errorf("severity = %q, want low", store.events[0].Severity)
	}
}

func TestAPI_RESTAuthProbe(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer fake-token")

	mod.handleREST(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	e := store.events[0]
	if e.Severity != "medium" {
		t.Errorf("severity = %q, want medium", e.Severity)
	}
	if len(e.Signatures) < 2 {
		t.Errorf("signatures = %v, want at least 2", e.Signatures)
	}
}

func TestAPI_RESTApiKeyProbe(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin", nil)
	req.Header.Set("X-Api-Key", "test-key-123")

	mod.handleREST(rec, req)

	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "medium" {
		t.Errorf("severity = %q, want medium", store.events[0].Severity)
	}
}

func TestAPI_GraphQLProbe(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/graphql", nil)

	mod.handleGraphQL(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "Must provide query string") {
		t.Error("response body missing expected error message")
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "medium" {
		t.Errorf("severity = %q, want medium", store.events[0].Severity)
	}
}

func TestAPI_GraphQLIntrospection(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/graphql",
		strings.NewReader(`{"query":"{ __schema { types { name } } }"}`))
	req.Header.Set("Content-Type", "application/json")

	mod.handleGraphQL(rec, req)

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

func TestAPI_GraphQLIntrospectionQuery(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/graphql",
		strings.NewReader(`{"query":"IntrospectionQuery { __schema { queryType { name } } }"}`))
	req.Header.Set("Content-Type", "application/json")

	mod.handleGraphQL(rec, req)

	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "high" {
		t.Errorf("severity = %q, want high", store.events[0].Severity)
	}
}

func TestAPI_GraphQLTypeIntrospection(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/graphql",
		strings.NewReader(`{"query":"{ __type(name: \"User\") { name fields { name } } }"}`))
	req.Header.Set("Content-Type", "application/json")

	mod.handleGraphQL(rec, req)

	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "high" {
		t.Errorf("severity = %q, want high", store.events[0].Severity)
	}
}

func TestAPI_GraphQLNoIntrospection(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/graphql",
		strings.NewReader(`{"query":"{ users { id name } }"}`))
	req.Header.Set("Content-Type", "application/json")

	mod.handleGraphQL(rec, req)

	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "medium" {
		t.Errorf("severity = %q, want medium", store.events[0].Severity)
	}
}

func TestAPI_Swagger(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/swagger/", nil)

	mod.handleSwagger(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "openapi") {
		t.Error("response body missing openapi field")
	}
	if !strings.Contains(body, "Internal API") {
		t.Error("response body missing API title")
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "medium" {
		t.Errorf("severity = %q, want medium", store.events[0].Severity)
	}
}

func TestAPI_SwaggerJSON(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/swagger.json", nil)

	mod.handleSwagger(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
}

func TestAPI_OpenAPIJSON(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)

	mod.handleSwagger(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "/api/v1/users") {
		t.Error("response body missing /api/v1/users endpoint")
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
}
