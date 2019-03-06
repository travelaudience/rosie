package rosie

import (
	"fmt"

	"github.com/travelaudience/rosie/pkg/dag"
)

var (
	_ Attacher = &GroupTask{}
	_ Joint    = &GroupTask{}
)

// GroupTask ...
type GroupTask struct {
	name           string
	beginning, end *task
}

// Group ...
func Group(name string, tasks ...Attacher) *GroupTask {
	anchorBeginning, anchorEnd := dag.New()
	b := &task{name: name}
	e := &task{name: fmt.Sprintf("%s-end", name)}

	b.setAnchor(anchorBeginning, b)
	e.setAnchor(anchorEnd, e)

	previous := anchorBeginning
	for _, task := range tasks {
		task.Node().Between(previous, anchorEnd)
		previous = task.Node()
	}

	return &GroupTask{
		name:      name,
		beginning: b,
		end:       e,
	}
}

// Name implements namer interface.
func (g *GroupTask) Name() string {
	return g.name
}

// Node implements noder interface.
func (g *GroupTask) Node() *dag.Node {
	return g.beginning.Node()
}

// Then implements Attacher interface.
func (g *GroupTask) Then(next Attacher) Attacher {
	g.end.anchor.After(next.Node())

	return next
}

// Iter ...
func (g *GroupTask) Iter() (*Iterator, error) {
	return newIterator(g.beginning.anchor)
}

// Beginning ...
func (g *GroupTask) Beginning() Attacher {
	return g.beginning
}

// End ...
func (g *GroupTask) End() Attacher {
	return g.end
}
