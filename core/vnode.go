package core

import (
	"strings"

	"github.com/phaikawl/wade/scope"
)

const (
	TextNode NodeType = 1 << iota
	MustacheNode
	ElementNode
	GroupNode
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
		attrs     map[string]interface{}
		Binds     []Bindage
		classes   map[string]bool
		scope     *scope.Scope
		callbacks []cbFunc
	}

	CondFn func(node VNode) bool
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

func VEmpty(tagName string) VNode {
	return VElem(tagName, NoAttr(), NoBind(), []VNode{})
}

func VWrap(tagName string, children []VNode) VNode {
	return VElem(tagName, NoAttr(), NoBind(), children)
}

func VText(text string) VNode {
	return VNode{
		Type:     TextNode,
		Data:     text,
		attrs:    NoAttr(),
		Binds:    NoBind(),
		Children: []VNode{},
	}
}

func VMustache(expr string) VNode {
	return VNode{
		Type:  MustacheNode,
		Data:  "",
		attrs: NoAttr(),
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
		attrs:    attrs,
		Binds:    binds,
		Children: children,
	}
}

func VElem(data string, attrs map[string]interface{}, binds []Bindage, children []VNode) VNode {
	return V(ElementNode, data, attrs, binds, children)
}

func (node VNode) TagName() string {
	if !(node.Type == ElementNode || node.Type == GroupNode) {
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
	v, ok = node.attrs[strings.ToLower(attr)]
	return
}

func (node *VNode) SetAttr(attr string, value interface{}) {
	node.attrs[strings.ToLower(attr)] = value
}

func (node *VNode) SetClass(className string, on bool) {
	if node.classes == nil {
		node.classes = map[string]bool{}
	}

	node.classes[className] = on
}

func (node VNode) HasClass(className string) bool {
	if node.classes == nil {
		return false
	}

	if has, ok := node.classes[className]; ok && has {
		return true
	}

	return false
}

func (node VNode) Attrs() map[string]interface{} {
	return node.attrs
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
	return node.CloneWithCond(nil)
}

func (node VNode) CloneWithCond(cond CondFn) (clone VNode) {
	if cond != nil && !cond(node) {
		return
	}

	clone = node
	clone.Children = make([]VNode, len(node.Children))
	for i, _ := range node.Children {
		clone.Children[i] = node.Children[i].Clone()
	}

	clone.attrs = make(map[string]interface{})
	for k, v := range node.attrs {
		clone.attrs[k] = v
	}

	if node.classes != nil {
		clone.classes = make(map[string]bool)
		for k, v := range node.classes {
			clone.classes[k] = v
		}
	}

	return
}
