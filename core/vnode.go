package core

import "github.com/phaikawl/wade/scope"

const (
	TextNode NodeType = 1 << iota
	ElementNode
	GhostNode
	DataNode
)

type (
	Bindage struct {
		Name string
		Expr string
	}

	NodeType uint

	VNode struct {
		Type       NodeType
		Data       string
		Children   []Node
		Attrs      map[string]interface{}
		Binds      []Bindage
		scope      *scope.Scope
		rerenderCb func(VNode)
	}

	Document struct {
		Root VNode
		RealDom
	}

	RealDom interface {
		Render(VNode)
		ToVNode() VNode
	}
)

func NodeWalk(node *VNode, fn func(*VNode, int)) {
	fn(node)
	for i, _ := range node.Children {
		NodeWalk(node, i)
	}
}

func NodeClone(node Node) (clone Node) {
	clone = node
	clone.Children = make([]Node, len(node.Children))
	for i, _ := range node.Children {
		clone.Children[i] = NodeClone(node.Children[i])
	}

	clone.Attrs = make(map[string]interface{})
	for k, v := range node.Attrs {
		clone.Attrs[k] = v
	}

	return
}
