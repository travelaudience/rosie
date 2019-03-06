package rosie

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/travelaudience/rosie/pkg/dag"
)

type namer interface {
	// Name should return the human-friendly name.
	// The name should not be used for anything else than just a presentation.
	Name() string
}

type noder interface {
	// Node returns pointer to the direct acyclic graph node.
	// It needs to be thread-safe.
	Node() *dag.Node
}

// Joint ...
type Joint interface {
	namer
	noder
}

// Resulter ...
type Resulter interface {
	namer

	// Result returns a product of a single task.
	// It needs to be safe to call it multiple times with no side-effects.
	// It needs to be thread-safe.
	Result() Result
}

// Executor ...
type Executor interface {
	Resulter
	noder

	// Exec if implemented by a task allows to execute it.
	// A caller will be notified about the progress through the channel.
	// It should return an error if something went wrong during the initialization phase (if such exist).
	Exec(context.Context) (<-chan Piece, error)
}

// Attacher ...
type Attacher interface {
	Joint

	// Then allows to chain tasks one after another.
	// It's up to the implementation to define how it is going to work.
	// It is not always clear what 'then' could mean in a given context.
	// It needs to be thread-safe.
	Then(Attacher) Attacher
}

var (
	_ Attacher = &task{}
	_ Joint    = &task{}
	_ Resulter = &task{}
)

type task struct {
	name, description string
	anchor            *dag.Node
	result            Result

	lock sync.RWMutex
}

func newHiddenTask(name string) *task {
	t := &task{
		name: name,
		result: Result{
			taskName: name,
		},
	}
	n := dag.Hidden(t)
	t.anchor = n

	return t
}

func (t *task) setErr(err error) {
	t.lock.Lock()
	t.result.err = err
	t.lock.Unlock()
}

func (t *task) setResult(res Result) {
	t.lock.Lock()
	t.result = res
	t.lock.Unlock()
}

func (t *task) run() error {
	t.lock.Lock()
	if t.result.err != nil {
		t.anchor.MarkAsFailed()
	} else {
		t.anchor.MarkAsDone()
	}
	t.lock.Unlock()

	return t.Result().Err()
}

// Name implements namer interface.
func (t *task) Name() string {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.name
}

func (t *task) Desc() string {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.description
}

// Then implements Attacher interface.
func (t *task) Then(next Attacher) Attacher {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.anchor.After(next.Node())

	return next
}

// Result implements Resulter interface.
func (t *task) Result() Result {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.result.value != nil {
		return t.result
	}
	if res := t.gatherParentResults(); res != nil {
		return res.Result()
	}
	return Result{}
}

// Node implements noder interface.
func (t *task) Node() *dag.Node {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.anchor
}

func (t *task) setAnchor(anchor *dag.Node, data interface{}) {
	t.anchor = anchor
	anchor.Data = data
}

// TODO: simplify
func (t *task) gatherParentResults() Resulter {
	switch len(t.anchor.Parents()) {
	case 0:
		return nil
	case 1:
		return t.anchor.Parents()[0].Data.(Resulter)
	default:
		var (
			combinedValue reflect.Value
		)
		if value, ok := initSomeMap(t.anchor); ok {
			for _, parent := range t.anchor.Parents() {
				if res, ok := parent.Data.(Resulter); ok {
					value.SetMapIndex(
						reflect.ValueOf(res.Result().key),
						reflect.ValueOf(res.Result().value),
					)
				}
			}
			combinedValue = value
		}
		if value, ok := initSomeSlice(t.anchor); ok {
			for _, parent := range t.anchor.Parents() {
				if res, ok := parent.Data.(Resulter); ok {
					value = reflect.Append(value, reflect.ValueOf(res.Result().value))
				}
			}
			combinedValue = value
		}

		combined := combinedResults{
			name:  t.name,
			value: combinedValue.Interface(),
		}
		for _, parent := range t.anchor.Parents() {
			if res, ok := parent.Data.(Resulter); ok {
				if err := res.Result().Err(); err != nil {
					combined.err = appendError(combined.err, err)
				}
			}
		}

		return combined
	}
}

type staticResulter struct {
	res Result
}

func (r staticResulter) Name() string {
	return fmt.Sprintf("%s-%s", r.res.taskName, r.res.key)
}

func (r staticResulter) Result() Result {
	return r.res
}

type combinedResults struct {
	name  string
	value interface{}
	err   error
}

// Namer implements namer interface.
func (r combinedResults) Name() string {
	return r.name
}

func (r combinedResults) Desc() string {
	return ""
}

// Result implements Resulter interface.
func (r combinedResults) Result() Result {
	return Result{
		value: r.value,
	}
}

// Piece ...
type Piece struct {
	Text string
	Err  error
}

var stringType = reflect.ValueOf("string").Type()

func initSomeMap(n *dag.Node) (reflect.Value, bool) {
	var (
		kinds = make(map[string]reflect.Type)
		kind  reflect.Type
	)
	for _, parent := range n.Parents() {
		if res, ok := parent.Data.(Resulter); ok {
			if res.Result().key != "" {
				tof := reflect.ValueOf(res.Result().value).Type()
				kinds[tof.String()] = tof
				kind = tof
			}
		}
	}
	switch len(kinds) {
	case 0:
		return reflect.Value{}, false
	case 1:
	default:
		var empty interface{}
		kind = reflect.TypeOf(empty)
	}

	return reflect.MakeMapWithSize(
		reflect.MapOf(stringType, kind),
		len(kinds),
	), true
}

func initSomeSlice(n *dag.Node) (reflect.Value, bool) {
	var (
		kinds = make(map[string]reflect.Type)
		kind  reflect.Type
	)
	for _, parent := range n.Parents() {
		if res, ok := parent.Data.(Resulter); ok {
			if res.Result().key == "" {
				tof := reflect.ValueOf(res.Result().value).Type()
				kinds[tof.String()] = tof
				kind = tof
			}
		}
	}
	switch len(kinds) {
	case 0:
		return reflect.Value{}, false
	case 1:
	default:
		var empty interface{}
		kind = reflect.TypeOf(empty)
	}

	return reflect.MakeSlice(reflect.SliceOf(kind), 0, len(kinds)), true
}
