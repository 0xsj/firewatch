# TLS and crypto/tls

## TLS Handshake Overview

Before any HTTP data is exchanged over HTTPS, the client and server
perform a TLS handshake:

```
Client                          Server
  |--- ClientHello --------------->|   (versions, ciphers, curves)
  |<------------- ServerHello -----|   (chosen version, cipher)
  |<------------ Certificate -----|   (server's cert)
  |--- ClientKeyExchange -------->|
  |--- Finished ------------------>|
  |<---------------- Finished -----|
  |=== Encrypted HTTP begins =====>|
```

The **ClientHello** is the fingerprinting goldmine — it reveals the
client's TLS implementation before any application data is sent.

---

## crypto/tls.Config

Go's TLS configuration struct:

```go
cfg := &tls.Config{
    MinVersion: tls.VersionTLS12,
    CurvePreferences: []tls.CurveID{
        tls.X25519,
        tls.CurveP256,
    },
    CipherSuites: []uint16{
        tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
        tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
    },
}
```

Key fields:
- `MinVersion` — reject clients using older TLS (1.0, 1.1)
- `CurvePreferences` — preferred elliptic curves for key exchange
- `CipherSuites` — allowed cipher suites (order = preference)

**Used in:** `internal/server/tls.go`

---

## GetConfigForClient Callback

`tls.Config.GetConfigForClient` is called during every TLS handshake
with the client's hello parameters:

```go
cfg.GetConfigForClient = func(info *tls.ClientHelloInfo) (*tls.Config, error) {
    // info contains the ClientHello parameters
    // Return nil, nil to use the default config
    return nil, nil
}
```

### ClientHelloInfo fields

```go
type ClientHelloInfo struct {
    CipherSuites      []uint16       // cipher suites offered
    SupportedVersions []uint16       // TLS versions offered
    SupportedCurves   []tls.CurveID  // elliptic curves offered
    SupportedPoints   []uint8        // point formats offered
    ServerName        string         // SNI hostname
    Conn              net.Conn       // the raw connection
}
```

This is how we capture JA3 data. The callback fires during the
handshake (before any HTTP handler runs), so we store the data
in a `JA3Store` and retrieve it later in the HTTP middleware.

### Return value

- `return nil, nil` → use the default TLS config (most common)
- `return customCfg, nil` → use a different config for this connection
- `return nil, err` → abort the handshake

**Used in:** `internal/fingerprint/ja3.go` — `TLSConfigCallback()`

---

## Bridging TLS and HTTP

The TLS handshake and HTTP handler run at different times:

```
Time ──────────────────────────────────────────────►
     │ TLS Handshake │       HTTP Request/Response │
     │               │                             │
     │ GetConfigFor  │  Middleware → Handler        │
     │ Client fires  │  reads JA3Store              │
     │ stores hello  │                              │
```

We bridge them with a concurrent map keyed by remote address:

```go
// During TLS handshake (ja3.go)
store.Put(info.Conn.RemoteAddr().String(), hello)

// During HTTP handling (middleware/fingerprint.go)
hello := store.Take(r.RemoteAddr)
```

`Take()` is read-and-delete — it retrieves the entry and removes
it in one atomic operation, preventing memory growth from completed
connections.

**Used in:** `internal/fingerprint/ja3.go` — `JA3Store`

---

## Why MD5 for JA3?

JA3 uses MD5 by convention, not security. The hash is a compact
identifier for the fingerprint string, not a cryptographic commitment.
MD5's speed and 32-character hex output make it convenient for
database lookups and log entries.

The full JA3 string (e.g., `771,4866-4867-4865-49196,0-23-65281,29-23,0`)
is kept as `ja3_raw` for detailed analysis. The MD5 hash (`ja3_hash`)
is used for quick matching against known fingerprint databases.

**Used in:** `internal/fingerprint/ja3.go` — `JA3()`
