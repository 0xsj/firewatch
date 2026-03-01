package errors

import (
	stderrors "errors"
	"net/http"
	"strings"
	"testing"
)

func TestE_Basic(t *testing.T) {
	err := E(Op("server.Start"), KindInternal, "bind failed")

	if err.Op != "server.Start" {
		t.Errorf("Op = %q, want server.Start", err.Op)
	}
	if err.Kind != KindInternal {
		t.Errorf("Kind = %v, want KindInternal", err.Kind)
	}
	if err.Message != "bind failed" {
		t.Errorf("Message = %q, want bind failed", err.Message)
	}
}

func TestE_WithCode(t *testing.T) {
	err := E(Op("config.Load"), KindValidation, CodeConfigInvalid, "port out of range")

	if err.Code != CodeConfigInvalid {
		t.Errorf("Code = %q, want %q", err.Code, CodeConfigInvalid)
	}
}

func TestE_WrapsError(t *testing.T) {
	inner := E(KindNotFound, "user not found")
	outer := E(Op("handler.GetUser"), inner)

	if outer.Err != inner {
		t.Error("wrapped error not preserved")
	}
	if outer.Kind != KindNotFound {
		t.Errorf("Kind = %v, want KindNotFound (inherited)", outer.Kind)
	}
}

func TestE_KindInheritance(t *testing.T) {
	inner := E(KindForbidden, "access denied")
	outer := E(Op("middleware.Auth"), inner)

	if outer.Kind != KindForbidden {
		t.Errorf("Kind = %v, want KindForbidden (inherited from inner)", outer.Kind)
	}
}

func TestE_KindNoOverride(t *testing.T) {
	inner := E(KindNotFound, "missing")
	outer := E(Op("handler"), KindValidation, inner)

	if outer.Kind != KindValidation {
		t.Errorf("Kind = %v, want KindValidation (explicit takes precedence)", outer.Kind)
	}
}

func TestE_StackTrace(t *testing.T) {
	err := E("test error")
	frames := err.StackTrace()
	if len(frames) == 0 {
		t.Error("expected stack trace frames")
	}
}

func TestNew(t *testing.T) {
	err := New("something went wrong")
	if err.Message != "something went wrong" {
		t.Errorf("Message = %q, want something went wrong", err.Message)
	}
	if err.StackTrace() == nil {
		t.Error("expected stack trace")
	}
}

func TestNewf(t *testing.T) {
	err := Newf("port %d invalid", 99999)
	if !strings.Contains(err.Message, "99999") {
		t.Errorf("Message = %q, want to contain 99999", err.Message)
	}
}

func TestWrap(t *testing.T) {
	cause := stderrors.New("connection refused")
	err := Wrap(cause, Op("storage.Connect"))

	if err.Op != "storage.Connect" {
		t.Errorf("Op = %q, want storage.Connect", err.Op)
	}
	if err.Err != cause {
		t.Error("wrapped error not preserved")
	}
}

func TestWrap_Nil(t *testing.T) {
	err := Wrap(nil, Op("noop"))
	if err != nil {
		t.Errorf("Wrap(nil) = %v, want nil", err)
	}
}

func TestWrapf(t *testing.T) {
	cause := stderrors.New("timeout")
	err := Wrapf(cause, "after %d retries", 3)

	if !strings.Contains(err.Message, "3 retries") {
		t.Errorf("Message = %q, want to contain 3 retries", err.Message)
	}
	if err.Err != cause {
		t.Error("wrapped error not preserved")
	}
}

func TestWrapf_Nil(t *testing.T) {
	err := Wrapf(nil, "noop %d", 1)
	if err != nil {
		t.Errorf("Wrapf(nil) = %v, want nil", err)
	}
}

func TestError_Format(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{
			name: "op and message",
			err:  &Error{Op: "server.Start", Message: "bind failed"},
			want: "server.Start: bind failed",
		},
		{
			name: "kind and message",
			err:  &Error{Kind: KindNotFound, Message: "user not found"},
			want: "not found: user not found",
		},
		{
			name: "full",
			err:  &Error{Op: "handler", Kind: KindValidation, Code: CodeConfigInvalid, Message: "bad port"},
			want: "handler: validation: [config_invalid] bad port",
		},
		{
			name: "message only",
			err:  &Error{Message: "simple error"},
			want: "simple error",
		},
		{
			name: "wrapped error no message",
			err:  &Error{Op: "outer", Err: stderrors.New("inner error")},
			want: "outer: inner error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUnwrap(t *testing.T) {
	inner := stderrors.New("root cause")
	err := E(Op("test"), inner)

	if err.Unwrap() != inner {
		t.Error("Unwrap() did not return inner error")
	}
}

func TestGetKind(t *testing.T) {
	inner := E(KindNotFound, "missing")
	outer := E(Op("handler"), inner)

	kind := GetKind(outer)
	if kind != KindNotFound {
		t.Errorf("GetKind() = %v, want KindNotFound", kind)
	}
}

func TestGetKind_NoKind(t *testing.T) {
	err := stderrors.New("plain error")
	kind := GetKind(err)
	if kind != 0 {
		t.Errorf("GetKind() = %v, want 0", kind)
	}
}

func TestGetCode(t *testing.T) {
	inner := E(CodeStorageQuery, "query failed")
	outer := E(Op("handler"), inner)

	code := GetCode(outer)
	if code != CodeStorageQuery {
		t.Errorf("GetCode() = %q, want %q", code, CodeStorageQuery)
	}
}

func TestGetCode_NoCode(t *testing.T) {
	err := E("no code")
	code := GetCode(err)
	if code != "" {
		t.Errorf("GetCode() = %q, want empty", code)
	}
}

func TestGetOp(t *testing.T) {
	err := E(Op("server.Start"), "failed")
	op := GetOp(err)
	if op != "server.Start" {
		t.Errorf("GetOp() = %q, want server.Start", op)
	}
}

func TestGetOp_NoOp(t *testing.T) {
	err := stderrors.New("plain")
	op := GetOp(err)
	if op != "" {
		t.Errorf("GetOp() = %q, want empty", op)
	}
}

func TestOps(t *testing.T) {
	inner := E(Op("config.Parse"), "bad value")
	mid := E(Op("config.Load"), inner)
	outer := E(Op("server.Start"), mid)

	ops := Ops(outer)
	if len(ops) != 3 {
		t.Fatalf("Ops() returned %d ops, want 3", len(ops))
	}
	if ops[0] != "server.Start" {
		t.Errorf("ops[0] = %q, want server.Start", ops[0])
	}
	if ops[1] != "config.Load" {
		t.Errorf("ops[1] = %q, want config.Load", ops[1])
	}
	if ops[2] != "config.Parse" {
		t.Errorf("ops[2] = %q, want config.Parse", ops[2])
	}
}

func TestHTTPStatus(t *testing.T) {
	tests := []struct {
		kind Kind
		want int
	}{
		{KindNotFound, http.StatusNotFound},
		{KindValidation, http.StatusBadRequest},
		{KindUnauthorized, http.StatusUnauthorized},
		{KindForbidden, http.StatusForbidden},
		{KindConflict, http.StatusConflict},
		{KindTimeout, http.StatusGatewayTimeout},
		{KindUnavailable, http.StatusServiceUnavailable},
		{KindRateLimit, http.StatusTooManyRequests},
		{KindCanceled, http.StatusRequestTimeout},
		{KindInternal, http.StatusInternalServerError},
		{KindUnexpected, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.kind.String(), func(t *testing.T) {
			err := E(tt.kind, "test")
			got := HTTPStatus(err)
			if got != tt.want {
				t.Errorf("HTTPStatus() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestHTTPStatus_PlainError(t *testing.T) {
	err := stderrors.New("plain")
	got := HTTPStatus(err)
	if got != 500 {
		t.Errorf("HTTPStatus() = %d, want 500", got)
	}
}

func TestKind_String(t *testing.T) {
	tests := []struct {
		kind Kind
		want string
	}{
		{KindUnexpected, "unexpected"},
		{KindNotFound, "not found"},
		{KindValidation, "validation"},
		{KindUnauthorized, "unauthorized"},
		{KindForbidden, "forbidden"},
		{KindConflict, "conflict"},
		{KindTimeout, "timeout"},
		{KindInternal, "internal"},
		{KindUnavailable, "unavailable"},
		{KindRateLimit, "rate limit"},
		{KindCanceled, "canceled"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.kind.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestKind_StringOutOfRange(t *testing.T) {
	k := Kind(255)
	if got := k.String(); got != "unknown" {
		t.Errorf("String() = %q, want unknown", got)
	}
}

func TestFrame_String(t *testing.T) {
	f := Frame{Function: "main.run", File: "/app/main.go", Line: 42}
	got := f.String()
	if !strings.Contains(got, "main.run") {
		t.Error("missing function name")
	}
	if !strings.Contains(got, "/app/main.go:42") {
		t.Error("missing file:line")
	}
}

func TestStack_Format(t *testing.T) {
	err := E("stack test")
	if err.stack == nil {
		t.Fatal("stack is nil")
	}
	formatted := err.stack.Format()
	if formatted == "" {
		t.Error("Format() returned empty string")
	}
}

func TestStack_NilFormat(t *testing.T) {
	var s *stack
	if got := s.Format(); got != "" {
		t.Errorf("nil stack Format() = %q, want empty", got)
	}
}

func TestIsAsUnwrapJoin(t *testing.T) {
	err1 := stderrors.New("error 1")
	err2 := stderrors.New("error 2")

	if !Is(err1, err1) {
		t.Error("Is() should match same error")
	}

	var target *Error
	e := E("test")
	if !As(e, &target) {
		t.Error("As() should find *Error")
	}

	wrapped := E(Op("test"), err1)
	if Unwrap(wrapped) != err1 {
		t.Error("Unwrap() should return inner error")
	}

	joined := Join(err1, err2)
	if joined == nil {
		t.Error("Join() should not return nil")
	}
}
