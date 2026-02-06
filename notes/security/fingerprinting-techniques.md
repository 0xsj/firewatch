# Fingerprinting Techniques

## Why Fingerprint?

IP addresses change. User-Agents are spoofed. But the _how_ of a
request — which TLS ciphers were offered, what order headers arrived
in, which standard headers are missing — reveals the underlying
client implementation.

Fingerprinting answers: "What tool is making this request?" regardless
of what it _claims_ to be.

---

## JA3 — TLS Client Fingerprinting

### What it captures

JA3 hashes five fields from the TLS ClientHello message:

```
TLSVersion, CipherSuites, Extensions, EllipticCurves, PointFormats
```

Each field is a dash-separated list of decimal values:

```
771,4866-4867-4865-49196-49200,0-23-65281-10-11-35,29-23-24,0
```

This string is MD5-hashed to produce the fingerprint:

```
e7d705a3286e19ea42f587b344ee6865
```

### What it reveals

Different TLS implementations offer different cipher suites in
different orders:

| Client          | JA3 Hash                         |
|-----------------|----------------------------------|
| Chrome 120      | `cd08e31494f9531f560d64c695473da9` |
| Firefox 121     | `b32309a26951912be7dba376398abc3b` |
| python-requests | `eb8db8404524e8ad95c3daa0e4e9d235` |
| curl 8.x        | `456523fc94726331a4d5a2e1d40b2cd7` |
| Go http client  | `a0e9f5d64349fb13191bc781f81f42e8` |

A request claiming `User-Agent: Mozilla/5.0 Chrome` but with a
python-requests JA3 hash is immediately suspicious.

### Limitations in Go

Go's `tls.ClientHelloInfo` doesn't expose the raw extensions list,
so our JA3 has an empty extensions section. The remaining four fields
still provide useful differentiation — cipher suite ordering alone
is a strong signal.

**Used in:** `internal/fingerprint/ja3.go`

---

## Header Analysis

### Header order

Different HTTP clients send headers in different orders. A browser
might send:

```
Host, Connection, Accept, User-Agent, Accept-Encoding, Accept-Language
```

While curl sends:

```
Host, User-Agent, Accept
```

And python-requests sends:

```
User-Agent, Accept-Encoding, Accept, Connection, Host
```

We hash the sorted key set to create a stable identifier. True
wire order isn't preserved by Go's `net/http` (headers are stored
in a map), but the key _set_ is still a useful signal.

### Anomaly detection

Real browsers always send certain headers. Their absence is a
strong indicator of automated tooling:

| Missing header      | What it suggests                    |
|---------------------|-------------------------------------|
| `Accept`            | Not a browser (scripts often skip)  |
| `Accept-Language`   | Not a browser (always sent by all)  |
| `Accept-Encoding`   | Very unusual — even curl sends this |
| `User-Agent`        | Extremely suspicious                |

### Known client matching

Pattern matching on User-Agent identifies common tools:

```go
var knownClients = []struct {
    pattern string
    name    string
}{
    {"python-requests", "python-requests"},
    {"Nuclei", "nuclei"},
    {"sqlmap", "sqlmap"},
    {"zgrab", "zgrab"},
    // ...
}
```

This is a first-pass classifier. Scanners that spoof their
User-Agent will still be caught by JA3 and header anomalies.

**Used in:** `internal/fingerprint/headers.go`

---

## Combining Signals

No single signal is definitive. The fingerprint engine combines
all available data:

```go
type Result struct {
    JA3Hash         string   // TLS implementation
    HeaderOrderHash string   // HTTP client behavior
    UserAgent       string   // Claimed identity
    KnownClient     string   // Matched scanner pattern
    Anomalies       []string // Missing expected headers
}
```

Strongest signals (hardest to spoof):
1. **JA3** — requires changing TLS implementation
2. **Header anomalies** — requires knowing what browsers send
3. **Header order hash** — requires matching browser internals
4. **Known client UA** — trivial to spoof

The detection engine (Phase 5) will combine these signals with
request patterns to classify threats.

---

## Context Propagation

The fingerprint result flows through the request lifecycle via
context:

```
TLS Handshake → JA3Store.Put()
                     ↓
HTTP Request  → Fingerprint Middleware → Engine.Analyze()
                     ↓                        ↓
              context.WithValue()     JA3Store.Take()
                     ↓                 + AnalyzeHeaders()
              Handler reads            → Result
              fingerprint.GetResult(ctx)
```

This keeps fingerprinting decoupled from handling. Handlers that
care about the fingerprint read it from context; handlers that
don't simply ignore it.

**Used in:** `internal/fingerprint/fingerprint.go`, `internal/middleware/fingerprint.go`
