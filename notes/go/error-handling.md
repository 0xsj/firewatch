# Error Handling in Go

## The error Interface

Go's `error` is a single-method interface:

```go
type error interface {
    Error() string
}
```

Any type with an `Error() string` method satisfies it. Our `*Error`
struct implements this.

---

## Wrapping and Unwrapping

Go 1.13+ introduced error chains via `Unwrap()`:

```go
func (e *Error) Unwrap() error {
    return e.Err  // the wrapped inner error
}
```

This lets the standard library walk the chain:

```go
errors.Is(err, target)   // does any error in the chain == target?
errors.As(err, &target)  // extract a specific type from the chain
errors.Unwrap(err)       // get the next error in the chain
```

### fmt.Errorf with %w

The simplest way to wrap:

```go
return fmt.Errorf("loading config: %w", err)
```

`%w` wraps the error so `Unwrap()` works. `%v` just formats the string
(no chain).

**Used in:** `pkg/errors/errors.go`

---

## Type Switch

A type switch dispatches on the concrete type of an interface value:

```go
for _, arg := range args {
    switch a := arg.(type) {
    case Op:
        e.Op = a        // a is typed as Op
    case Kind:
        e.Kind = a      // a is typed as Kind
    case string:
        e.Message = a   // a is typed as string
    case error:
        e.Err = a       // a is typed as error
    }
}
```

Key points:
- `arg.(type)` can only appear in a switch statement
- `a` gets the concrete type in each case branch
- Order matters — put specific types before general ones
  (`*Error` before `error`, since `*Error` satisfies `error`)

**Used in:** `pkg/errors/errors.go` — the `E()` variadic constructor

---

## Variadic Functions

`func E(args ...any) *Error` accepts any number of arguments of any type.

Inside the function, `args` is `[]any`. The type switch above inspects
each argument and routes it to the right field.

This gives an ergonomic API:

```go
// All of these are valid:
errors.E(errors.Op("server.Start"), errors.KindInternal, "bind failed")
errors.E(errors.KindNotFound, "user not found")
errors.E(errors.Op("handler.NextJS"), err)
```

Trade-off: you lose compile-time type checking. Passing an `int` would
silently be ignored. This is an intentional trade in the Upspin pattern —
ergonomics over strict safety for error construction.

---

## Walking the Error Chain

`errors.As` finds the first error of a given type in the chain:

```go
func GetKind(err error) Kind {
    var e *Error
    for stderrors.As(err, &e) {
        if e.Kind != 0 {
            return e.Kind
        }
        err = e.Err  // keep walking
    }
    return 0
}
```

The `for` loop pattern walks deeper into the chain on each iteration.
`errors.As` sets `e` to the next `*Error` it finds and returns true,
or returns false when the chain is exhausted.

**Used in:** `pkg/errors/errors.go` — `GetKind()`, `GetCode()`, `GetOp()`, `Ops()`

---

## Nil Guards on Wrap

Always guard against wrapping nil errors:

```go
func Wrap(err error, op Op) *Error {
    if err == nil {
        return nil
    }
    // ...
}
```

This lets callers write `return errors.Wrap(maybeNil, op)` without
checking first.

---

## Re-exporting stdlib

Since our package is named `errors`, it shadows the standard `errors`
package. We re-export the stdlib functions so callers only need one import:

```go
import "github.com/0xsj/firewatch/pkg/errors"

// These all work — no need to also import "errors":
errors.Is(err, target)
errors.As(err, &target)
errors.Unwrap(err)
errors.Join(err1, err2)
```

**Used in:** `pkg/errors/errors.go` — bottom of file
