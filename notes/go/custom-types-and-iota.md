# Custom Types and Iota

## Type Definitions

Go lets you create new types from existing ones. The new type is
distinct at compile time — you can't accidentally mix them up.

```go
type Op string       // operation name
type Code string     // error code
type Kind uint8      // error category
```

Even though `Op` and `Code` are both backed by `string`, they are
**different types**. You can't pass a `Code` where an `Op` is expected
without an explicit conversion.

```go
var o Op = "server.Start"
var c Code = "server_bind"
// o = c   ← compile error: cannot use c (Code) as Op
// o = Op(c) ← explicit conversion, compiles fine
```

**Used in:** `pkg/errors/errors.go` (Op), `pkg/errors/codes.go` (Code), `pkg/errors/kinds.go` (Kind)

---

## Iota

`iota` auto-increments within a `const` block, starting at 0.

```go
const (
    KindUnexpected Kind = iota  // 0
    KindNotFound                // 1
    KindValidation              // 2
    KindUnauthorized            // 3
)
```

The zero value (`KindUnexpected = 0`) is intentional — an unset `Kind`
field defaults to "unexpected", which is the safest fallback.

### Skipping zero

If you want to catch "was this ever set?", skip zero:

```go
const (
    _           Kind = iota  // 0 is unused
    KindNotFound             // 1
    KindValidation           // 2
)
```

Now `Kind(0)` means "not set" rather than a valid value.

**Used in:** `pkg/errors/kinds.go`

---

## Methods on Custom Types

Any named type can have methods attached. This is how `Kind` gets its
`String()` and `HTTPStatus()` methods:

```go
func (k Kind) String() string {
    if int(k) < len(kindNames) {
        return kindNames[k]
    }
    return "unknown"
}
```

The receiver `(k Kind)` is a value receiver — `k` is a copy of the Kind value.
For small types like `uint8` or `string`, value receivers are standard.

### Array-based lookup

Using an array indexed by the Kind value is faster than a switch or map
for `String()`:

```go
var kindNames = [...]string{
    KindUnexpected:   "unexpected",
    KindNotFound:     "not found",
    KindValidation:   "validation",
}
```

The `[...]` syntax means "size determined by the initializer."
`kindNames[KindNotFound]` → `"not found"`.

For methods that return different types (like `HTTPStatus() int`), a switch
is clearer than maintaining parallel arrays.

**Used in:** `pkg/errors/kinds.go`

---

## Zero Value Semantics

Every Go type has a zero value. Design types so the zero value is useful:

| Type       | Zero value | Our meaning          |
|------------|-----------|----------------------|
| `Kind(0)`  | `0`       | `KindUnexpected`     |
| `Code("")` | `""`      | no specific code set |
| `Op("")`   | `""`      | no operation context |
| `*Error`   | `nil`     | no error             |

The `GetKind()` function exploits this — it walks the error chain looking
for `Kind != 0`, meaning "first error that actually classified itself."
