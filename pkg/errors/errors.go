package errors

import (
	stderrors "errors"
	"fmt"
	"strings"
)

// Op describes the operation being performed, typically
// the method or function name. Example: "server.Start",
// "handler.NextJS.ServeHTTP", "storage.SaveEvent".
type Op string

// Error is the core error type for Firewatch. It carries structured
// context through the error chain: which operation failed, what kind
// of failure it was, a machine-readable code, and a stack trace.
type Error struct {
	Op      Op
	Kind    Kind
	Code    Code
	Message string
	Err     error
	stack   *stack
}

// E builds an Error from variadic arguments. Accepted types:
//
//   - Op:     sets the operation name
//   - Kind:   sets the error kind (broad category)
//   - Code:   sets the error code (specific identifier)
//   - string: sets the human-readable message
//   - *Error: sets as wrapped error, inherits Kind if unset
//   - error:  sets as wrapped error
//
// Example:
//
//	errors.E(errors.Op("server.Start"), errors.KindInternal, "failed to bind port")
//	errors.E(errors.Op("handler.NextJS"), errors.KindNotFound, err)
func E(args ...any) *Error {
	e := &Error{}
	for _, arg := range args {
		switch a := arg.(type) {
		case Op:
			e.Op = a
		case Kind:
			e.Kind = a
		case Code:
			e.Code = a
		case string:
			e.Message = a
		case *Error:
			if e.Kind == 0 {
				e.Kind = a.Kind
			}
			e.Err = a
		case error:
			e.Err = a
		}
	}
	e.stack = capture(3)
	return e
}

// New creates an error with a message.
func New(message string) *Error {
	return &Error{
		Message: message,
		stack:   capture(3),
	}
}

// Newf creates an error with a formatted message.
func Newf(format string, args ...any) *Error {
	return &Error{
		Message: fmt.Sprintf(format, args...),
		stack:   capture(3),
	}
}

// Wrap wraps an existing error with an operation name for context.
func Wrap(err error, op Op) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		Op:    op,
		Err:   err,
		stack: capture(3),
	}
}

// Wrapf wraps an existing error with a formatted message.
func Wrapf(err error, format string, args ...any) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		Message: fmt.Sprintf(format, args...),
		Err:     err,
		stack:   capture(3),
	}
}

// Error returns a human-readable string for the error chain.
// Format: "op: kind: [code] message" — fields are omitted when empty.
func (e *Error) Error() string {
	var b strings.Builder

	if e.Op != "" {
		b.WriteString(string(e.Op))
		b.WriteString(": ")
	}

	if e.Kind != 0 {
		b.WriteString(e.Kind.String())
		if e.Message != "" || e.Code != "" || e.Err != nil {
			b.WriteString(": ")
		}
	}

	if e.Code != "" {
		b.WriteString("[")
		b.WriteString(string(e.Code))
		b.WriteString("] ")
	}

	if e.Message != "" {
		b.WriteString(e.Message)
	} else if e.Err != nil {
		b.WriteString(e.Err.Error())
	}

	return b.String()
}

// Unwrap returns the wrapped error, supporting the standard errors chain.
func (e *Error) Unwrap() error {
	return e.Err
}

// StackTrace returns the call stack captured when the error was created.
// Returns nil if no stack was captured.
func (e *Error) StackTrace() []Frame {
	if e.stack == nil {
		return nil
	}
	return e.stack.frames()
}

// GetKind walks the error chain and returns the first non-zero Kind.
func GetKind(err error) Kind {
	var e *Error
	for stderrors.As(err, &e) {
		if e.Kind != 0 {
			return e.Kind
		}
		err = e.Err
	}
	return 0
}

// GetCode walks the error chain and returns the first non-empty Code.
func GetCode(err error) Code {
	var e *Error
	for stderrors.As(err, &e) {
		if e.Code != "" {
			return e.Code
		}
		err = e.Err
	}
	return ""
}

// GetOp returns the outermost Op in the error chain.
func GetOp(err error) Op {
	var e *Error
	if stderrors.As(err, &e) {
		if e.Op != "" {
			return e.Op
		}
	}
	return ""
}

// Ops collects all operations from the error chain, outermost first.
// This gives a call path like: ["server.Start", "config.Load", "config.Parse"].
func Ops(err error) []Op {
	var ops []Op
	var e *Error
	for stderrors.As(err, &e) {
		if e.Op != "" {
			ops = append(ops, e.Op)
		}
		err = e.Err
	}
	return ops
}

// HTTPStatus maps an error to an HTTP status code through its Kind.
// Returns 500 if no Kind is found.
func HTTPStatus(err error) int {
	kind := GetKind(err)
	if kind == 0 {
		return 500
	}
	return kind.HTTPStatus()
}

// Re-export standard library functions so callers only need one errors import.

// Is reports whether any error in err's chain matches target.
func Is(err, target error) bool {
	return stderrors.Is(err, target)
}

// As finds the first error in err's chain that matches target.
func As(err error, target any) bool {
	return stderrors.As(err, target)
}

// Unwrap returns the result of calling Unwrap on err.
func Unwrap(err error) error {
	return stderrors.Unwrap(err)
}

// Join returns an error that wraps the given errors.
func Join(errs ...error) error {
	return stderrors.Join(errs...)
}
