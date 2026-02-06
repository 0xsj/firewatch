# Module Uniformity & Shared Helpers

## The Problem

With 6 honeypot modules (nextjs, wordpress, exposure, api, cloud, + future ones), repeating the same event construction logic in each handler creates:
- Duplicated boilerplate across 30+ handlers
- Inconsistent field population if one handler forgets a field
- Shotgun surgery when the event structure changes

## The Shared Helper Pattern

A single `RecordEvent` function in the `handlers` package that all modules call:

```go
// internal/handlers/event.go
package handlers

func RecordEvent(store storage.Store, logger *slog.Logger, r *http.Request,
    module, severity string, signatures []string) {

    event := &models.Event{
        ID:         crypto.UUID4(),
        Timestamp:  timeutil.FormatRFC3339(timeutil.NowUTC()),
        RequestID:  middleware.RequestID(r.Context()),
        SourceIP:   httputil.ClientIP(r),
        Module:     module,
        Method:     r.Method,
        Path:       r.URL.Path,
        Query:      r.URL.RawQuery,
        Headers:    httputil.HeaderMap(r.Header),
        UserAgent:  r.UserAgent(),
        Severity:   severity,
        Signatures: signatures,
    }

    if err := store.SaveEvent(r.Context(), event); err != nil {
        logger.Error("failed to save event", ...)
    }
}
```

## Why it works

**Module-specific**: Only `module`, `severity`, and `signatures` vary per handler — the rest is extracted from the request automatically.

**Dependency passing**: Store and logger are passed as arguments (not globals), preserving testability. Each module already holds these from construction.

**Error contained**: Save failure is logged but doesn't crash the handler. The response still goes to the attacker.

## Module Structure Convention

Every module follows the exact same shape:

```
module/
├── module.go       # Struct, New(), Name(), Routes()
├── handler1.go     # Handler + signature consts
├── handler2.go     # Handler + signature consts
└── handler3.go     # Handler + signature consts
```

### module.go template

```go
package mymodule

const moduleName = "mymodule"

type MyModule struct {
    cfg    config.MyModuleConfig
    store  storage.Store
    logger *slog.Logger
}

func New(cfg config.MyModuleConfig, store storage.Store, logger *slog.Logger) *MyModule {
    return &MyModule{
        cfg:    cfg,
        store:  store,
        logger: logger.With("module", moduleName),
    }
}

func (m *MyModule) Name() string { return moduleName }

func (m *MyModule) Routes() []handlers.Route {
    return []handlers.Route{
        {Pattern: "GET /path", Handler: m.handleSomething},
    }
}
```

### handler.go template

```go
const sigMyProbe = "mymodule-probe-001"

func (m *MyModule) handleSomething(w http.ResponseWriter, r *http.Request) {
    // 1. Log the probe
    m.logger.Info("description", "ip", httputil.ClientIP(r))

    // 2. Record the event
    handlers.RecordEvent(m.store, m.logger, r, moduleName, "medium", []string{sigMyProbe})

    // 3. Return deceptive response
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("fake content"))
}
```

## Signature Naming Convention

Format: `<module>-<type>-<sequence>`

| Module    | Examples                                    |
|-----------|---------------------------------------------|
| nextjs    | nextjs-probe-001, nextjs-action-001         |
| wordpress | wp-login-001, wp-bruteforce-001, wp-xmlrpc-001 |
| exposure  | exposure-env-001, exposure-git-001          |
| api       | api-rest-001, api-graphql-001, api-swagger-001 |
| cloud     | cloud-metadata-001, cloud-iam-001           |

Sequence numbers within a module track escalation:
- `001` = basic probe/presence
- `002` = deeper interaction (payload, introspection, auth attempt)
- `003` = specialized sub-behavior
