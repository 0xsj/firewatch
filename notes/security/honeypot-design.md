# Honeypot Design Principles

## What a Honeypot Module Does

A honeypot module has three jobs:

1. **Deceive** — return realistic responses so the scanner continues
2. **Detect** — classify what the scanner is doing (signatures)
3. **Record** — save the full interaction as a structured event

```
Scanner → [Request] → Module → [Classify + Record]
                              → [Fake Response] → Scanner
```

The scanner should believe it found a real application. The more
convincing the deception, the more the scanner reveals about itself.

---

## Deception Techniques

### Realistic responses

Every response should match what the real application would return:

| What to match          | Example                              |
|------------------------|--------------------------------------|
| Content-Type           | `text/x-component` for RSC           |
| Response headers       | `X-Powered-By: Next.js`              |
| Cache headers          | `immutable` for static assets        |
| HTML structure          | Real `<div id="__next">` layout     |
| Error pages            | Next.js-style 404 with exact CSS     |
| Version artifacts      | `wp-login.php?ver=6.4.2`             |

### Breadcrumbs

Plant paths for scanners to follow. A fake HTML page references
`/_next/static/chunks/main-app.js` — a scanner that parses HTML
will request that path next, revealing more about its behavior.

### Honey tokens

Fake credentials that are monitored. The `.env` honeypot returns:
```
AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
```
If this key appears in AWS CloudTrail logs, someone used the
credentials they found in the honeypot.

### Dynamic elements

Nonces and timestamps change per request, making responses look
fresh rather than static:

```go
nonce, _ := crypto.RandomHex(16)
deception.NextJSPage("App", nonce)
```

**Used in:** `internal/deception/responses.go`

---

## Signature-Based Detection

Each suspicious behavior gets a signature ID:

```go
const (
    sigServerActionProbe   = "nextjs-action-001"
    sigServerActionPayload = "nextjs-action-002"
)
```

### Naming convention

`<module>-<type>-<sequence>`

| Part       | Examples                       |
|------------|--------------------------------|
| module     | `nextjs`, `wordpress`, `api`   |
| type       | `action`, `rsc`, `static`      |
| sequence   | `001`, `002`, `003`            |

### Layered signatures

A single request can trigger multiple signatures:

```go
sigs := []string{sigServerActionProbe}

if len(body) > 0 {
    sigs = append(sigs, sigServerActionPayload)
}
```

This lets detection rules correlate:
- `nextjs-action-001` alone = basic probe
- `nextjs-action-001` + `nextjs-action-002` = active exploitation attempt

**Used in:** `internal/handlers/nextjs/server_action.go`, `rsc.go`, `static.go`

---

## Severity Classification

Every event gets a severity based on what it reveals about intent:

| Severity | Meaning                     | Examples                        |
|----------|-----------------------------|---------------------------------|
| critical | Active exploitation         | Working exploit payloads        |
| high     | Targeted probing            | Server Action probes, debug endpoints |
| medium   | Informed enumeration        | RSC headers, build manifest     |
| low      | Generic scanning            | Static asset requests           |
| info     | Background noise            | Generic page loads              |

Severity escalates with specificity:
- `GET /` → info (could be anyone)
- `GET /` with `Rsc` header → medium (knows about Next.js internals)
- `POST /` with `Next-Action` header → high (targeting specific vuln)
- `POST /` with exploit payload → critical

**Used in:** all handler files

---

## Event Recording

Every interaction becomes an `Event` in storage:

```go
event := &models.Event{
    ID:         crypto.UUID4(),
    Timestamp:  timeutil.FormatRFC3339(timeutil.NowUTC()),
    RequestID:  middleware.RequestID(r.Context()),
    SourceIP:   httputil.ClientIP(r),
    Module:     moduleName,
    Method:     r.Method,
    Path:       r.URL.Path,
    Headers:    httputil.HeaderMap(r.Header),
    Severity:   severity,
    Signatures: signatures,
}
```

Note how this wires together packages from across the codebase:
- `pkg/crypto` → event ID
- `pkg/timeutil` → timestamp
- `pkg/httputil` → client IP, headers
- `middleware` → request ID from correlation middleware
- `storage/models` → the Event struct
- `storage.Store` → persistence

The module handler has the most context about what the request means
(which signatures matched, what severity), while the middleware
provides cross-cutting data (request ID, fingerprint).

**Used in:** `internal/handlers/nextjs/event.go`
