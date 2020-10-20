package dag

import (
	"errors"
	"io"
)

// Walker allows to travers a graph.
type Walker struct {
	stack
	previous *Node
}

// NewWalker instantiate a new Walker object.
// It will return an error if the given node is not a valid root node.
func NewWalker(root *Node) (*Walker, error) {
	if root.kind != TypeBeginning {
		return nil, errors.New("rosie: dag: start node expected")
	}

	w := &Walker{}
	w.push(root)

	return w, nil
}


// ErrBrokenGraph can be returned by Walk if given graph is not traversable.
var ErrBrokenGraph = errors.New("rosie: dag: broken graph")

// Walk traverse the graph.
// It returns nodes in the order they should be processed.
// It might visit a single node multiple times.
func (w *Walker) Walk() (*Node, error) {
	var memory stack
	defer func() {
		for {
			node, ok := memory.pop()
			if !ok {
				return
			}

			w.push(node)
		}
	}()

Start:
	if w.previous != nil && w.previous.status != statusFailed {
		for n := len(w.previous.children) - 1; n >= 0; n-- {
			if w.previous.children[n].status == statusNotSeen {
				w.previous.children[n].status = statusVisited
				w.push(w.previous.children[n])
			}
		}
	}

	node, ok := w.pop()
	if !ok {
		if !memory.isEmpty() {
			return nil, ErrBrokenGraph
		}
		return nil, io.EOF
	}

	if node.Done() && isMiddleType(node.kind) {
		goto Start
	}

	if node.parents.done() {
		if node.kind == TypeEnd {
			return nil, io.EOF
		}

		w.previous = node
		return node, nil
	}

	memory.push(node)
	goto Start
}
