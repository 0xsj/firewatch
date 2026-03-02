package cloud

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

func newTestModule() (*Cloud, *mockStore) {
	store := &mockStore{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	deception := config.DeceptionConfig{HoneyTokens: true, Breadcrumbs: true, FakeErrors: true}
	mod := New(config.CloudModuleConfig{Enabled: true}, deception, store, logger)
	return mod, store
}

func newTestModuleNoDeception() (*Cloud, *mockStore) {
	store := &mockStore{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	deception := config.DeceptionConfig{HoneyTokens: false, Breadcrumbs: false, FakeErrors: false}
	mod := New(config.CloudModuleConfig{Enabled: true}, deception, store, logger)
	return mod, store
}

func TestCloud_Name(t *testing.T) {
	mod, _ := newTestModule()
	if mod.Name() != "cloud" {
		t.Errorf("Name() = %q, want %q", mod.Name(), "cloud")
	}
}

func TestCloud_Routes(t *testing.T) {
	mod, _ := newTestModule()
	routes := mod.Routes()
	if len(routes) != 6 {
		t.Fatalf("Routes() returned %d routes, want 6", len(routes))
	}

	patterns := make(map[string]bool)
	for _, r := range routes {
		patterns[r.Pattern] = true
	}

	required := []string{
		"GET /latest/meta-data/",
		"GET /latest/meta-data/iam/",
		"GET /latest/meta-data/iam/security-credentials/",
		"GET /latest/user-data",
		"GET /metadata/v1/",
		"PUT /latest/api/token",
	}
	for _, p := range required {
		if !patterns[p] {
			t.Errorf("missing route pattern: %s", p)
		}
	}
}

func TestCloud_Metadata(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/latest/meta-data/", nil)

	mod.handleMetadata(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/plain" {
		t.Errorf("Content-Type = %q, want text/plain", ct)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "ami-id") {
		t.Error("response body missing ami-id")
	}
	if !strings.Contains(body, "instance-id") {
		t.Error("response body missing instance-id")
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "critical" {
		t.Errorf("severity = %q, want critical", store.events[0].Severity)
	}
}

func TestCloud_MetadataUserData(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/latest/user-data", nil)

	mod.handleMetadata(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "critical" {
		t.Errorf("severity = %q, want critical", store.events[0].Severity)
	}
}

func TestCloud_MetadataDO(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metadata/v1/", nil)

	mod.handleMetadata(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "critical" {
		t.Errorf("severity = %q, want critical", store.events[0].Severity)
	}
}

func TestCloud_IAM(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/latest/meta-data/iam/security-credentials/", nil)

	mod.handleIAM(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "AKIA") {
		t.Error("response body missing AKIA-prefixed access key")
	}
	if !strings.Contains(body, "AccessKeyId") {
		t.Error("response body missing AccessKeyId field")
	}
	if !strings.Contains(body, "SecretAccessKey") {
		t.Error("response body missing SecretAccessKey field")
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "critical" {
		t.Errorf("severity = %q, want critical", store.events[0].Severity)
	}
	// Verify honey tokens were saved (access key, secret key, session token)
	if len(store.honeyTokens) != 3 {
		t.Errorf("honeyTokens = %d, want 3", len(store.honeyTokens))
	}
}

func TestCloud_IAMRoot(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/latest/meta-data/iam/", nil)

	mod.handleIAM(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "critical" {
		t.Errorf("severity = %q, want critical", store.events[0].Severity)
	}
}

func TestCloud_IMDSv2(t *testing.T) {
	mod, store := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/latest/api/token", nil)
	req.Header.Set("X-Aws-Ec2-Metadata-Token-Ttl-Seconds", "21600")

	mod.handleIMDSv2(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "AQAAANjCpMCZjg_") {
		t.Error("response body missing IMDS token prefix")
	}
	ttl := rec.Header().Get("X-Aws-Ec2-Metadata-Token-Ttl-Seconds")
	if ttl != "21600" {
		t.Errorf("TTL header = %q, want 21600", ttl)
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	if store.events[0].Severity != "critical" {
		t.Errorf("severity = %q, want critical", store.events[0].Severity)
	}
}

func TestCloud_IAM_UniquePerRequest(t *testing.T) {
	mod, store := newTestModule()

	// First request.
	rec1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/latest/meta-data/iam/security-credentials/", nil)
	mod.handleIAM(rec1, req1)

	// Second request.
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/latest/meta-data/iam/security-credentials/", nil)
	mod.handleIAM(rec2, req2)

	if rec1.Body.String() == rec2.Body.String() {
		t.Error("two IAM requests returned identical credentials; expected unique per request")
	}
	// Should have 6 tokens total (3 per request).
	if len(store.honeyTokens) != 6 {
		t.Errorf("honeyTokens = %d, want 6", len(store.honeyTokens))
	}
}

func TestCloud_IAM_StaticWhenDisabled(t *testing.T) {
	mod, store := newTestModuleNoDeception()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/latest/meta-data/iam/security-credentials/", nil)

	mod.handleIAM(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "AKIAIOSFODNN7EXAMPLE") {
		t.Error("expected static access key AKIAIOSFODNN7EXAMPLE when HoneyTokens disabled")
	}
	if !strings.Contains(body, "wJalrXUtnFEMI") {
		t.Error("expected static secret key when HoneyTokens disabled")
	}
	// No honey tokens should be saved.
	if len(store.honeyTokens) != 0 {
		t.Errorf("honeyTokens = %d, want 0 when disabled", len(store.honeyTokens))
	}
}

func TestCloud_IMDSv2_StaticWhenDisabled(t *testing.T) {
	mod, _ := newTestModuleNoDeception()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/latest/api/token", nil)
	req.Header.Set("X-Aws-Ec2-Metadata-Token-Ttl-Seconds", "21600")

	mod.handleIMDSv2(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "fake_imds_token_do_not_use") {
		t.Error("expected static IMDS token when HoneyTokens disabled")
	}
}

func TestCloud_Metadata_BreadcrumbHeaders(t *testing.T) {
	mod, _ := newTestModule()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/latest/meta-data/", nil)

	mod.handleMetadata(rec, req)

	// Breadcrumbs enabled: should have X-Powered-By set by BreadcrumbHeaders.
	if xpb := rec.Header().Get("X-Powered-By"); xpb == "" {
		t.Error("expected X-Powered-By header from breadcrumbs")
	}
}

func TestCloud_Metadata_NoBreadcrumbsWhenDisabled(t *testing.T) {
	mod, _ := newTestModuleNoDeception()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/latest/meta-data/", nil)

	mod.handleMetadata(rec, req)

	if xde := rec.Header().Get("X-Debug-Endpoint"); xde != "" {
		t.Errorf("unexpected X-Debug-Endpoint header when breadcrumbs disabled: %q", xde)
	}
}
