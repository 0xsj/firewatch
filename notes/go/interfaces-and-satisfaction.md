# Interfaces and Satisfaction

## Implicit Satisfaction

Go interfaces are satisfied implicitly — no `implements` keyword.
If a type has all the methods an interface requires, it satisfies it.

```go
// Interface definition
type Store interface {
    SaveEvent(ctx context.Context, event *models.Event) error
    GetEvent(ctx context.Context, id string) (*models.Event, error)
    Close() error
}

// SQLiteStore satisfies Store because it has all the methods.
// No declaration needed.
type SQLiteStore struct { db *sql.DB }

func (s *SQLiteStore) SaveEvent(ctx context.Context, event *models.Event) error { ... }
func (s *SQLiteStore) GetEvent(ctx context.Context, id string) (*models.Event, error) { ... }
func (s *SQLiteStore) Close() error { ... }
```

This means you can define an interface in one package and satisfy it
from another without creating a dependency between them.

**Used in:** `internal/storage/storage.go` (interface), `sqlite.go` (implementation)

---

## The Small Interface Pattern

Define small, focused interfaces. The `scanner` interface in our
SQLite code abstracts over two standard library types:

```go
type scanner interface {
    Scan(dest ...any) error
}
```

Both `*sql.Row` and `*sql.Rows` have a `Scan` method, so both
satisfy this interface. This lets us write one `scanEvent` function
that works with both:

```go
// Used with QueryRow (single result)
row := db.QueryRowContext(ctx, query, id)
return scanEvent(row)     // row is *sql.Row

// Used with Query (multiple results)
rows, _ := db.QueryContext(ctx, query, args...)
for rows.Next() {
    e, _ := scanEvent(rows)  // rows is *sql.Rows
}
```

Without this interface, we'd need duplicate scanning logic for
single-row and multi-row queries.

**Used in:** `internal/storage/sqlite.go`

---

## Blank Identifier Imports

```go
import _ "modernc.org/sqlite"
```

The `_` means "import for side effects only." The package's `init()`
function runs, which registers the SQLite driver with `database/sql`:

```go
// Inside modernc.org/sqlite (simplified):
func init() {
    sql.Register("sqlite", &driver{})
}
```

After this import, `sql.Open("sqlite", path)` works. Without it,
you'd get: `unknown driver "sqlite"`.

This pattern is used by all `database/sql` drivers — the driver
name string (`"sqlite"`, `"postgres"`, `"mysql"`) is registered
at import time.

**Used in:** `internal/storage/sqlite.go`

---

## Interface as Contract

Interfaces define **what** a component does, not **how**. This lets
you swap implementations:

```go
// Production
var store storage.Store = storage.NewSQLite("./firewatch.db")

// Testing
var store storage.Store = storage.NewMockStore()

// Future
var store storage.Store = storage.NewPostgres(connString)
```

All three satisfy `Store`. The rest of the code only sees the
interface — it doesn't know or care which database is behind it.

### Accept interfaces, return structs

A common Go proverb. Functions should accept interface parameters
(flexibility) but return concrete types (clarity):

```go
// Good — caller gets the concrete type with all its methods
func NewSQLite(path string) (*SQLiteStore, error)

// The caller can assign it to the interface when needed
var store Store = must(NewSQLite("./data.db"))
```

**Used in:** `internal/storage/`

---

## Compile-time Interface Check

You can verify a type satisfies an interface at compile time:

```go
var _ Store = (*SQLiteStore)(nil)
```

This creates a nil pointer of type `*SQLiteStore` and assigns it to
a `Store` variable. If `SQLiteStore` is missing any methods, the
compiler catches it immediately. The `_` discards the value — this
line generates zero runtime code.

We don't currently use this, but it's a useful safety net when
interfaces and implementations are in different files.
