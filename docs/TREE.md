# Project Structure

## Directory Layout

```
firewatch/
├── cmd/
│   └── firewatch/
│       └── main.go                          # CLI entry point, module wiring
│
├── internal/
│   ├── config/
│   │   ├── config.go                        # Config types, Load, LoadOrDefault
│   │   ├── defaults.go                      # Default configuration values
│   │   └── validation.go                    # Config validation rules
│   │
│   ├── server/
│   │   ├── server.go                        # HTTP server, middleware assembly
│   │   ├── router.go                        # ServeMux wrapper, route registration
│   │   ├── tls.go                           # TLS 1.2+ configuration
│   │   └── graceful.go                      # Signal handling, ListenAndShutdown
│   │
│   ├── middleware/
│   │   ├── middleware.go                     # Middleware type, Chain(), responseWriter
│   │   ├── correlation.go                   # Request ID generation, context storage
│   │   ├── ipfilter.go                      # IP allowlist/blocklist filtering (CIDR)
│   │   ├── ipfilter_test.go                 # IP filter tests (8 tests)
│   │   ├── ratelimit.go                     # Per-IP token bucket rate limiting
│   │   ├── ratelimit_test.go                # Rate limit tests (7 tests)
│   │   ├── logging.go                       # Structured request logging (slog)
│   │   ├── geoip.go                         # GeoIP MaxMind lookup middleware
│   │   ├── fingerprint.go                   # Fingerprint engine middleware
│   │   ├── detection.go                     # Detection engine middleware
│   │   ├── detection_test.go                # Detection middleware tests
│   │   ├── behavior.go                      # Behavioral fingerprinting middleware
│   │   └── behavior_test.go                 # Behavior middleware tests (5 tests)
│   │
│   ├── handlers/
│   │   ├── handler.go                       # Module interface, Route type
│   │   ├── registry.go                      # Module registry with lookup
│   │   ├── event.go                         # Shared RecordEvent helper
│   │   │
│   │   ├── nextjs/
│   │   │   ├── nextjs.go                    # Module entry, 7 routes
│   │   │   ├── server_action.go             # Next-Action header detection
│   │   │   ├── rsc.go                       # React Server Component probes
│   │   │   ├── static.go                    # Static asset enumeration
│   │   │   ├── event.go                     # Module-specific event recording
│   │   │   └── nextjs_test.go               # Handler tests
│   │   │
│   │   ├── wordpress/
│   │   │   ├── wordpress.go                 # Module entry, 8 routes
│   │   │   ├── login.go                     # wp-login GET/POST, brute force
│   │   │   ├── admin.go                     # wp-admin, wp-json, static assets
│   │   │   └── xmlrpc.go                    # XML-RPC probe/payload detection
│   │   │
│   │   ├── api/
│   │   │   ├── api.go                       # Module entry, 9 routes
│   │   │   ├── rest.go                      # REST API probes, auth detection
│   │   │   ├── graphql.go                   # GraphQL probes, introspection
│   │   │   └── swagger.go                   # Swagger/OpenAPI fake spec
│   │   │
│   │   ├── exposure/
│   │   │   ├── exposure.go                  # Module entry, 14 routes
│   │   │   ├── env.go                       # .env file probes
│   │   │   ├── git.go                       # .git/config, .git/HEAD probes
│   │   │   └── config.go                    # Config file probes (403)
│   │   │
│   │   ├── admin/
│   │   │   ├── admin.go                     # Module entry, 16 routes
│   │   │   ├── phpmyadmin.go                # phpMyAdmin GET/POST
│   │   │   ├── adminer.go                   # Adminer GET/POST
│   │   │   ├── cpanel.go                    # cPanel GET/POST
│   │   │   ├── generic.go                   # Generic admin panel GET/POST
│   │   │   └── admin_test.go                # Handler tests
│   │   │
│   │   ├── cloud/
│   │   │   ├── cloud.go                     # Module entry, 6 routes
│   │   │   └── metadata.go                  # AWS/DO metadata, IAM, IMDSv2
│   │   │
│   │   └── cve/
│   │       ├── cve.go                       # Module entry, 14 routes, CVE filtering
│   │       ├── log4shell.go                 # CVE-2021-44228: Solr, JNDI detection
│   │       ├── spring4shell.go              # CVE-2022-22965: Actuator health/env
│   │       ├── moveit.go                    # CVE-2023-34362: MOVEit Transfer
│   │       ├── panos.go                     # CVE-2024-3400: GlobalProtect, HIP
│   │       ├── struts2.go                   # CVE-2017-5638: OGNL injection
│   │       ├── confluence.go                # CVE-2023-22515: Admin creation
│   │       └── cve_test.go                  # 18 handler tests
│   │
│   ├── detection/
│   │   ├── detector.go                      # Detector engine, field extraction
│   │   ├── signatures.go                    # Signature type, 26 built-in sigs
│   │   ├── patterns.go                      # Pattern type, 5 built-in patterns
│   │   ├── campaign.go                      # Campaign clustering + coordination detection
│   │   ├── correlator.go                    # Background campaign auto-correlator
│   │   ├── correlator_test.go               # Correlator tests (7 tests)
│   │   ├── behavior.go                      # Per-IP behavioral fingerprinting tracker
│   │   ├── behavior_test.go                 # Behavior tracker tests (9 tests)
│   │   ├── loader.go                        # YAML signature/pattern file loader
│   │   ├── loader_test.go                   # Loader tests (8 tests)
│   │   ├── detector_test.go                 # Detector tests
│   │   └── signatures_test.go               # Signature matching tests
│   │
│   ├── fingerprint/
│   │   ├── fingerprint.go                   # Engine, Result, context helpers
│   │   ├── ja3.go                           # JA3 TLS fingerprinting
│   │   ├── headers.go                       # Header ordering, anomaly detection
│   │   └── fingerprint_test.go              # Engine tests
│   │
│   ├── alerts/
│   │   ├── alerter.go                       # Alerter interface, Alert type
│   │   ├── manager.go                       # Concurrent dispatch, severity gating
│   │   ├── slack.go                         # Slack Block Kit webhooks
│   │   ├── discord.go                       # Discord embed webhooks
│   │   ├── webhook.go                       # Generic JSON webhooks
│   │   └── alerts_test.go                   # Alert dispatch tests
│   │
│   ├── intel/
│   │   ├── collector.go                     # Extract → enrich → detect → persist
│   │   ├── ioc/
│   │   │   ├── extractor.go                 # IOC extraction, dedup, tag merging
│   │   │   └── extractor_test.go            # Extraction tests
│   │   ├── enrichment/
│   │   │   ├── enricher.go                  # Enricher interface
│   │   │   ├── dns.go                       # Reverse DNS enrichment
│   │   │   └── geoip.go                     # GeoIP enrichment (placeholder)
│   │   └── export/
│   │       ├── exporter.go                  # Exporter interface
│   │       ├── stix.go                      # STIX 2.1 bundle export
│   │       ├── misp.go                      # MISP event format export
│   │       ├── csv.go                       # CSV export
│   │       └── export_test.go               # Export format tests
│   │
│   ├── deception/
│   │   └── responses.go                     # Fake HTML/JSON response generators
│   │
│   └── storage/
│       ├── storage.go                       # Store interface, filter types
│       ├── sqlite.go                        # SQLite implementation (WAL)
│       ├── profiling.go                     # ProfilingStore — auto attacker profiling
│       ├── profiling_test.go                # Profiling tests (7 tests)
│       ├── alerting.go                      # AlertingStore decorator
│       ├── alerting_test.go                 # AlertingStore tests
│       ├── sqlite_test.go                   # SQLite CRUD tests
│       └── models/
│           ├── event.go                     # Event, Fingerprint, GeoIPInfo
│           ├── attacker.go                  # Attacker profile model
│           ├── campaign.go                  # Campaign model
│           └── ioc.go                       # IOC model
│
├── pkg/
│   ├── crypto/
│   │   ├── hash.go                          # SHA256, MD5 hashing
│   │   ├── random.go                        # UUID v4, secure random strings
│   │   ├── hash_test.go                     # Hash tests
│   │   └── random_test.go                   # Random generation tests
│   ├── errors/
│   │   ├── errors.go                        # Core Error type, constructors
│   │   ├── kinds.go                         # Kind type (broad categories)
│   │   ├── codes.go                         # Code type (specific identifiers)
│   │   └── stack.go                         # Stack trace capture
│   ├── httputil/
│   │   ├── request.go                       # ClientIP, body reading helpers
│   │   ├── response.go                      # JSON response writers
│   │   ├── headers.go                       # Header normalization, ordering
│   │   └── request_test.go                  # Request helper tests
│   ├── netutil/
│   │   ├── ip.go                            # IP normalization, CIDR matching
│   │   └── dns.go                           # Reverse DNS helpers
│   ├── timeutil/
│   │   └── time.go                          # RFC3339 formatting, UTC helpers
│   └── validate/
│       └── validate.go                      # IP, URL, severity validators
│
├── notes/                                   # Learning notes (Obsidian-compatible)
│   ├── go/                                  # 21 Go concept notes
│   ├── patterns/                            # 9 design pattern notes
│   └── security/                            # 4 security domain notes
│
├── docs/
│   ├── ARCHITECTURE.md                      # System design and diagrams
│   └── TREE.md                              # This file
│
├── CLAUDE.md                                # AI development instructions
├── README.md                                # Project overview and usage
├── Makefile                                 # fmt, vet, test, build targets
├── Dockerfile                               # Multi-stage Go build
├── .dockerignore                            # Docker build exclusions
├── .gitignore                               # Git exclusions
├── firewatch.yaml                           # Default configuration
├── go.mod                                   # Go module definition
└── go.sum                                   # Dependency checksums
```

## Build Phases

### Phase 1: Foundation (zero internal dependencies)

- `pkg/errors/` — Error types, kinds, codes, stack traces
- `pkg/crypto/` — SHA256, MD5, UUID v4, random tokens
- `pkg/httputil/` — Request/response helpers, header utilities
- `pkg/netutil/` — IP normalization, CIDR, reverse DNS
- `pkg/timeutil/` — RFC3339 formatting, UTC
- `pkg/validate/` — Input validators

### Phase 2: Domain Models and Storage

- `internal/storage/models/` — Event, Attacker, Campaign, IOC
- `internal/storage/storage.go` — Store interface, filter types
- `internal/storage/sqlite.go` — SQLite implementation
- `internal/config/` — Configuration loading, defaults, validation

### Phase 3: Server Core

- `internal/middleware/middleware.go` — Middleware type, Chain
- `internal/middleware/correlation.go` — Request ID
- `internal/middleware/logging.go` — Request logging
- `internal/server/` — HTTP server, router, TLS, graceful shutdown

### Phase 4: Fingerprinting

- `internal/fingerprint/ja3.go` — JA3 TLS fingerprinting
- `internal/fingerprint/headers.go` — Header analysis
- `internal/fingerprint/fingerprint.go` — Engine, context storage
- `internal/middleware/fingerprint.go` — Middleware wiring

### Phase 5: Detection

- `internal/detection/signatures.go` — Signature type, matchers, 26 signatures
- `internal/detection/patterns.go` — Pattern type, 5 patterns
- `internal/detection/detector.go` — Detection engine
- `internal/detection/campaign.go` — Campaign detection
- `internal/middleware/detection.go` — Middleware wiring

### Phase 6: Alerting

- `internal/alerts/alerter.go` — Interface, Alert type
- `internal/alerts/manager.go` — Concurrent dispatch
- `internal/alerts/slack.go` — Slack webhooks
- `internal/alerts/discord.go` — Discord webhooks
- `internal/alerts/webhook.go` — Generic webhooks
- `internal/storage/alerting.go` — AlertingStore decorator

### Phase 7: Honeypot Modules

- `internal/handlers/handler.go` — Module interface
- `internal/handlers/event.go` — Shared RecordEvent
- `internal/deception/responses.go` — Fake responses
- `internal/handlers/nextjs/` — First module (7 routes)
- `internal/handlers/wordpress/` — 8 routes
- `internal/handlers/exposure/` — 14 routes
- `internal/handlers/api/` — 9 routes
- `internal/handlers/cloud/` — 6 routes
- `internal/handlers/admin/` — 16 routes
- `internal/handlers/cve/` — 14 routes

### Phase 8: Threat Intelligence

- `internal/intel/ioc/extractor.go` — IOC extraction
- `internal/intel/enrichment/` — Enricher interface, DNS, GeoIP
- `internal/intel/collector.go` — Pipeline orchestrator
- `internal/intel/export/` — STIX, MISP, CSV exporters

### Phase 9: Entry Point

- `cmd/firewatch/main.go` — CLI, dependency wiring, module registration

## Test Suites

| Package                     | File                  | Tests | Coverage Area                          |
|-----------------------------|-----------------------|-------|----------------------------------------|
| `pkg/crypto`                | `hash_test.go`        | 4     | SHA256, MD5 hashing                    |
| `pkg/crypto`                | `random_test.go`      | 3     | UUID, random string generation         |
| `pkg/httputil`              | `request_test.go`     | 3     | ClientIP, header extraction            |
| `internal/fingerprint`      | `fingerprint_test.go` | 3     | Engine, header analysis                |
| `internal/detection`        | `detector_test.go`    | 4     | Signature/pattern evaluation           |
| `internal/detection`        | `signatures_test.go`  | 5     | Matcher logic, built-in sigs           |
| `internal/detection`        | `loader_test.go`      | 8     | Custom YAML signature loading          |
| `internal/detection`        | `behavior_test.go`    | 9     | Behavioral fingerprinting tracker      |
| `internal/detection`        | `correlator_test.go`  | 7     | Campaign auto-correlation              |
| `internal/middleware`       | `detection_test.go`   | 3     | Detection middleware integration       |
| `internal/middleware`       | `ipfilter_test.go`    | 8     | IP allowlist/blocklist filtering       |
| `internal/middleware`       | `ratelimit_test.go`   | 7     | Per-IP token bucket rate limiting      |
| `internal/middleware`       | `behavior_test.go`    | 5     | Behavioral fingerprinting middleware   |
| `internal/handlers/nextjs`  | `nextjs_test.go`      | 6     | Next.js handler responses/events       |
| `internal/handlers/admin`   | `admin_test.go`       | 8     | Admin panel handler responses/events   |
| `internal/handlers/cve`     | `cve_test.go`         | 18    | All CVE handlers, route filtering      |
| `internal/alerts`           | `alerts_test.go`      | 3     | Alert dispatch, severity gating        |
| `internal/storage`          | `sqlite_test.go`      | 5     | SQLite CRUD operations                 |
| `internal/storage`          | `alerting_test.go`    | 3     | AlertingStore decorator                |
| `internal/storage`          | `profiling_test.go`   | 7     | Attacker auto-profiling                |
| `internal/intel/ioc`        | `extractor_test.go`   | 4     | IOC extraction, dedup                  |
| `internal/intel/export`     | `export_test.go`      | 4     | STIX, MISP, CSV formatting             |
