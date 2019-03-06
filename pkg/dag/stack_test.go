package dag

import "testing"

func TestStack(t *testing.T) {
	s := &stack{}
	given := []*Node{
		{Data: "1"},
		{Data: "2"},
		{Data: "3"},
		{Data: "4"},
		{Data: "5"},
	}

	for _, g := range given {
		s.push(g)
	}

	for i := range given {
		expected := given[len(given)-i-1].Data.(string)
		if node, ok := s.pop(); !ok || node.Data.(string) != expected {
			t.Errorf("wrong first value, expect %s but got %s", expected, node.Data)
		}
	}
}
