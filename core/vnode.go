package core

import (
	"strings"

	"github.com/phaikawl/wade/scope"
)

const (
	TextNode NodeType = 1 << iota
	MustacheNode
	ElementNode
	GhostNode
	DataNode
	DeadNode
)

const (
	AttrBind BindType = 1 << iota
	BinderBind
)

type (
	Bindage struct {
		Type BindType
		Name string
		Expr string
	}

	NodeType uint
	BindType uint

	cbFunc func() error

	VNode struct {
		Type      NodeType
		Data      string
		Children  []VNode
		Attrs     map[string]interface{}
		Binds     []Bindage
		scope     *scope.Scope
		callbacks []cbFunc
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

func (node *VNode) addCallback(cb cbFunc) {
	if node.callbacks == nil {
		node.callbacks = []cbFunc{cb}
		return
	}

	node.callbacks = append(node.callbacks, cb)
}

func BindBinder(name, expr string) Bindage {
	return Bindage{
		Type: BinderBind,
		Name: name,
		Expr: expr,
	}
}

func BindAttr(name, expr string) Bindage {
	return Bindage{
		Type: AttrBind,
		Name: name,
		Expr: expr,
	}
}

func NoAttr() map[string]interface{} {
	return map[string]interface{}{}
}

func NoBind() []Bindage {
	return []Bindage{}
}

func VText(text string) VNode {
	return VNode{
		Type:     TextNode,
		Data:     text,
		Attrs:    NoAttr(),
		Binds:    NoBind(),
		Children: []VNode{},
	}
}

func VMustache(expr string) VNode {
	return VNode{
		Type:  MustacheNode,
		Data:  expr,
		Attrs: NoAttr(),
		Binds: []Bindage{Bindage{
			Type: AttrBind,
			Expr: expr,
		}},
		Children: []VNode{},
	}
}

func NodeRoot(node VNode) (np *VNode) {
	np = new(VNode)
	*np = node
	return np
}

func V(typ NodeType, data string, attrs map[string]interface{}, binds []Bindage, children []VNode) VNode {
	return VNode{
		Type:     typ,
		Data:     data,
		Attrs:    attrs,
		Binds:    binds,
		Children: children,
	}
}

func VElem(data string, attrs map[string]interface{}, binds []Bindage, children []VNode) VNode {
	return V(ElementNode, data, attrs, binds, children)
}

func (node VNode) TagName() string {
	if !(node.Type == ElementNode || node.Type == GhostNode) {
		return ""
	}

	return node.Data
}

func (node VNode) Text() (s string) {
	if node.Type == TextNode || node.Type == MustacheNode {
		return node.Data
	}

	for _, c := range node.Children {
		s += c.Text()
	}

	return
}

func (node *VNode) Update() {
	if node.callbacks != nil {
		for _, cb := range node.callbacks {
			err := cb()
			if err != nil {
				go func() {
					panic(err)
				}()
			}
		}
	}

	for i, _ := range node.Children {
		(&node.Children[i]).Update()
	}
}

func (node VNode) Attr(attr string) (v interface{}, ok bool) {
	v, ok = node.Attrs[strings.ToLower(attr)]
	return
}

func NodeWalkX(node *VNode, fn func(*VNode, int)) {
	for i, _ := range node.Children {
		fn(node, i)
		NodeWalkX(&node.Children[i], fn)
	}
}

func NodeWalk(node *VNode, fn func(*VNode)) {
	fn(node)
	for i, _ := range node.Children {
		NodeWalk(&node.Children[i], fn)
	}
}

func (node VNode) Clone() (clone VNode) {
	clone = node
	clone.Children = make([]VNode, len(node.Children))
	for i, _ := range node.Children {
		clone.Children[i] = node.Children[i].Clone()
	}

	clone.Attrs = make(map[string]interface{})
	for k, v := range node.Attrs {
		clone.Attrs[k] = v
	}

	return
}
