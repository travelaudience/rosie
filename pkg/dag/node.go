package dag

//go:generate stringer -type=Type,status -output=node.stringer.go

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

const (
	// TypeMiddle represents the node type located in the middle of the graph.
	TypeMiddle Type = iota
	// TypeBeginning represents the node type that indicates beginning of the graph.
	TypeBeginning
	// TypeEnd represents the node type that indicates end of the graph.
	TypeEnd
	// TypeMiddleBeginning represents the node type that indicates the beginning of a subgraph.
	TypeMiddleBeginning
	// TypeMiddleEnd represents the node type that indicates the end of a subgraph.
	TypeMiddleEnd
	// TypeHidden represents the node type that is ignored by a drawer, other than that.
	// It can be considered equal to TypeMiddle.
	TypeHidden
)

const (
	statusNotSeen status = iota
	statusVisited
	statusDone
	statusFailed
)

type (
	// Type is a kind of a node.
	Type   int
	status int
)

// Node is a single vertex of a graph.
type Node struct {
	// Data stores logic associated with the node that may be executed by a visitor.
	Data              interface{}

	status            status
	kind              Type
	parents, children Nodes
	beginning, end    *Node
}

// New instantiate a new graph. It returns its beginning and its end.
// It's up to the caller how it wants to interact with the graph.
func New() (*Node, *Node) {
	b := &Node{kind: TypeBeginning}
	e := &Node{kind: TypeEnd}

	b.children.add(e)
	e.parents.add(b)
	b.end = e
	e.beginning = b

	return b, e
}

// Hidden instantiates an "invisible" node.
func Hidden(d interface{}) *Node {
	return &Node{
		Data: d,
		kind: TypeHidden,
	}
}

func (n *Node) isGraph() bool {
	return n.beginning != nil || n.end != nil
}

// Type returns node kind.
func (n *Node) Type() Type {
	return n.kind
}

// Done returns true if node can be considered visited.
func (n *Node) Done() bool {
	return n.status == statusDone || n.kind == TypeBeginning
}

// MarkAsDone marks node as already processed.
func (n *Node) MarkAsDone() {
	n.status = statusDone
}

// MarkAsFailed marks node as failed.
func (n *Node) MarkAsFailed() {
	n.status = statusFailed
}

// GoString implements fmt GoStringer interface.
func (n Node) GoString() string {
	buf := bytes.NewBuffer(nil)
	n.goString(buf, 0)

	return buf.String()
}

func (n Node) goString(w io.Writer, i int) {
	if dat, ok := n.Data.(interface{ Name() string }); ok {
		_, _ = fmt.Fprintf(w, fmt.Sprintf("%%%ds %%s\n", i*2), "", dat.Name())
	} else {
		_, _ = fmt.Fprintf(w, fmt.Sprintf("%%%ds %%v\n", i*2), "", n.Data)
	}

	for _, child := range n.children {
		child.goString(w, i+1)
	}
}

// After allows to attach node one after another.
// It takes care of already existing connections receiver node has,
// so that node that is injected become a proper bridge.
// Injected node loses its own connections.
func (n *Node) After(node *Node) {
	for _, child := range n.children {
		if node.isGraph() {
			child.parents.replace(n, node.end)
			node.end.children.add(child)
		} else {
			child.parents.replace(n, node)
			node.children.add(child)
		}
		n.children.remove(child)
	}

	if node.isGraph() {
		if node.beginning != nil {
			panic("cant pass last node of a group into after method")
		}
		node.kind = middleType(node.kind)
		node.end.kind = middleType(node.end.kind)
	}
	node.parents.add(n)
	n.children.add(node)
}

// Between allows to inject receiver node between given nodes.
// Receiver node loses its own connections.
func (n *Node) Between(beginning, end *Node) {
	if n.isGraph() {
		beginning.children.replace(end, n)
		end.parents.replace(beginning, n.end)

		n.parents.add(beginning)
		n.end.children.add(end)

		n.kind = middleType(n.kind)
		n.end.kind = middleType(n.end.kind)

		return
	}

	n.between(beginning, end)
}

func (n *Node) between(parent, child *Node) {
	for i, c := range parent.children {
		if c == child {
			parent.children = append(parent.children[:i], parent.children[i+1:]...)
			break
		}
	}
	for i, p := range child.parents {
		if p == parent {
			child.parents = append(child.parents[:i], child.parents[i+1:]...)
			break
		}
	}

	parent.children.replace(child, n)
	child.parents.replace(parent, n)

	n.parents.add(parent)
	n.children.add(child)
}

// Children returns a collection of nodes given node is the parent of.
func (n *Node) Children() Nodes {
	return n.children
}

// Parents returns a collection of nodes given node is child of.
func (n *Node) Parents() Nodes {
	return n.parents
}

// Nodes is a set of nodes.
type Nodes []*Node

// String implements fmt Stringer interface.
func (n Nodes) String() string {
	var parts []string
	for _, nn := range n {
		if str, ok := nn.Data.(interface{ Name() string }); ok {
			parts = append(parts, fmt.Sprint(str.Name()))
		} else {
			parts = append(parts, fmt.Sprint(nn.Data))
		}
	}

	return strings.Join(parts, "\n")
}

func (n Nodes) done() bool {
	done := true
	for _, node := range n {
		if !node.Done() {
			done = false
		}
	}
	return done
}

func (n *Nodes) add(node *Node) bool {
	for _, nn := range *n {
		if nn == node {
			return false
		}
	}
	*n = append(*n, node)
	return true
}

func (n *Nodes) replace(before, after *Node) {
	for i, nn := range *n {
		if nn == before {
			(*n)[i] = after
			return
		}
	}

	n.add(after)
}

func (n *Nodes) remove(node *Node) {
	for i, nn := range *n {
		if nn == node {
			*n = append((*n)[:i], (*n)[i+1:]...)
			return
		}
	}
}

func (n Nodes) contains(node *Node) bool {
	for _, nn := range n {
		if nn == node {
			return true
		}
	}

	return false
}

func middleType(t Type) Type {
	switch t {
	case TypeBeginning:
		return TypeMiddleBeginning
	case TypeEnd:
		return TypeMiddleEnd
	default:
		return t
	}
}

func isMiddleType(t Type) bool {
	switch t {
	case TypeMiddle, TypeMiddleBeginning, TypeMiddleEnd, TypeHidden:
		return true
	default:
		return false
	}
}
