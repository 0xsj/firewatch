# Repository Pattern

## What It Is

The repository pattern abstracts data access behind an interface.
Business logic talks to the interface, not the database. This
decouples storage decisions from application logic.

```
Handler → Store interface → SQLiteStore
                          → PostgresStore (future)
                          → MockStore (tests)
```

---

## Interface Design

```go
type Store interface {
    SaveEvent(ctx context.Context, event *models.Event) error
    GetEvent(ctx context.Context, id string) (*models.Event, error)
    ListEvents(ctx context.Context, f EventFilter) ([]*models.Event, error)
    CountEvents(ctx context.Context, f EventFilter) (int64, error)
    Close() error
}
```

Design decisions:
- **`context.Context` first** — Go convention for cancellation/timeouts
- **Pointer receivers for models** — avoids copying large structs
- **Return slices of pointers** (`[]*Event`) — nil slice means "no results"
- **Separate filter structs** — keeps method signatures clean
- **`Close()`** — databases hold resources, callers control cleanup

---

## Filter Structs

Instead of methods with many parameters:

```go
// Bad — hard to read, easy to mix up arguments
ListEvents(ctx, since, until, "", "nextjs", "high", 50, 0)
```

Use a struct with named fields:

```go
// Good — self-documenting, zero values mean "no filter"
ListEvents(ctx, EventFilter{
    Module:   "nextjs",
    Severity: "high",
    Limit:    50,
})
```

Zero values act as "unset":
- `time.Time{}` (zero) → no time filter
- `""` → no string filter
- `0` → no limit

This is why the query builder checks each field:

```go
if f.Module != "" {
    where = append(where, "module = ?")
    args = append(args, f.Module)
}
```

---

## Dynamic Query Building

Building SQL WHERE clauses from filters:

```go
func buildEventQuery(base string, f EventFilter) (string, []any) {
    var where []string
    var args []any

    if f.SourceIP != "" {
        where = append(where, "source_ip = ?")
        args = append(args, f.SourceIP)
    }
    if f.Module != "" {
        where = append(where, "module = ?")
        args = append(args, f.Module)
    }

    query := base
    if len(where) > 0 {
        query += " WHERE " + strings.Join(where, " AND ")
    }
    return query, args
}
```

Key points:
- `where` and `args` grow together — index N in `where` corresponds to
  index N in `args`
- `strings.Join(where, " AND ")` composes the conditions
- The `base` parameter lets the same builder work for both
  `SELECT *` (ListEvents) and `SELECT COUNT(*)` (CountEvents)
- Parameters use `?` placeholders, never string concatenation

---

## Error Wrapping in Storage

Every storage method wraps errors with context:

```go
func (s *SQLiteStore) SaveEvent(ctx context.Context, event *models.Event) error {
    _, err := s.db.ExecContext(ctx, query, args...)
    if err != nil {
        return errors.E(
            errors.Op("storage.sqlite.SaveEvent"),
            errors.KindInternal,
            errors.CodeStorageQuery,
            err,
        )
    }
    return nil
}
```

This means any caller can:
- `errors.GetOp(err)` → `"storage.sqlite.SaveEvent"` (where)
- `errors.GetKind(err)` → `KindInternal` (how)
- `errors.GetCode(err)` → `"storage_query"` (what)
- `errors.Unwrap(err)` → original SQLite error (why)

Special case — `sql.ErrNoRows` maps to `KindNotFound`:

```go
if err == sql.ErrNoRows {
    return nil, errors.E(op, errors.KindNotFound, "event not found")
}
```

---

## JSON Columns for Complex Types

Relational columns for scalar fields, JSON text for complex ones:

| Field type          | Column type | Example                    |
|---------------------|-------------|----------------------------|
| `string`            | `TEXT`      | `source_ip`, `module`      |
| `int`               | `INTEGER`   | `source_port`              |
| `map[string]string` | `TEXT`(JSON) | `headers`                 |
| `[]string`          | `TEXT`(JSON) | `signatures`, `tags`      |
| `struct`            | `TEXT`(JSON) | `fingerprint`, `geoip`    |

Trade-offs:
- **Pro**: Simple schema, no join tables for arrays
- **Pro**: SQLite's JSON functions can query into JSON columns
- **Con**: Can't index individual JSON values efficiently
- **Con**: Loses type safety at the database level

For a honeypot where most queries filter on scalar fields (IP, module,
severity, timestamp) and JSON data is read as a whole, this trade-off
makes sense.
