package vdom

func NewNodeList(nodes ...interface{}) []Node {
	var l []Node
	for _, n := range nodes {
		switch n := n.(type) {
		case []Node:
			l = append(l, n...)
		case *Element, *TextNode:
			l = append(l, n.(Node))
		default:
			panic("Invalid node type")
		}
	}

	return l
}
