# Architecture

## Overview

Firewatch is an HTTP honeypot server written in Go. It deploys fake vulnerable web application endpoints across 7 modules, captures every request through a middleware pipeline that fingerprints and classifies traffic, stores events in SQLite, dispatches real-time alerts, profiles attackers, correlates campaigns, and exports threat intelligence in standard formats.

## System Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                        HTTP Server                          │
│                   (net/http + TLS)                           │
├─────────────────────────────────────────────────────────────┤
│                     Middleware Chain                         │
│                                                             │
│  Request ──▶ Correlation ──▶ IPFilter ──▶ RateLimit ──▶    │
│              (request ID)    (allow/block) (token bucket)    │
│                                                             │
│  ──▶ Logging ──▶ GeoIP ──▶ Fingerprint ──▶ Detection ──▶  │
│      (slog)     (MaxMind)  (JA3/JA4)      (sig+pattern)    │
│                                                             │
│  ──▶ Behavior ──▶ Router ──▶ Module Handler                │
│      (temporal)                                             │
├─────────────────────────────────────────────────────────────┤
│                    Honeypot Modules                          │
│                                                             │
│  nextjs │ wordpress │ api │ exposure │ admin │ cloud │ cve  │
│                                                             │
│  Each module:                                               │
│    - Registers routes via handlers.Route                    │
│    - Serves deception responses                             │
│    - Records events via handlers.RecordEvent()              │
├──────────────────────┬──────────────────────────────────────┤
│      Alerting        │         Threat Intel                 │
│                      │                                      │
│  AlertingStore       │  Collector                           │
│   ├─ Slack           │   ├─ IOC Extractor                  │
│   ├─ Discord         │   ├─ Enrichment (GeoIP, DNS)        │
│   └─ Webhook         │   ├─ Campaign Correlator (bg)       │
│                      │   └─ Export (STIX, MISP, CSV)        │
├──────────────────────┴──────────────────────────────────────┤
│                     Store Decorators                         │
│            SQLite → ProfilingStore → AlertingStore           │
│                                                             │
│  Events │ Attackers │ Campaigns │ IOCs                      │
└─────────────────────────────────────────────────────────────┘
```

## Components

### HTTP Server (`internal/server/`)

**Responsibility:** Accept HTTP/HTTPS connections and route requests through middleware to module handlers.

- `server.go` — `Server` struct, constructor takes `(Config, Store, *fingerprint.Engine, *detection.Detector, *geoip.Reader, *slog.Logger)`. Builds middleware chain, starts background goroutines (rate limiter, behavior tracker, campaign correlator), and wraps the router.
- `router.go` — Go 1.22+ `http.ServeMux` with pattern-based routing.
- `tls.go` — TLS 1.2+ configuration with modern cipher suites.
- `graceful.go` — `ListenAndShutdown()` handles OS signals (SIGINT, SIGTERM) for clean shutdown.

**Key design:** Fingerprint engine and detector are optional — the server works without them. Middleware chain is assembled conditionally at startup.

### Middleware (`internal/middleware/`)

**Responsibility:** Process every request before it reaches a module handler.

```
type Middleware func(http.Handler) http.Handler
```

Chain is composed left-to-right via `Chain()`:

1. **Correlation** — Generates a UUID request ID, stores it in context.
2. **IPFilter** — Checks IP against allowlist/blocklist (CIDR + individual). Blocks rejected IPs early.
3. **RateLimit** — Per-IP token bucket rate limiting. Returns 429 when exceeded.
4. **Logging** — Structured request logging via `slog` (method, path, status, duration).
5. **GeoIP** — MaxMind lookup, stores country/city/ASN in context.
6. **Fingerprint** — Runs `fingerprint.Engine` (JA3/JA4 + headers), stores result in context.
7. **Detection** — Evaluates signatures and patterns against the request, records matches as events.
8. **Behavior** — Records request to `BehaviorTracker`, analyzes temporal patterns per-IP.

### Honeypot Modules (`internal/handlers/`)

**Responsibility:** Serve convincing fake responses and record attacker interactions.

All modules implement the `Module` interface:

```
type Module interface {
    Name() string
    Routes() []Route
}
```

Each module:
- Receives `(config, store, logger)` via constructor
- Returns a list of `Route{Pattern, Handler}` structs
- Calls `handlers.RecordEvent()` to persist events with severity and signature IDs
- Serves responses from `internal/deception/responses.go`

**Modules:**

| Module    | Routes | Key Behaviors                                           |
|-----------|--------|---------------------------------------------------------|
| nextjs    | 7      | RSC detection, server action probes, source map access  |
| wordpress | 8      | Login brute force, XML-RPC payload detection            |
| api       | 9      | REST/GraphQL enumeration, auth header sniffing          |
| exposure  | 14     | .env, .git, config file probes                          |
| admin     | 16     | phpMyAdmin, Adminer, cPanel, generic admin panels       |
| cloud     | 6      | AWS/DO metadata, IAM creds, IMDSv2 token               |
| cve       | 14     | Log4Shell JNDI, Spring4Shell, MOVEit, PAN-OS, Struts2 OGNL, Confluence |

### Detection Engine (`internal/detection/`)

**Responsibility:** Match requests against known scanner signatures and attack patterns.

- **Signatures** — AND-logic matchers (all conditions must match). Each has an ID, module scope, severity, and a list of field matchers. ~26 built-in signatures.
- **Patterns** — OR-logic rules grouped by category (e.g., "SQL Injection"). Any rule match fires the pattern.
- **Detector** — Extracts fields from the request (path, method, body, user-agent, query, headers), evaluates all signatures and patterns, returns the highest severity match.
- **Campaign Detection** — Clusters events by signature overlap and coordinated multi-IP module targeting.
- **Behavioral Tracking** — Per-IP temporal analysis detecting scan sweeps, brute force, module hopping, and progressive recon escalation.
- **Custom Signatures** — YAML-based loading from files and directories. Same-ID overrides built-in. Regex validated at load time.

Field matchers support: `equals`, `contains`, `prefix`, `suffix`, `regex`, `exists`.

### Campaign Correlation (`internal/detection/correlator.go`)

**Responsibility:** Background detection of coordinated attack campaigns.

`CampaignCorrelator` runs as a background goroutine (same lifecycle pattern as `BehaviorTracker` and `RateLimiter`). Every N seconds it:

1. Queries recent events from the store (sliding window)
2. Runs `CampaignDetector` to find signature clusters (same sigs, multiple IPs) and coordinated attacks (same module sets, multiple IPs)
3. Creates or updates campaigns with stable IDs (name -> ID map persisted across ticks)
4. Links events to campaigns via `UpdateEventLinks` (preserves existing attacker_id)

### Fingerprinting (`internal/fingerprint/`)

**Responsibility:** Build a multi-signal fingerprint for each request.

- **JA3** — TLS client hello fingerprinting via `GetConfigForClient` callback. Requires TLS enabled.
- **JA4** — Extended TLS fingerprint (protocol, version, SNI, cipher/extension hashes). Requires TLS enabled.
- **Headers** — Header key ordering hash, known client identification, anomaly detection.
- **Engine** — Combines JA3 + JA4 + header analysis into a `Result` stored in request context.

### Storage (`internal/storage/`)

**Responsibility:** Persist and query honeypot data.

Single `Store` interface with methods for four entity types:

```
Events     — SaveEvent, GetEvent, ListEvents, CountEvents
Attackers  — SaveAttacker, GetAttacker, GetAttackerByIP, ListAttackers
Campaigns  — SaveCampaign, GetCampaign, ListCampaigns
IOCs       — SaveIOC, ListIOCs
```

Each query method accepts a typed filter struct (e.g., `EventFilter` with `Since`, `Until`, `SourceIP`, `Module`, `Severity`, `Limit`, `Offset`).

**Implementation:** SQLite via `modernc.org/sqlite` (pure Go, no CGO). WAL mode enabled.

**Store Decorators** (applied in order):

1. **ProfilingStore** (`profiling.go`) — Intercepts `SaveEvent` to asynchronously create/update Attacker profiles per-IP. Tracks user agents, modules, paths, JA3, severity escalation, auto-tags.
2. **AlertingStore** (`alerting.go`) — Intercepts `SaveEvent` to dispatch alerts asynchronously via the alert Manager. Then delegates to the underlying store.

Wrapping order: `SQLiteStore -> ProfilingStore -> AlertingStore`

### Alerting (`internal/alerts/`)

**Responsibility:** Dispatch real-time notifications when events are saved.

- **Alerter interface** — `Send(ctx, Alert) error`
- **Manager** — Registers multiple alerters, each with a minimum severity threshold. Dispatches concurrently with 10-second timeout per alerter.
- **Implementations:** Slack (Block Kit), Discord (embeds), generic webhook (JSON + custom headers).

### Threat Intelligence (`internal/intel/`)

**Responsibility:** Extract, enrich, and export actionable intelligence.

```
Events ──▶ IOC Extractor ──▶ Enrichment ──▶ Campaign Detection ──▶ Storage
                                                                       │
                                                              Export ◀──┘
                                                          (STIX/MISP/CSV)
```

- **IOC Extractor** — Extracts IP addresses, URLs, user agents, JA3 hashes from events. Deduplicates and merges tags.
- **Enrichment** — Interface-based. Current implementations: reverse DNS, GeoIP (placeholder).
- **Campaign Detection** — Signature clustering + time-based coordination detection.
- **Export** — STIX 2.1 bundles, MISP event format, CSV.

### Deception (`internal/deception/`)

**Responsibility:** Generate convincing fake responses.

`responses.go` contains static functions returning realistic HTML/JSON for:
- Next.js pages, error pages, RSC payloads, build manifests
- WordPress login, phpMyAdmin, Adminer, cPanel, generic admin panels
- Apache Solr admin, Spring Boot actuator, MOVEit Transfer login
- Struts2 showcase, Confluence login
- Exposed .env files

## Data Flow

### Typical Request

```
Incoming Request
      │
      ▼
  Correlation ──▶ assigns request ID
      │
      ▼
  IPFilter ──▶ allow/blocklist check (rejects → 403 + event)
      │
      ▼
  RateLimit ──▶ token bucket check (rejects → 429 + event)
      │
      ▼
  Logging ──▶ logs method, path, IP
      │
      ▼
  GeoIP ──▶ MaxMind lookup → context
      │
      ▼
  Fingerprint ──▶ JA3/JA4 hash + header analysis → context
      │
      ▼
  Detection ──▶ matches signatures/patterns → saves detection event
      │
      ▼
  Behavior ──▶ records to BehaviorTracker → tags if pattern detected
      │
      ▼
  Router ──▶ dispatches to module handler
      │
      ▼
  Module Handler
    ├── RecordEvent() → Store.SaveEvent()
    │                        │
    │                   ProfilingStore intercepts
    │                     └── async: create/update Attacker profile
    │                        │
    │                   AlertingStore intercepts
    │                     ├── Slack
    │                     ├── Discord
    │                     └── Webhook
    │
    └── Write deception response

  Background:
    CampaignCorrelator ──▶ periodic: query events → detect campaigns → link events
```

### Alert Flow

```
Store.SaveEvent()
      │
      ▼
  AlertingStore.SaveEvent()
    ├── Delegate to underlying store
    └── Async: Manager.Send()
              ├── Check severity ≥ threshold
              ├── Goroutine per alerter
              └── 10s timeout each
```

## Key Design Decisions

**Pure Go SQLite.** Using `modernc.org/sqlite` avoids CGO, simplifying cross-compilation and Docker builds. Trades some performance for portability.

**Optional middleware.** Fingerprinting and detection can be nil. The server degrades gracefully — useful for minimal deployments or testing.

**AlertingStore decorator.** Rather than coupling alerting into every handler, a single store wrapper intercepts all event saves. Modules don't know about alerts.

**Module interface.** All 7 modules share the same `Module` interface and `RecordEvent` helper. Adding a new module means implementing `Name()` and `Routes()` — no changes to server code.

**Detection middleware + handler recording.** Detection middleware catches known patterns at the pipeline level. Handlers record module-specific events with their own signature IDs. Both paths feed the same storage.

**CVE filtering.** The CVE module accepts a `CVEs` config list. Empty means all 6 CVEs are active. Populated means only those CVEs register routes. This allows targeted deployments.
