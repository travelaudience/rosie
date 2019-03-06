package rosie

import (
	"context"
	"io"
	"reflect"
)

type (
	// FnClosure ...
	FnClosure func(ctx context.Context, w io.Writer, res Resulter) (ret interface{}, err error)
	// FnClosureString ...
	FnClosureString func(ctx context.Context, w io.Writer, res string) (ret interface{}, err error)
	// FnClosureStringSlice ...
	FnClosureStringSlice func(ctx context.Context, w io.Writer, res []string) (ret interface{}, err error)
)

// StringClosure is handy for FnClosure that brings type safety.
// It checks if the passed value is a string, if not it returns TypeError.
func StringClosure(fn FnClosureString) FnClosure {
	return func(ctx context.Context, w io.Writer, res Resulter) (interface{}, error) {
		if str, ok := res.Result().Value().(string); ok {
			return fn(ctx, w, str)
		}
		return nil, TypeError(reflect.String, res.Result().Value())
	}
}

// StringSliceClosure is handy for FnClosure that brings type safety.
// It checks if the passed value is a slice of strings, if not it returns TypeError.
func StringSliceClosure(fn FnClosureStringSlice) FnClosure {
	return func(ctx context.Context, w io.Writer, res Resulter) (interface{}, error) {
		if str, ok := res.Result().Value().([]string); ok {
			return fn(ctx, w, str)
		}
		return nil, TypeError(reflect.TypeOf([]string{}).Kind(), res.Result().Value())
	}
}
