# Strategy Pattern and Module Registry

## Strategy Pattern

The strategy pattern defines a family of algorithms (honeypot modules),
encapsulates each one behind a common interface, and makes them
interchangeable at runtime.

```go
type Module interface {
    Name() string
    Routes() []Route
}
```

Each honeypot module (`nextjs`, `wordpress`, `exposure`, etc.) implements
this interface. The server doesn't know or care which modules are active —
it works with the interface:

```go
for _, mod := range registry.Enabled(cfg.EnabledModules()) {
    for _, route := range mod.Routes() {
        router.Handle(route.Pattern, route.Handler)
    }
}
```

### Why not just use http.Handler?

The `Module` interface adds **identity** (`Name()`) and **route
declaration** (`Routes()`). A plain `http.Handler` can serve requests
but can't tell you what it handles or what to call it.

This matters for:
- Config: enabling/disabling modules by name
- Logging: tagging events with the module that caught them
- Routing: each module declares its own patterns
- Metrics: counting events per module

**Used in:** `internal/handlers/handler.go`, `internal/handlers/nextjs/`

---

## Registry Pattern

The registry holds all available modules and provides lookup:

```go
type Registry struct {
    modules map[string]Module
}
```

### Registration

Modules register themselves (typically at startup):

```go
reg := handlers.NewRegistry()
reg.Register(nextjs.New(cfg.Modules.NextJS, store, logger))
reg.Register(wordpress.New(cfg.Modules.WordPress, store, logger))
```

The registry panics on duplicate names — this is intentional.
A duplicate means a programming error (two modules claiming the
same name), and failing loudly at startup is better than silent
conflicts at runtime.

### Filtering

`Enabled(names)` returns only the modules the user asked for:

```go
// Config says: modules: [nextjs, exposure]
enabled := reg.Enabled(cfg.EnabledModules())
// Returns [NextJS, Exposure] modules — WordPress is skipped
```

Unknown names are silently skipped rather than erroring. This lets
users keep module names in config even if the binary was built
without that module.

**Used in:** `internal/handlers/registry.go`

---

## Route Declaration

Each module declares its routes as data, not registration calls:

```go
func (n *NextJS) Routes() []handlers.Route {
    return []handlers.Route{
        {Pattern: "POST /", Handler: n.handleServerAction},
        {Pattern: "GET /_next/static/", Handler: n.handleStatic},
        {Pattern: "GET /_rsc", Handler: n.handleRSC},
        {Pattern: "GET /", Handler: n.handlePage},
    }
}
```

Benefits:
- **Testable** — you can inspect routes without starting a server
- **Declarative** — all routes visible in one place
- **Order matters** — more specific patterns listed first
- **Decoupled** — the module doesn't import the router package

The server iterates `Routes()` and registers them. The module
never touches the router directly.

### Method values as handlers

`n.handleServerAction` is a **method value** — it binds the method
to the receiver `n`:

```go
// This:
Handler: n.handleServerAction

// Is equivalent to:
Handler: func(w http.ResponseWriter, r *http.Request) {
    n.handleServerAction(w, r)
}
```

The method value captures `n`, so the handler has access to the
module's config, store, and logger.

**Used in:** `internal/handlers/nextjs/nextjs.go`
