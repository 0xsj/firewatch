# Firewatch Development Guide

## Objective

Build a honeypot server that masquerades as vulnerable web applications to detect, fingerprint, and analyze attackers. Generate actionable threat intelligence from captured interactions while providing real-time alerting.

## Scope

**In Scope:**

- HTTP/HTTPS honeypot server
- Multiple honeypot modules (Next.js, WordPress, APIs, etc.)
- Request fingerprinting (JA3, headers, behavior)
- Attacker profiling and campaign detection
- Threat intelligence export (STIX, MISP, CSV)
- Real-time alerting (Slack, Discord, webhooks)
- Deception techniques (fake responses, honey tokens)

**Out of Scope:**

- Network-level honeypots (SSH, Telnet)
- Malware collection and analysis
- Active response or blocking
- Full production application emulation

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                      HTTP Server                         │
├─────────────────────────────────────────────────────────┤
│                      Middleware                          │
│    Logging │ Fingerprint │ GeoIP │ Correlation          │
├─────────────────────────────────────────────────────────┤
│                   Honeypot Modules                       │
│  Next.js │ WordPress │ API │ Exposure │ Cloud │ CVE    │
├─────────────────────────────────────────────────────────┤
│                      Detection                           │
│         Signatures │ Patterns │ Campaigns               │
├──────────────────────┬──────────────────────────────────┤
│       Alerting       │         Intel                     │
│ Slack │ Discord │ WH │  IOC │ Enrichment │ Export       │
├──────────────────────┴──────────────────────────────────┤
│                       Storage                            │
│                 SQLite / PostgreSQL                      │
└─────────────────────────────────────────────────────────┘
```

## Development Guidelines

### Testing Strategy

- Unit tests for fingerprinting, detection, IOC extraction
- Table-driven tests for signature matching
- HTTP tests with `httptest` for handlers
- Mock storage for handler tests
- Integration tests for full request flow
- Recorded scanner traffic as fixtures

### Design Patterns

- **Strategy pattern:** Swappable honeypot modules
- **Observer pattern:** Event → alert dispatch
- **Factory pattern:** Module creation
- **Chain of responsibility:** Middleware pipeline
- **Repository pattern:** Storage abstraction

### Error Handling

- Wrap errors: `fmt.Errorf("processing request: %w", err)`
- Never crash on malformed requests
- Log all errors with request context
- Graceful degradation if enrichment fails

### Logging

- Structured logging with zerolog or slog
- DEBUG: Raw requests, fingerprint details
- INFO: Detected attacks, alerts sent
- WARN: Enrichment failures, rate limits
- ERROR: Server errors only
- Include request ID in all log entries

## Learning Objectives

Document these concepts in `notes/`:

### Go Concepts (`notes/go/`)

- HTTP server with graceful shutdown
- Middleware chaining patterns
- TLS configuration and JA3 extraction
- Embed directive for fake responses

### Networking (`notes/networking/`)

- TLS fingerprinting (JA3/JA4)
- HTTP header ordering analysis
- GeoIP and ASN lookups
- Reverse DNS resolution

### Security (`notes/security/`)

- Honeypot design principles
- Common scanner patterns
- CVE emulation techniques
- Threat intelligence formats (STIX, MISP)

### Patterns (`notes/patterns/`)

- Middleware pipeline architecture
- Event-driven alerting
- Signature-based detection
- Campaign correlation algorithms

## Build Order

1. `internal/pkg/` — HTTP, crypto, net utilities
2. `internal/config/` — Configuration loading
3. `internal/storage/models/` — Event, attacker models
4. `internal/storage/sqlite.go` — SQLite storage
5. `internal/server/server.go` — HTTP server
6. `internal/middleware/logging.go` — Request logging
7. `internal/handlers/handler.go` — Handler interface
8. `internal/handlers/nextjs/` — First honeypot module
9. `internal/deception/responses.go` — Fake responses
10. `internal/fingerprint/` — Fingerprinting engine
11. `internal/detection/` — Detection engine
12. `internal/alerts/` — Alerting system
13. Additional honeypot modules
14. `internal/intel/` — Threat intelligence
15. `cmd/firewatch/main.go` — CLI

## Conventions

- Event IDs: UUID v4
- Timestamps: UTC, RFC3339 format
- IP addresses: Normalized IPv4/IPv6
- Severity: critical, high, medium, low, info
- Module names: lowercase (nextjs, wordpress, api)
- Signature IDs: `<module>-<type>-<sequence>` (e.g., `nextjs-probe-001`)
