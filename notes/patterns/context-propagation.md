# Context Propagation Pattern

## Problem

Data enriched in middleware (fingerprints, GeoIP lookups, correlation
IDs) needs to flow to handlers and detection logic without:

1. **Polluting function signatures** — adding parameters to every handler
2. **Global state** — mutable globals create race conditions
3. **Tight coupling** — handlers shouldn't know about middleware internals

## Solution: Context Values

Go's `context.Context` provides request-scoped storage that flows
through the call chain:

```
Middleware → ctx.WithValue(key, data)
                     ↓
Handler    → ctx.Value(key) → data
```

The context is immutable and thread-safe — each `WithValue` creates
a new context wrapping the parent.

---

## Implementation Pattern

### 1. Define private key type

Prevents key collisions across packages:

```go
type contextKey string

const fingerprintKey contextKey = "fingerprint"
```

**Why private?** Only this package should create/read this context key.
External code uses exported functions.

### 2. Wrap with type-safe functions

```go
// WithResult stores a fingerprint Result in context
func WithResult(ctx context.Context, result Result) context.Context {
    return context.WithValue(ctx, fingerprintKey, result)
}

// GetResult extracts the fingerprint Result from context
// Returns zero value if not present
func GetResult(ctx context.Context) Result {
    result, _ := ctx.Value(fingerprintKey).(Result)
    return result
}
```

**Why wrappers?** Type safety. Callers get `Result`, not `interface{}`.

### 3. Store in middleware

```go
func Fingerprint(engine *fingerprint.Engine) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            result := engine.Analyze(r)

            // Inject into context
            ctx := fingerprint.WithResult(r.Context(), result)

            // Pass enriched context downstream
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### 4. Read in handlers

```go
func RecordEvent(store storage.Store, logger *slog.Logger, r *http.Request, ...) {
    // Extract from context
    fp := fingerprint.GetResult(r.Context())

    event := &models.Event{
        Fingerprint: models.Fingerprint{
            JA3:     fp.JA3Raw,
            JA3Hash: fp.JA3Hash,
            JA4:     fp.JA4,
        },
    }

    store.SaveEvent(r.Context(), event)
}
```

---

## Used Throughout Firewatch

| Package       | Key              | Data Stored                    |
|---------------|------------------|--------------------------------|
| `middleware`  | `requestID`      | UUID correlation ID            |
| `fingerprint` | `fingerprint`    | JA3, JA4, header analysis      |
| `geoip`       | `geoip`          | Country, city, ASN, org        |
| `detection`   | (future)         | Could store detection results  |

Each package owns its context key namespace. No conflicts.

---

## Advantages

**Decoupling:** Handlers don't import middleware packages. They only
import the data package (`fingerprint`, `geoip`).

**Optional consumption:** Handlers that don't care about fingerprints
simply don't call `GetResult()`. No required parameters.

**Request-scoped:** Data is tied to the request lifetime. Automatically
cleaned up when the request completes.

**Thread-safe:** Contexts are immutable. Safe to read concurrently.

---

## Gotchas

### Don't overuse

Context is for request-scoped data. Don't use it for:
- Application config (use dependency injection)
- Database connections (use `sql.DB` pool)
- Logger instances (pass explicitly or use DI)

### Always provide zero-value fallback

```go
func GetResult(ctx context.Context) Result {
    result, _ := ctx.Value(fingerprintKey).(Result)
    return result  // returns zero value if not found
}
```

This prevents nil pointer panics when middleware is missing.

### Keep keys private

```go
// ❌ Bad — exported key can be used by other packages
const FingerprintKey = "fingerprint"

// ✅ Good — private key forces use of exported functions
type contextKey string
const fingerprintKey contextKey = "fingerprint"
```

---

## Pattern Evolution

This pattern has been used in:

1. **Request ID** (correlation tracking)
   - `middleware.WithRequestID()`
   - `middleware.RequestID()`

2. **GeoIP enrichment** (location data)
   - `geoip.WithGeoIP()`
   - `geoip.FromContext()`

3. **Fingerprinting** (TLS + HTTP analysis)
   - `fingerprint.WithResult()`
   - `fingerprint.GetResult()`

Each implementation follows the same structure:
1. Private context key type
2. Exported `With*` function to store
3. Exported `Get*`/`From*` function to retrieve
4. Zero-value fallback on missing data

---

## References

- [Go Blog: Context](https://go.dev/blog/context)
- [Effective Go: Context](https://go.dev/doc/effective_go#contexts)

**Implemented in:**
- `internal/middleware/correlation.go` (request ID)
- `internal/geoip/context.go` (geolocation)
- `internal/fingerprint/fingerprint.go` (TLS/HTTP fingerprints)
