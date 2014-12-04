package core

import (
	"fmt"
	"strings"

	"github.com/phaikawl/wade/scope"
)

const (
	NotsetNode NodeType = iota
	TextNode
	MustacheNode
	ElementNode
	GroupNode
	DataNode
	DeadNode
)

const (
	AttrBind BindType = iota
	BinderBind
)

type (
	Attributes map[string]interface{}

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
		Attrs     Attributes
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

func PreprocessVNode(v *VNode) {
	if v.Type == NotsetNode {
		if v.Data == "" {
			panic("Uninitialized node detected, no node type and node data.")
		}

		v.Type = ElementNode
	}

	if v.Type != TextNode && v.Type != MustacheNode {
		if v.Attrs == nil {
			v.Attrs = make(map[string]interface{})
		}

		if v.Binds == nil {
			v.Binds = []Bindage{}
		}

		if v.Children == nil {
			v.Children = []VNode{}
		}
	}
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

func VText(text string) VNode {
	return VNode{
		Type:     TextNode,
		Data:     text,
		Attrs:    make(map[string]interface{}),
		Binds:    []Bindage{},
		Children: []VNode{},
	}
}

func VMustache(expr string) VNode {
	return VNode{
		Type:  MustacheNode,
		Data:  "",
		Attrs: make(map[string]interface{}),
		Binds: []Bindage{Bindage{
			Type: AttrBind,
			Expr: expr,
		}},
		Children: []VNode{},
	}
}

func (node VNode) Ptr() (np *VNode) {
	np = new(VNode)
	*np = node
	return np
}

func VPrep(node VNode) (r VNode) {
	r = node
	prepRec(&r)
	return
}

func prepRec(node *VNode) {
	PreprocessVNode(node)
	for i := range node.Children {
		prepRec(&node.Children[i])
	}
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
	v, ok = node.Attrs[strings.ToLower(attr)]
	return
}

func (node VNode) ChildElems() (l []*VNode) {
	l = []*VNode{}
	for i := range node.Children {
		item := &node.Children[i]
		if item.Type == ElementNode {
			l = append(l, item)
		}

		if item.Type == GroupNode {
			l = append(l, item.ChildElems()...)
		}
	}

	return
}

func (node *VNode) SetAttr(attr string, value interface{}) {
	node.Attrs[strings.ToLower(attr)] = value
}

func (node *VNode) ClassStr() (s string) {
	for className, enabled := range node.classes {
		if enabled {
			s += className + " "
		}
	}

	s = strings.TrimSpace(s)
	return
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

func (node VNode) FindTag(tagName string) (result []*VNode) {
	result = []*VNode{}
	NodeWalk(&node, func(node *VNode) {
		if node.Type == ElementNode && node.TagName() == tagName {
			result = append(result, node)
		}
	})

	return
}

func NodeDebug(node VNode, level int) {
	sp := ""
	for i := 0; i < level; i++ {
		sp += "  "
	}
	fmt.Print(sp)
	switch node.Type {
	case TextNode:
		fmt.Printf(`< "%v" >`, node.Data)
	case MustacheNode:
		fmt.Printf(`< {%v}"%v" >`, node.Binds[0].Name, node.Data)
	default:
		fmt.Printf("<%v:%v {%+v} [%v]>", node.Type, node.TagName(), node.Attrs, node.ClassStr())
	}
	fmt.Println()

	for i, _ := range node.Children {
		NodeDebug(node.Children[i], level+1)
	}
}

func NodeWalkX(node *VNode, fn func(*VNode, int)) {
	for i, _ := range node.Children {
		fn(node, i)
		NodeWalkX(&node.Children[i], fn)
	}
}

func NodeWalk(node *VNode, fn func(*VNode)) {
	PreprocessVNode(node)
	fn(node)
	for i, _ := range node.Children {
		NodeWalk(&node.Children[i], fn)
	}
}

func (node VNode) Clone() (clone VNode) {
	return node.CloneWithCond(nil)
}

func (node VNode) CloneWithCond(cond CondFn) (clone VNode) {
	clone = node
	PreprocessVNode(&clone)
	clone.Children = make([]VNode, 0)
	for i := range node.Children {
		if cond == nil || cond(node.Children[i]) {
			clone.Children = append(clone.Children,
				node.Children[i].CloneWithCond(cond))
		}
	}

	clone.Attrs = make(map[string]interface{})
	for k, v := range node.Attrs {
		clone.Attrs[k] = v
	}

	if node.classes != nil {
		clone.classes = make(map[string]bool)
		for k, v := range node.classes {
			clone.classes[k] = v
		}
	}

	return
}
