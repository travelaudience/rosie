package rosie

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strconv"

	"gopkg.in/yaml.v2"

	"github.com/travelaudience/rosie/pkg/dag"
	"github.com/travelaudience/rosie/pkg/vars"
)

var (
	_ Executor = &FnTask{}
)

// FnTask ...
type FnTask struct {
	*task
	closure          FnClosure
	previousResulter Resulter
}

// Fn ...
func Fn(name string, closure FnClosure) *FnTask {
	t := &FnTask{
		task:    &task{name: name},
		closure: closure,
	}
	t.setAnchor(&dag.Node{}, t)
	return t
}

type transformKey struct{}

// Nothing can be used in combination with Transform to exclude given object from a collection.
var Nothing = &transformKey{}

func isNothing(obj interface{}) bool {
	if key, ok := obj.(*transformKey); ok && key == Nothing {
		return true
	}
	return false
}

// Transform ...
// TODO: is it a right name?
func Transform(name string, filter FnClosure) *FnTask {
	return Fn(name, func(ctx context.Context, w io.Writer, res Resulter) (interface{}, error) {
		if res.Result().Value() == nil {
			return nil, nil
		}
		val := reflect.ValueOf(res.Result().Value())
		switch val.Type().Kind() {
		case reflect.Slice:
			x := reflect.MakeSlice(reflect.SliceOf(val.Type().Elem()), 0, val.Cap())
			for i := 0; i < val.Len(); i++ {
				obj := val.Index(i).Interface()
				got, err := filter(ctx, w, staticResulter{res: Result{
					taskName: name,
					key:      strconv.FormatInt(int64(i), 10),
					value:    obj,
				}})
				if err != nil {
					return nil, err
				}
				if !isNothing(got) {
					x = reflect.Append(x, reflect.ValueOf(got))
				}
			}
			return x.Interface(), nil
		case reflect.Map:
			x := reflect.MakeMap(val.Type())
			for _, key := range val.MapKeys() {
				obj := val.MapIndex(key).Interface()
				got, err := filter(ctx, w, staticResulter{res: Result{
					taskName: name,
					key:      key.String(),
					value:    obj,
				}})
				if err != nil {
					return nil, err
				}
				if !isNothing(got) {
					x.SetMapIndex(key, reflect.ValueOf(got))
				}
			}
			return x.Interface(), nil
		default:
			return nil, errors.New("rosie: transform: unknown type")
		}
	})
}

// Without modifies a collection by removing entries.
// If the given argument is a slice, it will compare values. In case it is a map, it will remove keys.
func Without(without interface{}) *FnTask {
	val := reflect.ValueOf(without)
	switch val.Type().Kind() {
	case reflect.Slice:
		return Transform("without", func(_ context.Context, _ io.Writer, res Resulter) (interface{}, error) {
			for i := 0; i < val.Len(); i++ {
				if reflect.DeepEqual(val.Index(i).Interface(), res.Result().Value()) {
					return Nothing, nil
				}
			}
			return res.Result().Value(), nil
		})
	case reflect.Map:
		return Transform("without", func(_ context.Context, _ io.Writer, res Resulter) (interface{}, error) {
			for _, key := range val.MapKeys() {
				if key.String() == res.Result().key {
					return Nothing, nil
				}
			}
			return res.Result().Value(), nil
		})
	default:
		return Transform("without", func(_ context.Context, _ io.Writer, res Resulter) (interface{}, error) {
			return nil, fmt.Errorf("rosie: without: unknown type: %T", without)
		})
	}
}

// UnmarshalFile ...
func UnmarshalFile(into interface{}) *FnTask {
	return Fn("unmarshal-file", func(_ context.Context, w io.Writer, res Resulter) (interface{}, error) {
		filePath, ok := res.Result().Value().(string)
		if !ok {
			return nil, TypeError(reflect.String, res.Result().Value())
		}

		/* #nosec */
		buf, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		if _, err = fmt.Fprintf(w, "file %s read %dB\n", filePath, len(buf)); err != nil {
			return nil, err
		}

		ext := path.Ext(filePath)
		switch ext {
		case vars.ExtYAML:
			if err = yaml.Unmarshal(buf, into); err != nil {
				return nil, err
			}
		case vars.ExtJSON:
			if err = json.Unmarshal(buf, into); err != nil {
				return nil, err
			}
		}

		if _, err = fmt.Fprintf(w, "file content unmarshaled using %s unmarshaller\n", ext); err != nil {
			return nil, err
		}

		return into, nil
	})
}

// Exec implements Executor interface.
func (t *FnTask) Exec(ctx context.Context) (<-chan Piece, error) {
	if t.previousResulter == nil {
		t.previousResulter = t.gatherParentResults()
	}

	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	done := make(chan struct{})
	out := make(chan Piece)

	go func() {
		previousResult := t.previousResulter.Result()
		val, err := t.closure(ctx, w, t.previousResulter)
		if err != nil {
			out <- Piece{Err: err}
		}
		t.setResult(Result{key: previousResult.key, value: val, err: err})
		if err := w.Close(); err != nil {
			out <- Piece{Err: err}
		}
		close(done)
	}()

	go func() {
		sc := bufio.NewScanner(r)
		for sc.Scan() {
			out <- Piece{Text: sc.Text()}
		}
		<-done
		if err := t.task.run(); err != nil {
			t.setErr(err)
			out <- Piece{Err: err}
		}
		close(out)
	}()

	return out, nil
}
