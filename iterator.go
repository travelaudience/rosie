package rosie

import (
	"io"

	"github.com/travelaudience/rosie/pkg/dag"
)

// Iterator ...
type Iterator struct {
	*dag.Walker
}

func newIterator(node *dag.Node) (*Iterator, error) {
	walk, err := dag.NewWalker(node)
	if err != nil {
		return nil, err
	}
	return &Iterator{
		Walker: walk,
	}, nil
}

// Next ...
func (i *Iterator) Next() (Joint, bool) {
Start:
	node, err := i.Walk()
	if err != nil {
		if err == io.EOF {
			return nil, false
		}
		panic(err) // TODO: do better
	}

	switch data := node.Data.(type) {
	case Executor:
		return data, true
	case Joint:
		node.MarkAsDone()

		if node.Type() != dag.TypeEnd {
			return data, true
		}
	default:
		node.MarkAsDone()
	}

	goto Start
}
