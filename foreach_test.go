package rosie_test

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/travelaudience/rosie"
	"github.com/travelaudience/rosie/pkg/runner/testrunner"
)

func TestForEach(t *testing.T) {
	var (
		gotSlice []string
		gotMap   map[string]string
	)
	cases := map[string]struct {
		init     func() *rosie.GroupTask
		expSlice []string
		expMap   map[string]string
	}{
		"nil": {
			expSlice: []string{"nothing"},
			init: func() *rosie.GroupTask {
				g := rosie.Group("test-foreach-nil")
				g.Beginning().
					Then(rosie.ForEach("print", nil)).
					Then(rosie.Fn("step-after-empty-foreach", func(_ context.Context, _ io.Writer, _ rosie.Resulter) (interface{}, error) {
						gotSlice = []string{"nothing"}
						return nil, nil
					}))
				return g
			},
		},
		"slice": {
			expSlice: []string{"3/3", "2/3", "1/3"},
			init: func() *rosie.GroupTask {
				g := rosie.Group("test-foreach-slice")
				g.Beginning().
					Then(rosie.Fn("create-slice", func(_ context.Context, _ io.Writer, _ rosie.Resulter) (interface{}, error) {
						return []string{"A", "B", "C"}, nil
					})).
					Then(rosie.ForEach("print", func(key string) rosie.Attacher {
						gotSlice = append(gotSlice, key)
						return rosie.Cmd("echo", "echo", fmt.Sprintf("%s -> [[ .Result ]]", key))
					}))
				return g
			},
		},
		"map": {
			expMap: map[string]string{"A": "2", "B": "4", "C": "6"},
			init: func() *rosie.GroupTask {
				g := rosie.Group("test-foreach-map")
				g.Beginning().
					Then(rosie.Fn("create-slice", func(_ context.Context, _ io.Writer, _ rosie.Resulter) (interface{}, error) {
						return map[string]int{"A": 1, "B": 2, "C": 3}, nil
					})).
					Then(rosie.ForEach("", func(key string) rosie.Attacher {
						return rosie.Fn("multiply", func(_ context.Context, _ io.Writer, res rosie.Resulter) (interface{}, error) {
							return strconv.FormatInt(int64(res.Result().Value().(int)*2), 10), nil
						})
					})).
					Then(rosie.Fn("result", func(_ context.Context, _ io.Writer, res rosie.Resulter) (interface{}, error) {
						gotMap = res.Result().Value().(map[string]string)
						return nil, nil
					}))
				return g
			},
		},
	}

	for hint, c := range cases {
		t.Run(hint, func(t *testing.T) {
			gotSlice = []string{}
			gotMap = make(map[string]string)

			testrunner.Run(t, c.init(), noError)

			switch {
			case c.expSlice != nil:
				expLength := len(c.expSlice)
				gotLength := len(gotSlice)

				if gotLength != expLength {
					t.Fatalf("wrong result length expected %d but got %d", expLength, gotLength)
				}
				for _, value := range gotSlice {
					if !strings.Contains(strings.Join(c.expSlice, "-"), value) {
						t.Error("such key does not exists")
					}
				}
			case c.expMap != nil:
				expLength := len(c.expMap)
				gotLength := len(gotMap)

				if gotLength != expLength {
					t.Fatalf("wrong result length expected %d but got %d", expLength, gotLength)
				}
				if !reflect.DeepEqual(gotMap, c.expMap) {
					t.Errorf("got wrong map expected:\n	%v\nbut got:\n	%v", c.expMap, gotMap)
				}
			}
		})
	}
}
