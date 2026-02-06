# Error Architecture — Kind / Code / Op

## The Problem

In a layered system (HTTP → handler → storage), errors cross boundaries.
Each layer needs different information:

- **HTTP layer**: what status code to return?
- **Logging**: what operation failed, and where in the call chain?
- **Alerting**: how severe is this? Is it the same class of failure?
- **Debugging**: what's the stack trace?

A single `fmt.Errorf("failed: %w", err)` doesn't carry enough structure.

---

## Three Dimensions of an Error

### Kind — "How did it fail?"

Broad category. Maps directly to HTTP status codes and log levels.

```
KindNotFound     → 404, INFO
KindValidation   → 400, WARN
KindInternal     → 500, ERROR
KindTimeout      → 504, WARN
KindRateLimit    → 429, INFO
```

There are ~11 kinds. They rarely change. Every layer understands them.

### Code — "What specifically broke?"

Machine-readable identifier scoped to a domain:

```
config_invalid     — config file is malformed
storage_connect    — can't reach the database
alert_send         — failed to dispatch an alert
fingerprint_ja3    — JA3 extraction failed
```

Codes are more specific than kinds. `CodeStorageConnect` and
`CodeStorageQuery` are both `KindInternal`, but monitoring can
distinguish them.

### Op — "Where in the code?"

The function/method that created or wrapped the error:

```go
errors.E(errors.Op("storage.SaveEvent"), ...)
errors.Wrap(err, errors.Op("handler.NextJS.ServeHTTP"))
errors.Wrap(err, errors.Op("server.Handle"))
```

`Ops(err)` returns the full call path:
```
["server.Handle", "handler.NextJS.ServeHTTP", "storage.SaveEvent"]
```

This is lighter than a full stack trace and gives meaningful
application-level context.

---

## Inspiration: Upspin

This pattern comes from the [Upspin project](https://commandcenter.blogspot.com/2017/12/error-handling-in-upspin.html)
by Rob Pike and the Go team. Key ideas borrowed:

1. **Variadic `E()` constructor** — accepts any combination of Op, Kind,
   Code, string, error. Type switch routes each arg to the right field.

2. **Kind propagation** — if you wrap an error without setting a Kind,
   `GetKind()` walks the chain to find the inner error's Kind. The
   outermost layer doesn't need to know the category — it bubbles up.

3. **Op chain** — each layer adds its own Op when wrapping, building
   a call path without repeating the message.

---

## Usage Pattern

```go
// Bottom of the stack — where the error originates
func (s *Store) SaveEvent(e Event) error {
    if err := s.db.Insert(e); err != nil {
        return errors.E(
            errors.Op("storage.SaveEvent"),
            errors.KindInternal,
            errors.CodeStorageQuery,
            err,
        )
    }
    return nil
}

// Middle — just add context via Op
func (h *NextJSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if err := h.store.SaveEvent(event); err != nil {
        err = errors.Wrap(err, errors.Op("handler.NextJS.ServeHTTP"))
        httputil.Error(w, err)  // Kind → HTTP status automatically
        return
    }
}

// Top — extract what you need
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // ...
        kind := errors.GetKind(err)    // KindInternal
        code := errors.GetCode(err)    // "storage_query"
        ops  := errors.Ops(err)        // ["handler.NextJS...", "storage.SaveEvent"]
        status := errors.HTTPStatus(err) // 500
    })
}
```

---

## pkg/ vs internal/

The `pkg/` directory is **importable by external code**. The `internal/`
directory is **private to this module**.

```
pkg/errors/      ← any Go module can import this
pkg/httputil/    ← reusable HTTP helpers
internal/server/ ← only firewatch can use this
internal/handlers/ ← only firewatch can use this
```

We put `errors`, `httputil`, `crypto`, `netutil`, `timeutil`, and
`validate` in `pkg/` because they have no business logic coupling —
another project could use them as-is.

Everything in `internal/` depends on Firewatch-specific domain concepts
(honeypot modules, detection rules, attacker profiles).
