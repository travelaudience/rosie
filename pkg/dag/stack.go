package dag

type stackNode struct {
	dagNode       *Node
	nextStackNode *stackNode
}

type stack struct {
	top *stackNode
}

func (s *stack) pop() (*Node, bool) {
	if s.top == nil {
		return nil, false
	}

	n := s.top
	s.top = s.top.nextStackNode
	return n.dagNode, true
}

func (s *stack) push(t *Node) {
	nn := &stackNode{dagNode: t, nextStackNode: s.top}
	s.top = nn
}

func (s *stack) isEmpty() bool {
	return s.top == nil
}
