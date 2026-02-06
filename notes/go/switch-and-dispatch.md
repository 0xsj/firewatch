# Switch Statements & Dispatch Patterns

## Basic switch

Go's `switch` is cleaner than C-style — no fallthrough by default, no `break` needed:

```go
switch r.URL.Path {
case "/.git/config":
    // handle config
case "/.git/HEAD":
    // handle HEAD
default:
    // handle everything else
}
```

### Key differences from C/Java

| Feature          | Go                          | C/Java                  |
|------------------|-----------------------------|-------------------------|
| Fallthrough      | Explicit (`fallthrough`)    | Default behavior        |
| Break            | Implicit                    | Required                |
| Expression       | Optional (bare switch)      | Required                |
| Type             | Any comparable type         | int/string/enum         |
| Multiple values  | `case "a", "b":`            | Separate case lines     |

## Bare switch (no expression)

Acts like if/else chain but reads cleaner:

```go
switch {
case r.Method == http.MethodPost && len(body) > 0:
    severity = "critical"
case r.Method == http.MethodPost:
    severity = "high"
default:
    severity = "medium"
}
```

## Switch as dispatch

In the git handler, switch dispatches on URL path to return different fake content. This is a common Go pattern — simpler than a map when you need different logic per case, not just different data.

```go
func (e *Exposure) handleGit(w http.ResponseWriter, r *http.Request) {
    sigs := []string{sigGitProbe}

    switch r.URL.Path {
    case "/.git/config":
        sigs = append(sigs, sigGitConfig)
        w.Write([]byte("[core]\n\trepositoryformatversion = 0\n..."))

    case "/.git/HEAD":
        sigs = append(sigs, sigGitHEAD)
        w.Write([]byte("ref: refs/heads/main\n"))

    default:
        http.Error(w, "403 Forbidden", http.StatusForbidden)
    }

    // Common code after switch — runs for all cases
    handlers.RecordEvent(e.store, e.logger, r, moduleName, severity, sigs)
}
```

### Pattern: Pre-declare, switch-mutate, then act

```go
sigs := []string{baseSignature}  // 1. Start with base
severity := "medium"

switch {                          // 2. Mutate based on conditions
case condition1:
    sigs = append(sigs, extraSig)
    severity = "high"
case condition2:
    severity = "critical"
}

RecordEvent(store, logger, r, module, severity, sigs)  // 3. Act on result
```

This "accumulate then act" pattern appears throughout Firewatch's handlers.

## Type switch

Used earlier in `pkg/errors/E()` — switches on the runtime type of an `interface{}`:

```go
switch v := arg.(type) {
case Op:
    e.Op = v
case Kind:
    e.Kind = v
case string:
    e.Message = v
case error:
    e.Err = v
}
```

`arg.(type)` is only valid inside a switch. For single-type checks, use type assertion: `v, ok := arg.(Op)`.
