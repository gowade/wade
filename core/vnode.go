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
		Rendered  interface{} // data field to save the real rendered DOM element
		preUpdate bool
		metaAttrs Attributes
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

func preprocessVNode(v *VNode) {
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

		v.processClassAttr()

		if v.metaAttrs == nil {
			v.metaAttrs = make(Attributes)
		}

		for attr, value := range v.Attrs {
			if string(attr[0]) == MetaAttrPrefix {
				delete(v.Attrs, attr)

				if attr == "!group" {
					v.Type = GroupNode
					continue
				}

				v.metaAttrs[attr[1:]] = value
			}
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
	preprocessVNode(node)
	for i := range node.Children {
		prepRec(&node.Children[i])
	}
}

func (node *VNode) processClassAttr() {
	if class, ok := node.Attr("class"); ok {
		if node.classes == nil {
			node.classes = map[string]bool{}
		}

		classes := strings.Split(class.(string), " ")
		for _, cls := range classes {
			node.classes[cls] = true
		}
	}
}

func (node *VNode) Prep() {
	prepRec(node)
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
	node.update(false)
}

func (node *VNode) update(preUpdate bool) {
	if node.preUpdate && !preUpdate {
		return
	}

	if node.callbacks != nil {
		for _, cb := range node.callbacks {
			//gopherjs:blocking
			err := cb()
			if err != nil {
				go func() {
					panic(err)
				}()
			}
		}
	}

	for i, _ := range node.Children {
		c := &node.Children[i]
		c.update(preUpdate)
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

func (node VNode) IsElement() bool {
	return node.Type == ElementNode || node.Type == GroupNode
}

func (node *VNode) SetAttr(attr string, value interface{}) {
	node.Attrs[strings.ToLower(attr)] = value
}

func (node *VNode) MetaAttr(attr string) (value interface{}, ok bool) {
	value, ok = node.metaAttrs[attr]
	return
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

func (node VNode) DebugInfo() string {
	return nodeDebug(node, 0)
}

func nodeDebug(node VNode, level int) (s string) {
	sp := ""
	for i := 0; i < level; i++ {
		sp += "  "
	}

	s += sp
	suffix := ""
	if node.Type == GroupNode {
		suffix = "GROUP"
	}

	if node.Type == DeadNode {
		suffix = "DEAD"
	}

	switch node.Type {
	case TextNode:
		text := strings.TrimSpace(node.Data)
		if text != "" {
			s += fmt.Sprintf(`"%v"`, text)
		}
	case MustacheNode:
		s += fmt.Sprintf(`{{%v}"%v" }`, node.Binds[0].Expr, node.Data)
	case DataNode:
		s += fmt.Sprintf(`//%v`, node.Data)
	default:
		s += fmt.Sprintf("<%v:%v {%+v} [%v]>", node.Data, suffix, node.Attrs, node.ClassStr())
	}
	s += "\n"

	for i, _ := range node.Children {
		s += nodeDebug(node.Children[i], level+1)
	}

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
	for i := range node.Children {
		NodeWalk(&node.Children[i], fn)
	}
}

func (node VNode) Clone() (clone VNode) {
	return node.CloneWithCond(nil)
}

func (node VNode) CloneWithCond(cond CondFn) (clone VNode) {
	clone = node
	preprocessVNode(&clone)
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
