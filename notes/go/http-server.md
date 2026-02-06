# HTTP Server Fundamentals

## The Handler Interface

Everything in `net/http` revolves around one interface:

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

Any type with a `ServeHTTP` method can handle HTTP requests. The
router, middleware, and honeypot modules all satisfy this interface.

---

## HandlerFunc Adapter

Writing a full struct for every handler is verbose. `http.HandlerFunc`
is a type that lets a plain function satisfy `Handler`:

```go
type HandlerFunc func(ResponseWriter, *Request)

func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
    f(w, r)
}
```

Usage:

```go
// These are equivalent:
mux.Handle("/path", http.HandlerFunc(myFunc))
mux.HandleFunc("/path", myFunc)
```

`HandlerFunc` is an **adapter** — it adds a method to a function type
so it satisfies an interface. This is a common Go pattern.

**Used in:** `internal/middleware/` — every middleware returns `http.HandlerFunc`

---

## http.Server

The standard HTTP server struct:

```go
s := &http.Server{
    Addr:      ":8080",
    Handler:   myHandler,
    TLSConfig: tlsCfg,
}

// Start listening (blocks until error or shutdown)
err := s.ListenAndServe()
err := s.ListenAndServeTLS(certFile, keyFile)
```

- `Addr` — the address to listen on (`:8080` means all interfaces, port 8080)
- `Handler` — the root handler; every request flows through this
- Returns `http.ErrServerClosed` after a graceful `Shutdown()` — this
  is expected, not an error:

```go
err := s.ListenAndServe()
if err != nil && err != http.ErrServerClosed {
    // actual error
}
```

**Used in:** `internal/server/server.go`

---

## ServeMux (Go 1.22+)

`http.ServeMux` is the standard request router. Go 1.22 enhanced it
with method and wildcard support:

```go
mux := http.NewServeMux()

// Exact match
mux.Handle("/wp-login.php", loginHandler)

// Prefix match (trailing slash)
mux.Handle("/_next/", nextjsHandler)

// Method-specific (Go 1.22+)
mux.Handle("POST /api/login", apiLoginHandler)

// Wildcards (Go 1.22+)
mux.Handle("GET /api/users/{id}", userHandler)
```

### Pattern precedence

More specific patterns win:
- `/api/users/123` beats `/api/users/{id}` beats `/api/`
- `POST /login` beats `/login` (method-specific beats general)

### Detecting unmatched routes

```go
handler, pattern := mux.Handler(req)
if pattern == "" {
    // No route matched — use fallback
}
```

This is how our Router sends unmatched requests to the fallback
handler instead of returning the default 404.

**Used in:** `internal/server/router.go`

---

## StripPrefix

Removes a path prefix before forwarding to a handler:

```go
router.Handle("/wp-admin/",
    http.StripPrefix("/wp-admin", wordpressHandler))
```

A request for `/wp-admin/edit.php` reaches `wordpressHandler` as
`/edit.php`. This is how `Mount()` works — modules see paths relative
to their prefix.

**Used in:** `internal/server/router.go` — `Mount()`

---

## ResponseWriter

The interface for writing HTTP responses:

```go
type ResponseWriter interface {
    Header() http.Header       // get response headers (before WriteHeader)
    WriteHeader(statusCode int) // send status code (can only call once)
    Write([]byte) (int, error)  // write body bytes
}
```

Ordering rules:
1. Set headers with `w.Header().Set(...)` **before** calling `WriteHeader`
2. Call `WriteHeader(code)` **before** calling `Write`
3. If you call `Write` without calling `WriteHeader`, it defaults to 200

Once `WriteHeader` is called, headers are frozen — further `Header().Set()`
calls have no effect on the response.

**Used in:** `pkg/httputil/response.go`, `internal/middleware/middleware.go`

---

## Request.WithContext

Creates a shallow copy of the request with a new context:

```go
ctx := context.WithValue(r.Context(), key, value)
next.ServeHTTP(w, r.WithContext(ctx))
```

This is how middleware passes data downstream (e.g., request IDs).
The original request is not modified — `WithContext` returns a new
`*Request` that shares the same body and headers.

**Used in:** `internal/middleware/correlation.go`
