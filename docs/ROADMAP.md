# Firewatch Roadmap

## Completed

### Core Infrastructure
- [x] HTTP/HTTPS server with TLS 1.2+, graceful shutdown
- [x] Middleware pipeline (correlation, logging, fingerprint, detection)
- [x] Configuration system with YAML loading, defaults, validation
- [x] SQLite storage with WAL mode (pure Go, no CGO)
- [x] Structured logging via slog (JSON/text)
- [x] Makefile, Dockerfile, .dockerignore

### Honeypot Modules (7/7)
- [x] **nextjs** — 7 routes: RSC probes, server actions, source maps, static assets
- [x] **wordpress** — 8 routes: wp-login brute force, wp-admin, XML-RPC payloads
- [x] **api** — 9 routes: REST enumeration, GraphQL introspection, Swagger/OpenAPI
- [x] **exposure** — 14 routes: .env, .git, config file probes
- [x] **admin** — 16 routes: phpMyAdmin, Adminer, cPanel, generic panels
- [x] **cloud** — 6 routes: AWS/DO metadata, IAM credentials, IMDSv2
- [x] **cve** — 14 routes: Log4Shell, Spring4Shell, MOVEit, PAN-OS, Struts2, Confluence

### Detection & Fingerprinting
- [x] Signature engine — 26 built-in signatures, AND-logic matchers
- [x] Pattern engine — 5 attack patterns, OR-logic rules
- [x] Campaign detection — signature clustering, time-based coordination
- [x] JA3 TLS fingerprinting (requires TLS enabled)
- [x] Header ordering analysis, anomaly detection, known client matching

### Alerting
- [x] Alert manager with concurrent dispatch, per-channel severity thresholds
- [x] Slack (Block Kit), Discord (embeds), generic webhook (JSON + custom headers)
- [x] AlertingStore decorator — transparent alert dispatch on event save

### Threat Intelligence
- [x] IOC extraction from events (IP, URL, user agent, JA3 hash)
- [x] Deduplication and tag merging
- [x] Reverse DNS enrichment
- [x] Intel collection pipeline (extract → enrich → detect campaigns → persist)
- [x] Export: STIX 2.1, MISP event format, CSV

### Deception
- [x] Fake response generators — login pages, API responses, error pages, config files
- [x] Responses for: Next.js, WordPress, phpMyAdmin, Adminer, cPanel, Solr, Spring Boot, MOVEit, Struts2, Confluence
- [x] Fake .env file with realistic credentials

### Utilities (pkg/)
- [x] `crypto` — SHA256, MD5, UUID v4, secure random strings
- [x] `errors` — Error type with Kind/Code, stack traces
- [x] `httputil` — Request/response helpers, header normalization
- [x] `netutil` — IP normalization, CIDR, reverse DNS
- [x] `timeutil` — RFC3339, UTC helpers
- [x] `validate` — IP, URL, severity validators

### Detection Intelligence (2026-02-14)
- [x] IP allowlist/blocklist with CIDR support + file-based lists
- [x] Custom YAML signatures (file + directory loading, merge with built-in)
- [x] Attacker auto-profiling (ProfilingStore decorator, async per-IP tracking)
- [x] Behavioral fingerprinting (scan sweep, brute force, module hopping, progressive recon)
- [x] Campaign auto-correlation (background ticker, signature clustering, coordinated detection)

### Testing
- [x] 18 test suites, 156 tests across all layers
- [x] Mock store pattern for handler tests
- [x] Table-driven tests for signature matching

### Documentation & Notes
- [x] README.md, ARCHITECTURE.md, TREE.md
- [x] 33 learning notes (21 Go, 8 patterns, 4 security)

### GeoIP Enrichment
- [x] MaxMind GeoLite2 integration with graceful degradation
- [x] GeoIP middleware in request pipeline
- [x] Country, city, ASN, org added to event data
- [x] Context propagation via `geoip.WithGeoIP` / `geoip.FromContext`

### Integration Tests
- [x] End-to-end test: start server → send request → verify event in storage
- [x] 14 integration tests in `test/integration/`

### CI Pipeline
- [x] GitHub Actions: lint, test (with coverage), build
- [x] golangci-lint configuration
- [x] Dependabot for Go modules and Actions
- [x] Makefile targets: lint, coverage, docker-build, ci

### CLI Subcommands
- [x] `firewatch events` — query by IP, module, severity, time range
- [x] `firewatch export` — STIX/MISP/CSV export
- [x] `firewatch stats` — event summary, top attackers, top signatures
- [x] Manual subcommand dispatch with `flag`

---

## Future Ideas

Features that would extend the project in meaningful directions. Not committed — just tracked.

### Additional Storage Backends
- PostgreSQL implementation of the Store interface
- Useful for multi-instance deployments or higher query volume

### Email Alerts
- SMTP-based alerter implementation
- `internal/alerts/email.go` — same Alerter interface

### Dashboard
- Web UI for visualizing events, attacker profiles, campaigns
- Could be a separate service reading from the same SQLite/Postgres
- `internal/dashboard/` or standalone frontend

### Kubernetes Deployment
- Deployment and service manifests
- Helm chart for configurable deployments
- `deployments/kubernetes/`

### Honey Tokens & Breadcrumbs
- Embed trackable tokens in deception responses (API keys, URLs)
- Track when tokens appear in subsequent requests (indicates attacker reuse)
- `internal/deception/tokens.go`, `internal/deception/breadcrumbs.go`
