package dag

import (
	"testing"
)

func TestNode_Between_graph(t *testing.T) {
	b1, e1 := New()
	b2, e2 := New()

	b1.Data = "b1"
	b2.Data = "b2"
	e1.Data = "e1"
	e2.Data = "e2"

	b2.Between(b1, e1)

	assertContains(t, b2, b1.Children())
	assertContainsNot(t, e1, b1.Children())
	assertContains(t, e2, e1.Parents())
	assertContainsNot(t, b1, e1.Parents())
}

func TestNode_Between_node(t *testing.T) {
	b, e := New()
	n := &Node{}

	b.Data = "b"
	e.Data = "e"
	n.Data = "n"
	n.Between(b, e)

	assertContains(t, n, b.Children())
	assertContainsNot(t, e, b.Children())
	assertContains(t, n, e.Parents())
	assertContainsNot(t, b, e.Parents())
}

func TestNode_After_graph(t *testing.T) {
	b1, e1 := New()
	b2, e2 := New()

	b1.Data = "b1"
	b2.Data = "b2"
	e1.Data = "e1"
	e2.Data = "e2"

	b1.After(b2)

	assertContains(t, b2, b1.Children())
	assertContainsNot(t, e1, b1.Children())
	assertContains(t, e2, e1.Parents())
	assertContainsNot(t, b1, e1.Parents())
}

func TestNode_After_node(t *testing.T) {
	b, e := New()
	n := &Node{}

	b.Data = "b"
	e.Data = "e"
	n.Data = "n"
	b.After(n)

	assertContains(t, n, b.Children())
	assertContainsNot(t, e, b.Children())
	assertContains(t, n, e.Parents())
	assertContainsNot(t, b, e.Parents())
}

func assertContains(t *testing.T, n *Node, nodes Nodes) {
	t.Helper()

	if !nodes.contains(n) {
		t.Errorf("node (%v) is missing in\n	%s", n.Data, nodes)
	}
}

func assertContainsNot(t *testing.T, n *Node, nodes Nodes) {
	t.Helper()

	if nodes.contains(n) {
		t.Error("node is present")
	}
}
