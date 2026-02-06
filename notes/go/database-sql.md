# database/sql

## Overview

`database/sql` is Go's standard database interface. It provides a
common API — the actual database driver is plugged in via imports.

```go
import (
    "database/sql"
    _ "modernc.org/sqlite"  // registers "sqlite" driver
)

db, err := sql.Open("sqlite", "./firewatch.db")
```

`sql.Open` doesn't actually connect — it validates the driver name
and returns a `*sql.DB` pool. The first real connection happens on
the first query.

---

## The Three Query Methods

### Exec — for writes (INSERT, UPDATE, DELETE, DDL)

```go
result, err := db.ExecContext(ctx, "INSERT INTO events (id, module) VALUES (?, ?)", id, module)

// result.LastInsertId() — last auto-increment ID
// result.RowsAffected() — number of rows changed
```

### QueryRow — for single-row reads

```go
row := db.QueryRowContext(ctx, "SELECT id, module FROM events WHERE id = ?", id)

var id, module string
err := row.Scan(&id, &module)
if err == sql.ErrNoRows {
    // not found
}
```

`Scan` copies column values into the provided pointers. Column order
must match the SELECT order exactly.

### Query — for multi-row reads

```go
rows, err := db.QueryContext(ctx, "SELECT id, module FROM events ORDER BY timestamp DESC")
if err != nil { ... }
defer rows.Close()  // ALWAYS close rows

for rows.Next() {
    var id, module string
    err := rows.Scan(&id, &module)
    // ...
}
// Check for errors from iteration
if err := rows.Err(); err != nil { ... }
```

Key rules:
- **Always `defer rows.Close()`** — leaking rows exhausts the connection pool
- **Always check `rows.Err()`** after the loop — iteration can fail mid-stream
- `rows.Next()` returns false on both completion and error

**Used in:** `internal/storage/sqlite.go`

---

## Parameter Placeholders

```go
// SQLite and MySQL use ?
db.Query("SELECT * FROM events WHERE ip = ? AND module = ?", ip, module)

// PostgreSQL uses $1, $2, ...
db.Query("SELECT * FROM events WHERE ip = $1 AND module = $2", ip, module)
```

**Never** concatenate user input into SQL strings:
```go
// WRONG — SQL injection
db.Query("SELECT * FROM events WHERE ip = '" + ip + "'")

// RIGHT — parameterized
db.Query("SELECT * FROM events WHERE ip = ?", ip)
```

The driver escapes parameters automatically.

**Used in:** `internal/storage/sqlite.go` — all queries

---

## sql.NullString

SQL NULL and Go zero values don't map cleanly. A TEXT column can be:
- `"hello"` — a string
- `""` — an empty string
- `NULL` — no value at all

Go's `string` can't distinguish empty from NULL. `sql.NullString` can:

```go
var raw sql.NullString
row.Scan(&raw)

if raw.Valid {
    // raw.String has the value
} else {
    // column was NULL
}
```

There are also `sql.NullInt64`, `sql.NullFloat64`, `sql.NullBool`,
and `sql.NullTime`.

We use `NullString` for JSON columns that might be NULL (no data
stored yet):

```go
var headers sql.NullString
row.Scan(&headers)
if headers.Valid {
    json.Unmarshal([]byte(headers.String), &event.Headers)
}
```

**Used in:** `internal/storage/sqlite.go` — all scan functions

---

## Context Variants

Every query method has a `Context` variant:

| Without context     | With context               |
|--------------------|---------------------------|
| `db.Exec()`        | `db.ExecContext(ctx)`      |
| `db.Query()`       | `db.QueryContext(ctx)`     |
| `db.QueryRow()`    | `db.QueryRowContext(ctx)`  |

**Always use the Context variants.** They let you:
- Set deadlines (query timeout)
- Cancel long-running queries
- Propagate request-scoped cancellation

```go
func (s *SQLiteStore) GetEvent(ctx context.Context, id string) (*models.Event, error) {
    row := s.db.QueryRowContext(ctx, "SELECT ... WHERE id = ?", id)
    // If ctx is canceled, the query is interrupted
}
```

**Used in:** `internal/storage/sqlite.go` — every method

---

## Pragmas (SQLite-specific)

Pragmas are SQLite configuration commands:

```go
db.Exec("PRAGMA journal_mode=WAL")
```

### WAL (Write-Ahead Logging)

Default SQLite uses rollback journaling — only one writer OR reader
at a time. WAL mode allows:
- **Concurrent readers** while writing
- **Better write performance** (appends vs rewrites)
- Slight increase in disk usage (WAL file alongside the database)

For a honeypot capturing many concurrent requests, WAL is essential.

Other useful pragmas:
```sql
PRAGMA foreign_keys = ON;     -- enforce foreign key constraints
PRAGMA busy_timeout = 5000;   -- wait 5s instead of failing on lock
PRAGMA synchronous = NORMAL;  -- faster writes, still crash-safe with WAL
```

**Used in:** `internal/storage/sqlite.go` — `NewSQLite()`

---

## UPSERT Pattern

"Insert or update" in one statement:

```sql
INSERT INTO attackers (id, ip, total_events, ...)
VALUES (?, ?, ?, ...)
ON CONFLICT(ip) DO UPDATE SET
    total_events = excluded.total_events,
    last_seen = excluded.last_seen
```

- `ON CONFLICT(ip)` — triggers when the UNIQUE constraint on `ip` is violated
- `excluded.*` — refers to the values that would have been inserted
- Only the columns listed in `SET` are updated

This is atomic — no race condition between checking existence and
inserting/updating. Perfect for updating attacker profiles as new
events arrive.

**Used in:** `internal/storage/sqlite.go` — `SaveAttacker()`, `SaveCampaign()`, `SaveIOC()`
