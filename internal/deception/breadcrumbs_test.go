package deception

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInjectHTML_AddsContent(t *testing.T) {
	html := `<html><body><h1>Login</h1></body></html>`
	cfg := BreadcrumbConfig{
		Domain:         "example.com",
		EnabledModules: []string{"admin", "api", "cloud", "wordpress"},
	}

	result := InjectHTML(html, "cve", cfg)

	if result == html {
		t.Error("InjectHTML did not modify the HTML")
	}
	if !strings.Contains(result, "</body>") {
		t.Error("InjectHTML removed </body> tag")
	}
	// Should contain at least one breadcrumb marker (comment or hidden link)
	if !strings.Contains(result, "<!--") && !strings.Contains(result, "display:none") {
		t.Error("InjectHTML did not inject any breadcrumb content")
	}
}

func TestInjectHTML_NoSelfReference(t *testing.T) {
	html := `<html><body><h1>Admin</h1></body></html>`
	cfg := BreadcrumbConfig{
		Domain:         "example.com",
		EnabledModules: []string{"admin", "api"},
	}

	// Inject as "admin" — should only get "api" breadcrumbs
	result := InjectHTML(html, "admin", cfg)

	// Should not contain admin-specific URLs (phpmyadmin, /admin/)
	// injected as breadcrumbs
	for _, crumb := range htmlBreadcrumbs["admin"] {
		if strings.Contains(result, crumb) {
			t.Errorf("InjectHTML included self-referencing breadcrumb: %s", crumb)
		}
	}
}

func TestInjectHTML_NoOtherModulesEnabled(t *testing.T) {
	html := `<html><body><h1>Login</h1></body></html>`
	cfg := BreadcrumbConfig{
		Domain:         "example.com",
		EnabledModules: []string{"admin"}, // only self
	}

	result := InjectHTML(html, "admin", cfg)

	if result != html {
		t.Errorf("InjectHTML modified HTML when no other modules are enabled: got %q", result)
	}
}

func TestInjectHTML_EmptyModules(t *testing.T) {
	html := `<html><body></body></html>`
	cfg := BreadcrumbConfig{
		Domain:         "example.com",
		EnabledModules: []string{},
	}

	result := InjectHTML(html, "admin", cfg)

	if result != html {
		t.Error("InjectHTML modified HTML with no enabled modules")
	}
}

func TestInjectEnv_AppendsLines(t *testing.T) {
	env := "DB_HOST=localhost\nDB_USER=root\n"
	cfg := BreadcrumbConfig{
		Domain:         "app.example.com",
		EnabledModules: []string{"exposure", "admin", "api"},
	}

	result := InjectEnv(env, "exposure", cfg)

	if result == env {
		t.Error("InjectEnv did not modify the env")
	}
	if !strings.Contains(result, "DB_HOST=localhost") {
		t.Error("InjectEnv removed original content")
	}
	if !strings.Contains(result, "app.example.com") {
		t.Error("InjectEnv did not substitute domain")
	}
}

func TestInjectEnv_NoSelfReference(t *testing.T) {
	env := "APP=test\n"
	cfg := BreadcrumbConfig{
		Domain:         "localhost",
		EnabledModules: []string{"admin", "api"},
	}

	result := InjectEnv(env, "admin", cfg)

	// admin env breadcrumbs should not appear
	if strings.Contains(result, "ADMIN_URL") || strings.Contains(result, "PMA_URL") {
		t.Error("InjectEnv included self-referencing breadcrumbs")
	}
}

func TestBreadcrumbHeaders_SetsHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	cfg := BreadcrumbConfig{
		Domain:         "localhost",
		EnabledModules: []string{"cloud", "admin", "api"},
	}

	BreadcrumbHeaders(rec, "wordpress", cfg)

	if rec.Header().Get("X-Powered-By") == "" {
		t.Error("BreadcrumbHeaders did not set X-Powered-By")
	}
}

func TestBreadcrumbHeaders_NoModules(t *testing.T) {
	rec := httptest.NewRecorder()
	cfg := BreadcrumbConfig{
		Domain:         "localhost",
		EnabledModules: []string{},
	}

	BreadcrumbHeaders(rec, "admin", cfg)

	if rec.Header().Get("X-Debug-Endpoint") != "" {
		t.Error("BreadcrumbHeaders set header with no enabled modules")
	}
}
