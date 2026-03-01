package deception

import (
	"strings"
	"testing"
)

func TestNextJSPage(t *testing.T) {
	page := NextJSPage("My App", "abc123")

	if !strings.Contains(page, "<title>My App</title>") {
		t.Error("missing title")
	}
	if !strings.Contains(page, `nonce="abc123"`) {
		t.Error("missing nonce")
	}
	if !strings.Contains(page, `id="__next"`) {
		t.Error("missing __next div")
	}
	if !strings.Contains(page, "_next/static") {
		t.Error("missing Next.js static references")
	}
}

func TestNextJSErrorPage_404(t *testing.T) {
	page := NextJSErrorPage(404)

	if !strings.Contains(page, "404") {
		t.Error("missing 404 status")
	}
	if !strings.Contains(page, "could not be found") {
		t.Error("missing 404 message")
	}
	if !strings.Contains(page, "next-error-h1") {
		t.Error("missing Next.js error styling")
	}
}

func TestNextJSErrorPage_500(t *testing.T) {
	page := NextJSErrorPage(500)

	if !strings.Contains(page, "500") {
		t.Error("missing 500 status")
	}
	if !strings.Contains(page, "Internal Server Error") {
		t.Error("missing 500 message")
	}
}

func TestNextJSRSCPayload(t *testing.T) {
	payload := NextJSRSCPayload()

	if !strings.Contains(payload, "app-pages-browser") {
		t.Error("missing RSC module reference")
	}
	if !strings.Contains(payload, "layout.tsx") {
		t.Error("missing layout reference")
	}
	if !strings.Contains(payload, "page.tsx") {
		t.Error("missing page reference")
	}
}

func TestNextJSServerActionResponse(t *testing.T) {
	resp := NextJSServerActionResponse()

	if !strings.Contains(resp, "actionResult") {
		t.Error("missing actionResult field")
	}
	if !strings.Contains(resp, "$undefined") {
		t.Error("missing $undefined value")
	}
}

func TestNextJSBuildManifest(t *testing.T) {
	manifest := NextJSBuildManifest()

	if !strings.Contains(manifest, "__BUILD_MANIFEST") {
		t.Error("missing BUILD_MANIFEST variable")
	}
	if !strings.Contains(manifest, "polyfillFiles") {
		t.Error("missing polyfillFiles")
	}
	if !strings.Contains(manifest, "webpack.js") {
		t.Error("missing webpack chunk")
	}
}

func TestWordPressLoginPage(t *testing.T) {
	page := WordPressLoginPage("6.4.2")

	if !strings.Contains(page, "WordPress") {
		t.Error("missing WordPress branding")
	}
	if !strings.Contains(page, "6.4.2") {
		t.Error("missing version number")
	}
	if !strings.Contains(page, "wp-login.php") {
		t.Error("missing form action")
	}
	if !strings.Contains(page, `name="log"`) {
		t.Error("missing username field")
	}
	if !strings.Contains(page, `name="pwd"`) {
		t.Error("missing password field")
	}
}

func TestPhpMyAdminLoginPage(t *testing.T) {
	page := PhpMyAdminLoginPage()

	if !strings.Contains(page, "phpMyAdmin") {
		t.Error("missing phpMyAdmin branding")
	}
	if !strings.Contains(page, "5.2.1") {
		t.Error("missing version")
	}
	if !strings.Contains(page, "pma_username") {
		t.Error("missing username field")
	}
	if !strings.Contains(page, "pma_password") {
		t.Error("missing password field")
	}
}

func TestAdminerLoginPage(t *testing.T) {
	page := AdminerLoginPage()

	if !strings.Contains(page, "Adminer") {
		t.Error("missing Adminer branding")
	}
	if !strings.Contains(page, "4.8.1") {
		t.Error("missing version")
	}
	if !strings.Contains(page, "auth[username]") {
		t.Error("missing username field")
	}
	if !strings.Contains(page, "auth[password]") {
		t.Error("missing password field")
	}
}

func TestCPanelLoginPage(t *testing.T) {
	page := CPanelLoginPage()

	if !strings.Contains(page, "cPanel") {
		t.Error("missing cPanel branding")
	}
	if !strings.Contains(page, `name="user"`) {
		t.Error("missing user field")
	}
	if !strings.Contains(page, `name="pass"`) {
		t.Error("missing password field")
	}
}

func TestGenericAdminLoginPage(t *testing.T) {
	page := GenericAdminLoginPage()

	if !strings.Contains(page, "Admin Panel") {
		t.Error("missing Admin Panel heading")
	}
	if !strings.Contains(page, "/admin/login") {
		t.Error("missing form action")
	}
	if !strings.Contains(page, `name="username"`) {
		t.Error("missing username field")
	}
	if !strings.Contains(page, `name="password"`) {
		t.Error("missing password field")
	}
}

func TestSolrAdminPage(t *testing.T) {
	page := SolrAdminPage()

	if !strings.Contains(page, "core0") {
		t.Error("missing core name")
	}
	if !strings.Contains(page, "numDocs") {
		t.Error("missing numDocs field")
	}
	if !strings.Contains(page, "solrconfig.xml") {
		t.Error("missing solrconfig reference")
	}
}

func TestSpringBootHealthJSON(t *testing.T) {
	health := SpringBootHealthJSON()

	if !strings.Contains(health, `"status":"UP"`) {
		t.Error("missing status UP")
	}
	if !strings.Contains(health, "PostgreSQL") {
		t.Error("missing database type")
	}
	if !strings.Contains(health, "diskSpace") {
		t.Error("missing diskSpace component")
	}
}

func TestSpringBootEnvJSON(t *testing.T) {
	env := SpringBootEnvJSON()

	if !strings.Contains(env, "production") {
		t.Error("missing active profile")
	}
	if !strings.Contains(env, "postgresql") {
		t.Error("missing datasource URL")
	}
	if !strings.Contains(env, "******") {
		t.Error("missing masked password")
	}
}

func TestMOVEitLoginPage(t *testing.T) {
	page := MOVEitLoginPage()

	if !strings.Contains(page, "MOVEit Transfer") {
		t.Error("missing MOVEit branding")
	}
	if !strings.Contains(page, "human.aspx") {
		t.Error("missing form action")
	}
	if !strings.Contains(page, "2023.0.1") {
		t.Error("missing version")
	}
}

func TestStrutsShowcasePage(t *testing.T) {
	page := StrutsShowcasePage()

	if !strings.Contains(page, "Struts2 Showcase") {
		t.Error("missing Struts branding")
	}
	if !strings.Contains(page, "2.5.30") {
		t.Error("missing Struts version")
	}
	if !strings.Contains(page, "Tomcat") {
		t.Error("missing Tomcat server info")
	}
}

func TestConfluenceLoginPage(t *testing.T) {
	page := ConfluenceLoginPage()

	if !strings.Contains(page, "Confluence") {
		t.Error("missing Confluence branding")
	}
	if !strings.Contains(page, "os_username") {
		t.Error("missing username field")
	}
	if !strings.Contains(page, "os_password") {
		t.Error("missing password field")
	}
	if !strings.Contains(page, "8.5.3") {
		t.Error("missing version")
	}
}

func TestExposedEnvFile(t *testing.T) {
	env := ExposedEnvFile()

	if !strings.Contains(env, "DB_PASSWORD=") {
		t.Error("missing DB_PASSWORD")
	}
	if !strings.Contains(env, "AWS_ACCESS_KEY_ID=") {
		t.Error("missing AWS_ACCESS_KEY_ID")
	}
	if !strings.Contains(env, "AWS_SECRET_ACCESS_KEY=") {
		t.Error("missing AWS_SECRET_ACCESS_KEY")
	}
	if !strings.Contains(env, "STRIPE_KEY=") {
		t.Error("missing STRIPE_KEY")
	}
	if !strings.Contains(env, "production") {
		t.Error("missing production environment")
	}
}
