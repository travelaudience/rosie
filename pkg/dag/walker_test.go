package dag

import (
	"io"
	"testing"
)

func TestWalker_Walk(t *testing.T) {
	nodeA, nodeG := New()
	nodeA.Data = "A"
	nodeB := &Node{Data: "B"}
	nodeC := &Node{Data: "C"}
	nodeD := &Node{Data: "D"}
	nodeE := &Node{Data: "E"}
	nodeF := &Node{Data: "F"}
	nodeG.Data = "G"

	nodeB.Between(nodeA, nodeG)
	nodeC.Between(nodeA, nodeG)
	nodeD.Between(nodeA, nodeG)
	nodeF.Between(nodeB, nodeG)
	nodeF.Between(nodeC, nodeG)
	nodeE.Between(nodeB, nodeG)

	if len(nodeA.Children()) != 3 {
		t.Fatalf("node A wrong number of children: %d", len(nodeA.Children()))
	}
	if len(nodeC.Children()) != 1 {
		t.Fatalf("node C wrong number of children: %d", len(nodeC.Children()))
	}
	if len(nodeC.Parents()) != 1 {
		t.Fatalf("node C wrong number of parents: %d", len(nodeC.Parents()))
	}
	if len(nodeG.Parents()) != 3 {
		t.Fatalf("node G wrong number of parents: %d", len(nodeG.Parents()))
	}

	unique := make(map[string]int)
	w, err := NewWalker(nodeA)
	if err != nil {
		t.Fatal(err)
	}
	for {
		node, err := w.Walk()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatal(err)
		}

		node.MarkAsDone()

		if _, ok := unique[node.Data.(string)]; !ok {
			unique[node.Data.(string)] = 1
		} else {
			unique[node.Data.(string)]++
		}
	}

	for name, occurrences := range unique {
		if occurrences > 1 {
			t.Errorf("%s occurred more than once: %d", name, occurrences)
		}
	}
}
