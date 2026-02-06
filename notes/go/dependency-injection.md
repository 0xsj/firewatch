# Dependency Injection

## Constructor Injection

Pass dependencies through the constructor, not globals or init():

```go
func New(cfg config.NextJSModuleConfig, store storage.Store, logger *slog.Logger) *NextJS {
    return &NextJS{
        cfg:    cfg,
        store:  store,
        logger: logger.With("module", moduleName),
    }
}
```

The module receives everything it needs. It doesn't reach into
global state, it doesn't call `config.Load()` itself, it doesn't
create its own database connection.

### Benefits

- **Testable** — pass a mock store, a test logger
- **Explicit** — the constructor signature documents all dependencies
- **No hidden coupling** — if a dependency changes, the compiler tells you
- **Composable** — different configurations for different contexts

### Anti-patterns

```go
// Bad — hidden dependency on global
func (n *NextJS) handlePage(w http.ResponseWriter, r *http.Request) {
    db := database.Global()  // where did this come from?
    cfg := config.Get()      // who set this up?
}

// Good — explicit dependency on struct fields
func (n *NextJS) handlePage(w http.ResponseWriter, r *http.Request) {
    n.store.SaveEvent(...)   // injected via constructor
    n.cfg.Endpoints          // injected via constructor
}
```

**Used in:** `internal/handlers/nextjs/nextjs.go`

---

## slog.Logger.With()

`slog.Logger.With()` creates a child logger with persistent fields:

```go
logger: logger.With("module", moduleName)
```

Every log call from this logger automatically includes `module=nextjs`.
No need to repeat it:

```go
// Without With — repetitive
n.logger.Info("probe detected", "module", "nextjs", "path", path)
n.logger.Info("event saved", "module", "nextjs", "id", id)

// With With — module is automatic
n.logger.Info("probe detected", "path", path)
n.logger.Info("event saved", "id", id)
// Output: level=INFO msg="probe detected" module=nextjs path=/_rsc
```

### Logger hierarchy

```go
// Root logger (created at startup)
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

// Server logger (adds component)
serverLogger := logger.With("component", "server")

// Module logger (adds module name)
moduleLogger := logger.With("module", "nextjs")
```

Each layer adds its own context. A log entry from a module handler
carries both the module name and any request-level context.

**Used in:** `internal/handlers/nextjs/nextjs.go`

---

## Interface Parameters

Accept interfaces in constructors, not concrete types:

```go
// Good — accepts the Store interface
func New(cfg config.NextJSModuleConfig, store storage.Store, ...) *NextJS

// Bad — accepts the concrete SQLite type
func New(cfg config.NextJSModuleConfig, store *storage.SQLiteStore, ...) *NextJS
```

The first version works with SQLite, PostgreSQL, or a mock.
The second locks you into SQLite.

This is the "accept interfaces, return structs" principle in action
at the wiring layer.

**Used in:** `internal/handlers/nextjs/nextjs.go`

---

## Method Values

Go methods can be used as values, capturing their receiver:

```go
// n.handleServerAction is a method value
// It captures `n` and satisfies func(http.ResponseWriter, *http.Request)
{Pattern: "POST /", Handler: n.handleServerAction}
```

This is different from a **method expression** which doesn't bind
a receiver:

```go
// Method value — bound to n
f := n.handleServerAction
f(w, r)  // calls n.handleServerAction(w, r)

// Method expression — unbound, needs receiver as first arg
f := (*NextJS).handleServerAction
f(n, w, r)  // explicit receiver
```

Method values are the natural way to wire handlers in Go. The
receiver provides access to the module's dependencies (store, config,
logger) without closures or globals.

**Used in:** `internal/handlers/nextjs/nextjs.go` — `Routes()` method
