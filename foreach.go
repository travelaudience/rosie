package rosie

import (
	"context"
	"fmt"
	"io"
	"reflect"

	"github.com/travelaudience/rosie/pkg/dag"
)

// ForEach allows executing logic for each piece of the result produced by step before.
// It will panic if received data is not a slice or a map.
// Finally once each end every piece of work is done all slice are gathered by the closing task (group end) in form of a slice.
func ForEach(name string, fn func(key string) Attacher) *GroupTask {
	beginning := &FnTask{
		task: newHiddenTask(fmt.Sprintf("for-each(%s)", name)),
	}
	end := &FnTask{
		task: newHiddenTask(fmt.Sprintf("for-each(%s)-gather-slice", name)),
	}

	anchorBeginning, anchorEnd := dag.New()
	beginning.setAnchor(anchorBeginning, beginning)
	end.setAnchor(anchorEnd, end)

	beginning.closure = func(_ context.Context, _ io.Writer, res Resulter) (interface{}, error) {
		if fn == nil {
			return nil, nil
		}

		add := func(key string, res Result) {
			staticInputTask := newHiddenTask(fmt.Sprintf("%s-static-input", key))
			staticInputTask.setResult(res)
			staticInputTask.anchor.Between(anchorBeginning, anchorEnd)

			fn(key).Node().Between(staticInputTask.anchor, anchorEnd)
		}
		val := reflect.ValueOf(res.Result().value)
		switch val.Type().Kind() {
		case reflect.Slice:
			if val.IsNil() {
				return nil, nil
			}
			for i := 0; i < val.Len(); i++ {
				add(fmt.Sprintf("%d/%d", i+1, val.Len()), Result{
					value: val.Index(i).Interface(),
				})
			}
		case reflect.Map:
			for _, key := range val.MapKeys() {
				add(fmt.Sprintf("%v", key.Interface()), Result{
					key:   fmt.Sprintf("%v", key.Interface()),
					value: val.MapIndex(key).Interface(),
				})
			}
		default:
			return nil, fmt.Errorf("rosie: for-each: unexpected type: %T", res.Result())
		}
		return nil, nil
	}

	end.closure = func(_ context.Context, _ io.Writer, res Resulter) (interface{}, error) {
		return res.Result().Value(), nil
	}
	return &GroupTask{
		name:      fmt.Sprintf("for-each(%s)", name),
		beginning: beginning.task,
		end:       end.task,
	}
}
