# Firewatch Roadmap

## Project Tree

```
firewatch/
├── cmd/
│   └── firewatch/
│       └── main.go                     # CLI entry point
│
├── internal/
│   ├── config/
│   │   ├── config.go                   # Configuration loading
│   │   ├── defaults.go                 # Default values
│   │   └── validation.go               # Config validation
│   │
│   ├── server/
│   │   ├── server.go                   # HTTP/HTTPS server
│   │   ├── router.go                   # Route registration
│   │   ├── tls.go                      # TLS configuration
│   │   └── graceful.go                 # Graceful shutdown
│   │
│   ├── middleware/
│   │   ├── middleware.go               # Middleware chain
│   │   ├── logging.go                  # Request logging
│   │   ├── fingerprint.go              # Request fingerprinting
│   │   ├── geoip.go                    # GeoIP enrichment
│   │   ├── ratelimit.go                # Rate limiting
│   │   └── correlation.go              # Request correlation
│   │
│   ├── handlers/
│   │   ├── handler.go                  # Handler interface
│   │   ├── registry.go                 # Handler registry
│   │   ├── event.go                    # Shared event recording helper
│   │   ├── nextjs/
│   │   │   ├── nextjs.go               # Next.js module
│   │   │   ├── server_action.go        # next-action honeypot
│   │   │   ├── rsc.go                  # RSC endpoint honeypot
│   │   │   └── static.go               # _next/static honeypot
│   │   ├── wordpress/
│   │   │   ├── wordpress.go            # WordPress module
│   │   │   ├── login.go                # wp-login honeypot
│   │   │   ├── admin.go                # wp-admin honeypot
│   │   │   └── xmlrpc.go               # xmlrpc honeypot
│   │   ├── api/
│   │   │   ├── api.go                  # API module
│   │   │   ├── rest.go                 # REST honeypot
│   │   │   ├── graphql.go              # GraphQL honeypot
│   │   │   └── swagger.go              # Swagger honeypot
│   │   ├── exposure/
│   │   │   ├── exposure.go             # Exposure module
│   │   │   ├── env.go                  # .env honeypot
│   │   │   ├── git.go                  # .git honeypot
│   │   │   └── config.go               # Config file honeypot
│   │   ├── admin/
│   │   │   ├── admin.go                # Admin module
│   │   │   └── panels.go               # Admin panel honeypots
│   │   ├── cloud/
│   │   │   ├── cloud.go                # Cloud module
│   │   │   └── metadata.go             # Metadata endpoint honeypot
│   │   └── cve/
│   │       ├── cve.go                  # CVE module
│   │       ├── log4j.go                # Log4Shell honeypot
│   │       └── spring4shell.go         # Spring4Shell honeypot
│   │
│   ├── fingerprint/
│   │   ├── fingerprint.go              # Fingerprinting engine
│   │   ├── ja3.go                      # JA3 fingerprinting
│   │   ├── ja4.go                      # JA4 fingerprinting
│   │   ├── headers.go                  # Header analysis
│   │   └── behavioral.go               # Behavioral fingerprinting
│   │
│   ├── detection/
│   │   ├── detector.go                 # Detection engine
│   │   ├── signatures.go               # Scanner signatures
│   │   ├── patterns.go                 # Attack patterns
│   │   └── campaign.go                 # Campaign detection
│   │
│   ├── intel/
│   │   ├── collector.go                # Intel collection orchestrator
│   │   ├── enrichment/
│   │   │   ├── enricher.go             # Enrichment interface
│   │   │   ├── geoip.go                # GeoIP enrichment (placeholder)
│   │   │   └── dns.go                  # Reverse DNS enrichment
│   │   ├── ioc/
│   │   │   └── extractor.go            # IOC extraction & deduplication
│   │   └── export/
│   │       ├── exporter.go             # Export interface
│   │       ├── stix.go                 # STIX 2.1 bundle export
│   │       ├── misp.go                 # MISP event format export
│   │       └── csv.go                  # CSV export
│   │
│   ├── deception/
│   │   ├── tokens.go                   # Honey token generation
│   │   ├── breadcrumbs.go              # Breadcrumb trails
│   │   └── responses.go                # Fake response generation
│   │
│   ├── alerts/
│   │   ├── alerter.go                  # Alert interface
│   │   ├── manager.go                  # Alert manager
│   │   ├── slack.go                    # Slack alerts
│   │   ├── discord.go                  # Discord alerts
│   │   ├── webhook.go                  # Webhook alerts
│   │   └── email.go                    # Email alerts
│   │
│   ├── storage/
│   │   ├── storage.go                  # Storage interface
│   │   ├── models/
│   │   │   ├── event.go                # Event model
│   │   │   ├── attacker.go             # Attacker model
│   │   │   ├── campaign.go             # Campaign model
│   │   │   └── ioc.go                  # IOC model
│   │   ├── sqlite.go                   # SQLite implementation
│   │   └── postgres.go                 # PostgreSQL implementation
│   │
│   ├── dashboard/
│   │   ├── dashboard.go                # Dashboard server
│   │   ├── api.go                      # Dashboard API
│   │   └── handlers.go                 # Dashboard handlers
│   │
├── pkg/                                      # Public, reusable packages
│   ├── errors/
│   │   ├── errors.go                         # Core Error type, constructors, chain helpers
│   │   ├── kinds.go                          # Kind type — broad error categories
│   │   ├── codes.go                          # Code type — specific error identifiers
│   │   └── stack.go                          # Stack trace capture and formatting
│   ├── httputil/
│   │   ├── request.go                        # Request helpers (body, headers)
│   │   ├── response.go                       # Response writers (JSON, errors)
│   │   └── headers.go                        # Header normalization, ordering
│   ├── crypto/
│   │   ├── hash.go                           # SHA256, MD5 for fingerprints
│   │   └── random.go                         # UUID v4, secure random strings
│   ├── netutil/
│   │   ├── ip.go                             # IPv4/IPv6 normalization, CIDR
│   │   └── dns.go                            # Reverse DNS lookup helpers
│   ├── timeutil/
│   │   └── time.go                           # RFC3339 formatting, UTC, durations
│   └── validate/
│       └── validate.go                       # Reusable validation (IPs, URLs, severity)
│
├── web/
│   ├── dashboard/                      # Dashboard frontend
│   │   ├── src/
│   │   └── package.json
│   └── assets/
│       └── fake-responses/             # Fake response files
│
├── signatures/
│   ├── scanners.yaml                   # Scanner signatures
│   └── patterns.yaml                   # Attack patterns
│
├── configs/
│   └── firewatch.yaml                  # Default config
│
├── deployments/
│   ├── docker/
│   │   ├── Dockerfile
│   │   └── docker-compose.yaml
│   └── kubernetes/
│       ├── deployment.yaml
│       └── service.yaml
│
├── docs/
│   └── ROADMAP.md                      # This file
│
├── notes/                                # Learning notes
│   ├── go/
│   │   ├── custom-types-and-iota.md      # Type defs, iota, zero values, methods
│   │   ├── error-handling.md             # error interface, wrapping, type switches
│   │   ├── crypto-and-binary.md          # crypto/rand, bit ops, hex encoding
│   │   ├── io-and-net.md                 # LimitReader, ParseIP, CIDR, context, DNS
│   │   ├── struct-tags-and-marshaling.md # yaml/json tags, omitempty, unmarshal defaults
│   │   ├── interfaces-and-satisfaction.md # Implicit satisfaction, scanner pattern, blank imports
│   │   ├── database-sql.md              # Exec/Query/QueryRow, NullString, WAL, upsert
│   │   ├── http-server.md               # Handler, ServeMux, ResponseWriter, StripPrefix
│   │   ├── concurrency-and-signals.md   # Goroutines, channels, select, signal handling
│   │   ├── dependency-injection.md      # Constructor injection, slog.With, method values
│   │   ├── tls-and-crypto-tls.md        # TLS handshake, tls.Config, GetConfigForClient
│   │   ├── sync-and-concurrency-primitives.md # RWMutex, map safety, put/take pattern
│   │   ├── regex.md                     # regexp, RE2 vs PCRE, raw strings, compile caching
│   │   ├── map-as-type.md              # Named map types, lookup tables, struct slices
│   │   ├── http-client.md              # Client, NewRequestWithContext, response handling
│   │   ├── waitgroup-and-fan-out.md    # WaitGroup, fan-out, loop variable capture
│   │   ├── form-parsing-and-http-redirect.md # ParseForm, FormValue, http.Redirect
│   │   ├── raw-strings-and-string-search.md  # Backtick literals, manual substring search
│   │   ├── switch-and-dispatch.md      # Switch patterns, bare switch, type switch
│   │   ├── encoding-csv-and-json.md    # csv.Writer, MarshalIndent, bytes.Buffer
│   │   ├── net-url-and-host-parsing.md # url.Parse, SplitHostPort, IP vs domain
│   │   └── sort-and-stable-keys.md     # sort.Strings, copy-before-sort, composite keys
│   ├── networking/
│   ├── security/
│   │   ├── honeypot-design.md           # Deception, signatures, severity, event recording
│   │   ├── fingerprinting-techniques.md # JA3, header analysis, anomalies, signal combining
│   │   ├── response-mimicry.md         # Headers, status codes, honey tokens, graduated response
│   │   └── threat-intel-formats.md    # STIX 2.1, MISP, CSV for intel sharing
│   └── patterns/
│       ├── error-architecture.md         # Kind/Code/Op design, pkg vs internal
│       ├── repository-pattern.md         # Store interface, filters, query builder
│       ├── middleware-chain.md           # Composition, closures, embedding, context values
│       ├── strategy-and-registry.md     # Module interface, registry, route declaration
│       ├── detection-engine.md          # Sig vs pattern, field extraction, severity ranking
│       ├── observer-and-dispatch.md    # Observer pattern, fan-out dispatch, severity gating
│       ├── module-uniformity.md        # Shared helpers, DRY RecordEvent, module template
│       ├── pipeline-orchestrator.md    # Extract → enrich → detect → persist pipeline
│       └── clustering-and-dedup.md     # Map dedup, composite keys, set operations
│
├── test/
│   ├── integration/
│   └── fixtures/
│
├── CLAUDE.md                             # Development prompt
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Build Phases

### Phase 1: Foundation

- `pkg/errors/` — Error types, kinds, codes, stack traces ✅
- `pkg/httputil/` — Request/response helpers, header constants ✅
- `pkg/crypto/` — SHA256, MD5 hashing, UUID v4, random tokens ✅
- `pkg/netutil/` — IP normalization, CIDR matching, reverse DNS ✅
- `pkg/timeutil/` — RFC3339 formatting, UTC, duration parsing ✅
- `pkg/validate/` — IP, URL, port, severity, generic validators ✅
- `internal/config/` — Configuration loading, defaults, validation ✅
- `internal/storage/models/` — Event, Attacker, Campaign, IOC models ✅
- `internal/storage/storage.go` — Store interface, filter types ✅
- `internal/storage/sqlite.go` — SQLite implementation (WAL, migrations, CRUD) ✅

### Phase 2: Server Core

- `internal/middleware/middleware.go` — Middleware type, Chain, responseWriter ✅
- `internal/middleware/correlation.go` — Request ID generation/propagation ✅
- `internal/middleware/logging.go` — Structured request logging (slog) ✅
- `internal/server/server.go` — HTTP server, Start, Shutdown ✅
- `internal/server/router.go` — Route registration, fallback, Mount ✅
- `internal/server/tls.go` — TLS 1.2+ with modern cipher suites ✅
- `internal/server/graceful.go` — Signal handling, ListenAndShutdown ✅

### Phase 3: First Honeypot Module

- `internal/handlers/handler.go` — Module interface, Route type ✅
- `internal/handlers/registry.go` — Module registry with lookup/filter ✅
- `internal/handlers/nextjs/nextjs.go` — Module entry, route table ✅
- `internal/handlers/nextjs/server_action.go` — next-action header detection ✅
- `internal/handlers/nextjs/rsc.go` — RSC endpoint/header probes ✅
- `internal/handlers/nextjs/static.go` — Static asset enumeration ✅
- `internal/handlers/nextjs/event.go` — Event recording, page serving ✅
- `internal/deception/responses.go` — Fake Next.js, WordPress, .env responses ✅

### Phase 4: Fingerprinting

- `internal/fingerprint/ja3.go` — JA3 computation, JA3Store, TLS callback ✅
- `internal/fingerprint/headers.go` — Header order hash, anomaly detection, known client matching ✅
- `internal/fingerprint/fingerprint.go` — Engine, Result type, context storage ✅
- `internal/middleware/fingerprint.go` — Middleware wiring engine into request pipeline ✅

### Phase 5: Detection

- `internal/detection/signatures.go` — Signature type, Matcher engine, 14 built-in signatures ✅
- `internal/detection/patterns.go` — Pattern type, Rule/Category, 5 built-in patterns ✅
- `internal/detection/detector.go` — Detector engine, field extraction, severity ranking ✅

### Phase 6: Alerting

- `internal/alerts/alerter.go` — Alerter interface, Alert type, severity filtering ✅
- `internal/alerts/manager.go` — Concurrent dispatch, per-alerter severity thresholds ✅
- `internal/alerts/slack.go` — Slack Block Kit formatted webhooks ✅
- `internal/alerts/discord.go` — Discord embed formatted webhooks ✅
- `internal/alerts/webhook.go` — Generic JSON webhook with custom headers ✅

### Phase 7: Additional Modules

- `internal/handlers/event.go` — Shared RecordEvent helper for all modules ✅
- `internal/handlers/wordpress/wordpress.go` — Module entry, 8 routes ✅
- `internal/handlers/wordpress/login.go` — wp-login GET/POST, brute force detection ✅
- `internal/handlers/wordpress/admin.go` — wp-admin, wp-json API, static assets ✅
- `internal/handlers/wordpress/xmlrpc.go` — XML-RPC probe/payload detection ✅
- `internal/handlers/exposure/exposure.go` — Module entry, 14 routes ✅
- `internal/handlers/exposure/env.go` — .env file probes with fake content ✅
- `internal/handlers/exposure/git.go` — .git/config, .git/HEAD probes ✅
- `internal/handlers/exposure/config.go` — Config file probes (403 responses) ✅
- `internal/handlers/api/api.go` — Module entry, 9 routes ✅
- `internal/handlers/api/rest.go` — REST API probes, auth header detection ✅
- `internal/handlers/api/graphql.go` — GraphQL probes, introspection detection ✅
- `internal/handlers/api/swagger.go` — Swagger/OpenAPI fake spec ✅
- `internal/handlers/cloud/cloud.go` — Module entry, 6 routes (AWS, DigitalOcean) ✅
- `internal/handlers/cloud/metadata.go` — Metadata listing, fake IAM creds, IMDSv2 ✅

### Phase 8: Intel & Export

- `internal/intel/ioc/extractor.go` — IOC extraction from events, deduplication, tag merging ✅
- `internal/intel/enrichment/enricher.go` — Enricher interface ✅
- `internal/intel/enrichment/dns.go` — Reverse DNS enrichment for IP-type IOCs ✅
- `internal/intel/enrichment/geoip.go` — GeoIP enrichment placeholder ✅
- `internal/intel/collector.go` — Orchestrator: extract → enrich → campaign detect → persist ✅
- `internal/detection/campaign.go` — Campaign detection: signature clustering + coordination detection ✅
- `internal/intel/export/exporter.go` — Exporter interface (Name, ContentType, ExportIOCs, ExportCampaigns) ✅
- `internal/intel/export/stix.go` — STIX 2.1 Bundle with Indicator/Campaign SDOs ✅
- `internal/intel/export/misp.go` — MISP event format with attributes ✅
- `internal/intel/export/csv.go` — CSV export for IOCs and campaigns ✅

### Phase 9: CLI & Polish

- `cmd/firewatch/main.go` — CLI
- Dashboard (optional)
- Documentation, tests
