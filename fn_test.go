package rosie_test

import (
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/travelaudience/rosie"
	"github.com/travelaudience/rosie/pkg/runner/testrunner"
)

func TestTransform(t *testing.T) {
	cases := map[string]struct {
		init   func(t *testing.T) *rosie.GroupTask
		assert func(*testing.T, error)
	}{
		"slice": {
			init: func(t *testing.T) *rosie.GroupTask {
				g := rosie.Group("test-group")
				g.Beginning().
					Then(rosie.Fn("stub", func(_ context.Context, _ io.Writer, _ rosie.Resulter) (interface{}, error) {
						return []string{"A", "B", "c", "d"}, nil
					})).
					Then(rosie.Transform("lowercase", rosie.StringClosure(func(_ context.Context, _ io.Writer, res string) (interface{}, error) {
						if strings.ToLower(res) != res {
							return rosie.Nothing, nil
						}
						return res, nil
					}))).
					Then(assert(t, []string{"c", "d"}))
				return g
			},
			assert: noError,
		},
		"map": {
			init: func(t *testing.T) *rosie.GroupTask {
				g := rosie.Group("test-group")
				g.Beginning().
					Then(rosie.Fn("stub", func(_ context.Context, _ io.Writer, _ rosie.Resulter) (interface{}, error) {
						return map[int]string{
							1: "A",
							2: "B",
							3: "c",
							4: "d",
						}, nil
					})).
					Then(rosie.Transform("lowercase", rosie.StringClosure(func(_ context.Context, _ io.Writer, res string) (interface{}, error) {
						if strings.ToLower(res) != res {
							return rosie.Nothing, nil
						}
						return res, nil
					}))).
					Then(assert(t, map[int]string{3: "c", 4: "d"}))
				return g
			},
			assert: noError,
		},
		"nil-argument": {
			init: func(t *testing.T) *rosie.GroupTask {
				g := rosie.Group("test-group")
				g.Beginning().
					Then(rosie.Fn("stub", func(_ context.Context, _ io.Writer, _ rosie.Resulter) (interface{}, error) {
						return nil, nil
					})).
					Then(rosie.Transform("nil-works-just-ok", func(_ context.Context, _ io.Writer, val rosie.Resulter) (interface{}, error) {
						t.Error("should not be called", val)
						return nil, errors.New("should not be executed")
					}))
				return g
			},
			assert: noError,
		},
		"wrong-type": {
			init: func(t *testing.T) *rosie.GroupTask {
				g := rosie.Group("test-group")
				g.Beginning().
					Then(rosie.Fn("stub", func(_ context.Context, _ io.Writer, _ rosie.Resulter) (interface{}, error) {
						return struct{}{}, nil
					})).
					Then(rosie.Transform("nil-works-just-ok", func(_ context.Context, _ io.Writer, res rosie.Resulter) (interface{}, error) {
						t.Error("should not be called", res.Result())
						return nil, errors.New("should not be executed")
					}))
				return g
			},
			assert: func(t *testing.T, e error) {
				t.Helper()

				if e.Error() != "rosie: transform: unknown type" {
					t.Fatal(e)
				}
			},
		},
		"return-an-error": {
			init: func(t *testing.T) *rosie.GroupTask {
				g := rosie.Group("test-group")
				g.Beginning().
					Then(rosie.Fn("stub", func(_ context.Context, _ io.Writer, _ rosie.Resulter) (interface{}, error) {
						return []string{"a", "b"}, nil
					})).
					Then(rosie.Transform("nil-works-just-ok", func(_ context.Context, _ io.Writer, _ rosie.Resulter) (interface{}, error) {
						return nil, io.EOF
					}))
				return g
			},
			assert: func(t *testing.T, e error) {
				t.Helper()

				if e != io.EOF {
					t.Fatal(e)
				}
			},
		},
	}

	for hint, c := range cases {
		t.Run(hint, func(t *testing.T) {
			testrunner.Run(t, c.init(t), c.assert)
		})
	}
}

func TestWithout(t *testing.T) {
	cases := map[string]struct {
		init   func(t *testing.T) *rosie.GroupTask
		assert func(*testing.T, error)
	}{
		"slice": {
			init: func(t *testing.T) *rosie.GroupTask {
				g := rosie.Group("test-group")
				g.Beginning().
					Then(rosie.Fn("stub", func(_ context.Context, _ io.Writer, _ rosie.Resulter) (interface{}, error) {
						return []string{"A", "B", "c", "d"}, nil
					})).
					Then(rosie.Without([]string{"c", "d"})).
					Then(assert(t, []string{"A", "B"}))
				return g
			},
			assert: noError,
		},
		"map-without-value": {
			init: func(t *testing.T) *rosie.GroupTask {
				g := rosie.Group("test-group")
				g.Beginning().
					Then(rosie.Fn("stub", func(_ context.Context, _ io.Writer, _ rosie.Resulter) (interface{}, error) {
						return map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}, nil
					})).
					Then(rosie.Without([]int{2, 4})).
					Then(assert(t, map[string]int{"a": 1, "c": 3}))
				return g
			},
			assert: noError,
		},
		"map-without-key": {
			init: func(t *testing.T) *rosie.GroupTask {
				g := rosie.Group("test-group")
				g.Beginning().
					Then(rosie.Fn("stub", func(_ context.Context, _ io.Writer, _ rosie.Resulter) (interface{}, error) {
						return map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}, nil
					})).
					Then(rosie.Without(map[string]int{"a": 2, "b": 4})).
					Then(assert(t, map[string]int{"c": 3, "d": 4}))
				return g
			},
			assert: noError,
		},
		"wrong-type": {
			init: func(t *testing.T) *rosie.GroupTask {
				g := rosie.Group("test-group")
				g.Beginning().
					Then(rosie.Fn("stub", func(_ context.Context, _ io.Writer, _ rosie.Resulter) (interface{}, error) {
						return []string{"a", "b"}, nil
					})).
					Then(rosie.Without(struct{}{}))
				return g
			},
			assert: func(t *testing.T, e error) {
				t.Helper()

				if e.Error() != "rosie: without: unknown type: struct {}" {
					t.Fatal(e)
				}
			},
		},
	}

	for hint, c := range cases {
		t.Run(hint, func(t *testing.T) {
			testrunner.Run(t, c.init(t), c.assert)
		})
	}
}

func noError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
}

func assert(t *testing.T, exp interface{}) *rosie.FnTask {
	t.Helper()

	return rosie.Fn("assert", func(_ context.Context, _ io.Writer, res rosie.Resulter) (interface{}, error) {
		if !reflect.DeepEqual(res.Result().Value(), exp) {
			t.Errorf("wrong product, expected %v but got %v", exp, res)
		}
		return nil, nil
	})
}
