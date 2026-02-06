# Middleware Chain Pattern

## What Middleware Is

Middleware wraps an HTTP handler to add behavior before and/or after
the inner handler runs:

```go
type Middleware func(http.Handler) http.Handler
```

It takes a handler, returns a new handler. The new handler can:
- Inspect/modify the request before passing it along
- Inspect/modify the response after the inner handler writes it
- Short-circuit (return early without calling the inner handler)
- Add values to the request context

```go
func Logging(logger *slog.Logger) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            next.ServeHTTP(w, r)  // call the inner handler
            logger.Info("request", "duration", time.Since(start))
        })
    }
}
```

**Used in:** `internal/middleware/`

---

## Closures for Configuration

When middleware needs configuration (a logger, a rate limit, etc.),
use a closure — a function that returns a Middleware:

```go
// The outer function takes config, returns a Middleware
func Logging(logger *slog.Logger) Middleware {
    // The Middleware captures `logger` from the enclosing scope
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // `logger` is available here via closure
            logger.Info("request", "path", r.URL.Path)
            next.ServeHTTP(w, r)
        })
    }
}
```

Middleware without config (like `Correlation`) is just a function:

```go
func Correlation(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // ...
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

- `Correlation` has signature `func(http.Handler) http.Handler` → it IS a `Middleware`
- `Logging(logger)` returns `func(http.Handler) http.Handler` → it RETURNS a `Middleware`

**Used in:** `internal/middleware/logging.go`, `correlation.go`

---

## Chain Composition

`Chain(A, B, C)(handler)` produces: `A(B(C(handler)))`.

Execution order: A → B → C → handler → C → B → A

```go
func Chain(mws ...Middleware) Middleware {
    return func(final http.Handler) http.Handler {
        for i := len(mws) - 1; i >= 0; i-- {
            final = mws[i](final)
        }
        return final
    }
}
```

Why iterate in reverse? We're wrapping from the inside out:
1. Start with `final = handler`
2. `final = C(handler)`
3. `final = B(C(handler))`
4. `final = A(B(C(handler)))`

So the first middleware in the list (`A`) is the outermost wrapper —
it runs first on the way in and last on the way out.

```go
chain := middleware.Chain(
    middleware.Correlation,       // 1st: assign request ID
    middleware.Logging(logger),   // 2nd: log with request ID available
)
handler := chain(router)
```

Order matters: `Correlation` must run before `Logging` so the
request ID is available when the log entry is written.

**Used in:** `internal/server/server.go`

---

## Wrapping ResponseWriter

The standard `http.ResponseWriter` doesn't expose the status code
after `WriteHeader` is called. To log the status, we wrap it:

```go
type responseWriter struct {
    http.ResponseWriter  // embed the original
    status  int
    size    int
    written bool
}
```

### Embedding

`http.ResponseWriter` is embedded (no field name). This means:
- All methods of `http.ResponseWriter` are promoted to `responseWriter`
- `responseWriter` automatically satisfies `http.ResponseWriter`
- We only override the methods we care about (`WriteHeader`, `Write`)
- All other methods (like `Header()`) delegate to the original

```go
// Override WriteHeader to capture the status code
func (w *responseWriter) WriteHeader(code int) {
    if !w.written {
        w.status = code
        w.written = true
    }
    w.ResponseWriter.WriteHeader(code)  // call the original
}
```

### Optional interface detection

`http.ResponseWriter` might also implement `http.Flusher` (for
streaming responses). We preserve this with a type assertion:

```go
func (w *responseWriter) Flush() {
    if f, ok := w.ResponseWriter.(http.Flusher); ok {
        f.Flush()
    }
}
```

- `w.ResponseWriter.(http.Flusher)` — type assertion: "does this
  ResponseWriter also implement Flusher?"
- `ok` is `true` if it does, `false` otherwise
- The comma-ok form prevents panics on failed assertions

**Used in:** `internal/middleware/middleware.go`

---

## Context Values for Request-Scoped Data

Middleware passes data to downstream handlers via context:

```go
// Define an unexported key type to avoid collisions
type contextKey string
const requestIDKey contextKey = "request_id"

// Set in middleware
ctx := context.WithValue(r.Context(), requestIDKey, id)
next.ServeHTTP(w, r.WithContext(ctx))

// Read downstream
func RequestID(ctx context.Context) string {
    id, _ := ctx.Value(requestIDKey).(string)
    return id
}
```

Why an unexported type for the key?
- `context.WithValue` uses `==` to match keys
- If the key were just `string("request_id")`, any package could
  accidentally collide with it
- Using `type contextKey string` makes the key type unique to
  this package — only code with access to `contextKey` can match it

The comma-ok type assertion (`id, _ := ...`) returns zero value
if the key isn't in the context, which is safe.

**Used in:** `internal/middleware/correlation.go`
