package config

import (
	"context"
	"fmt"
	"reflect"
)

// NoOpValidator is a validator that does nothing.
type NoOpValidator struct{}

func (NoOpValidator) Validate(ctx context.Context, v any) error {
	return nil
}

// FuncValidator wraps a validation function.
type FuncValidator struct {
	fn func(context.Context, any) error
}

// NewFuncValidator creates a validator from a function.
func NewFuncValidator(fn func(context.Context, any) error) *FuncValidator {
	return &FuncValidator{fn: fn}
}

func (v *FuncValidator) Validate(ctx context.Context, target any) error {
	return v.fn(ctx, target)
}

// ============================================================================
// Helpers
// ============================================================================

func validateTarget(target any) error {
	if target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}

	if rv.IsNil() {
		return fmt.Errorf("target pointer cannot be nil")
	}

	elem := rv.Elem()
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("target must point to a struct, got %s", elem.Kind())
	}

	return nil
}
