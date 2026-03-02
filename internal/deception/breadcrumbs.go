package deception

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"strings"
)

// BreadcrumbConfig controls breadcrumb injection.
type BreadcrumbConfig struct {
	Domain         string
	EnabledModules []string
}

// htmlBreadcrumbs maps module names to HTML snippets that lure
// attackers toward those endpoints.
var htmlBreadcrumbs = map[string][]string{
	"admin": {
		`<!-- TODO: remove /admin/backup before deploy -->`,
		`<a href="/phpmyadmin/" style="display:none">db</a>`,
		`<!-- admin panel: /admin/ (remember to disable) -->`,
	},
	"wordpress": {
		`<!-- wp-login redirect: /wp-login.php -->`,
		`<a href="/wp-admin/" style="display:none">cms</a>`,
		`<!-- WordPress staging at /wp-json/ -->`,
	},
	"exposure": {
		`<!-- config backup: /.env.backup -->`,
		`<a href="/.git/config" style="display:none">.</a>`,
		`<!-- debug: check /.env for DB creds -->`,
	},
	"cloud": {
		`<!-- SSRF test: /latest/meta-data/ -->`,
		`<a href="/latest/meta-data/iam/" style="display:none">m</a>`,
	},
	"api": {
		`<!-- internal API docs: /swagger/ -->`,
		`<a href="/api/v1/users" style="display:none">api</a>`,
		`<!-- openapi spec: /openapi.json -->`,
	},
	"cve": {
		`<!-- Solr admin: /solr/admin/cores -->`,
		`<a href="/actuator/env" style="display:none">.</a>`,
		`<!-- Confluence: /wiki/ -->`,
	},
	"nextjs": {
		`<!-- debug endpoint: /__nextjs_original-stack-frame -->`,
		`<a href="/_next/data/" style="display:none">.</a>`,
	},
}

// envBreadcrumbs maps module names to env var hints.
var envBreadcrumbs = map[string][]string{
	"admin": {
		"ADMIN_URL=http://%s/admin/",
		"PMA_URL=http://%s/phpmyadmin/",
	},
	"api": {
		"INTERNAL_API=http://%s/api/v1/",
		"SWAGGER_URL=http://%s/swagger/",
	},
	"cloud": {
		"METADATA_ENDPOINT=http://%s/latest/meta-data/",
	},
	"wordpress": {
		"WP_HOME=http://%s/",
		"WP_SITEURL=http://%s/wp-admin/",
	},
	"cve": {
		"SOLR_URL=http://%s/solr/admin/cores",
		"BACKUP_DB_HOST=http://%s/actuator/env",
	},
}

// headerBreadcrumbs maps module names to response header hints.
var headerBreadcrumbs = map[string][]string{
	"admin": {
		"X-Debug-Endpoint: /admin/",
	},
	"api": {
		"X-Debug-Endpoint: /api/v1/health",
	},
	"cloud": {
		"X-Debug-Endpoint: /latest/meta-data/",
	},
}

// InjectHTML appends hidden breadcrumbs before </body> in the given
// HTML string. Only references modules other than currentModule.
func InjectHTML(html string, currentModule string, cfg BreadcrumbConfig) string {
	candidates := collectBreadcrumbs(htmlBreadcrumbs, currentModule, cfg.EnabledModules)
	if len(candidates) == 0 {
		return html
	}

	// Pick 2-3 breadcrumbs
	count := 2
	if len(candidates) > 2 {
		n, _ := rand.Int(rand.Reader, big.NewInt(2))
		count = 2 + int(n.Int64()) // 2 or 3
	}
	if count > len(candidates) {
		count = len(candidates)
	}

	selected := pickRandom(candidates, count)
	injection := "\n" + strings.Join(selected, "\n") + "\n"

	if idx := strings.LastIndex(html, "</body>"); idx != -1 {
		return html[:idx] + injection + html[idx:]
	}
	return html + injection
}

// InjectEnv appends env var breadcrumbs pointing to other enabled
// honeypot modules.
func InjectEnv(env string, currentModule string, cfg BreadcrumbConfig) string {
	candidates := collectEnvBreadcrumbs(currentModule, cfg)
	if len(candidates) == 0 {
		return env
	}

	count := 2
	if count > len(candidates) {
		count = len(candidates)
	}
	selected := pickRandom(candidates, count)

	result := strings.TrimRight(env, "\n")
	result += "\n\n# Internal endpoints\n"
	result += strings.Join(selected, "\n") + "\n"
	return result
}

// BreadcrumbHeaders sets response headers that hint at other
// honeypot endpoints.
func BreadcrumbHeaders(w http.ResponseWriter, currentModule string, cfg BreadcrumbConfig) {
	candidates := collectBreadcrumbs(headerBreadcrumbs, currentModule, cfg.EnabledModules)
	if len(candidates) == 0 {
		return
	}

	selected := pickRandom(candidates, 1)
	for _, h := range selected {
		parts := strings.SplitN(h, ": ", 2)
		if len(parts) == 2 {
			w.Header().Set(parts[0], parts[1])
		}
	}
	w.Header().Set("X-Powered-By", "nginx/1.24.0")
}

// collectBreadcrumbs gathers breadcrumb strings from the pool for
// modules other than currentModule that are in enabledModules.
func collectBreadcrumbs(pool map[string][]string, currentModule string, enabledModules []string) []string {
	enabled := make(map[string]bool)
	for _, m := range enabledModules {
		enabled[m] = true
	}

	var candidates []string
	for mod, crumbs := range pool {
		if mod == currentModule || !enabled[mod] {
			continue
		}
		candidates = append(candidates, crumbs...)
	}
	return candidates
}

// collectEnvBreadcrumbs gathers env var lines for modules other
// than currentModule, substituting the domain.
func collectEnvBreadcrumbs(currentModule string, cfg BreadcrumbConfig) []string {
	enabled := make(map[string]bool)
	for _, m := range cfg.EnabledModules {
		enabled[m] = true
	}

	domain := cfg.Domain
	if domain == "" {
		domain = "localhost"
	}

	var candidates []string
	for mod, templates := range envBreadcrumbs {
		if mod == currentModule || !enabled[mod] {
			continue
		}
		for _, tmpl := range templates {
			candidates = append(candidates, fmt.Sprintf(tmpl, domain))
		}
	}
	return candidates
}

// pickRandom selects n unique random items from candidates.
func pickRandom(candidates []string, n int) []string {
	if n >= len(candidates) {
		return candidates
	}

	// Fisher-Yates partial shuffle
	picked := make([]string, len(candidates))
	copy(picked, candidates)
	for i := 0; i < n; i++ {
		j, _ := rand.Int(rand.Reader, big.NewInt(int64(len(picked)-i)))
		idx := i + int(j.Int64())
		picked[i], picked[idx] = picked[idx], picked[i]
	}
	return picked[:n]
}
