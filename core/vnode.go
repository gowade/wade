package core

import (
	"fmt"
	"strings"
	"unicode"
)

const (
	GroupNodeTagName   = "w_group"
	IncludeTagName     = "w_include"
	ComponentTagName   = "w_component"
	ComponentTagPrefix = "c:"
)

const (
	NotsetNode NodeType = iota
	MustacheNode
	TextNode
	ElementNode
	GroupNode
	DeadNode
)

type (
	Attributes map[string]interface{}

	NodeType uint

	BindFunc func(*VNode)

	VNode struct {
		Type     NodeType
		Data     string
		Children []*VNode
		Attrs    Attributes
		classes  map[string]bool
		Binds    []BindFunc
		Rendered interface{} // data field to save the real rendered DOM element
	}

	CondFn func(node *VNode) bool
)

func (node *VNode) addBind(cb BindFunc) {
	if node.Binds == nil {
		node.Binds = []BindFunc{cb}
		return
	}

	node.Binds = append(node.Binds, cb)
}

func preprocessVNode(v *VNode) {
	if v.Type == NotsetNode {
		if v.Data == "" {
			panic("Uninitialized node detected, no node type and node data.")
		}

		v.Type = ElementNode
	}

	if v.Type != TextNode && v.Type != MustacheNode {
		if v.Data == GroupNodeTagName {
			v.Type = GroupNode
		}

		if v.Attrs == nil {
			v.Attrs = make(map[string]interface{})
		}

		if v.Children == nil {
			v.Children = []*VNode{}
		}

		v.processClassAttr()
	}
}

func trimSpace(text string) string {
	start := 0
	r := []rune(text)
	for {
		if start == len(r)-1 || !unicode.IsSpace(r[start]) {
			break
		}

		start++
	}

	end := len(r)
	for {
		if end <= start || !unicode.IsSpace(r[end-1]) {
			break
		}

		end--
	}

	preSp := ""
	if start > 0 {
		preSp = " "
	}

	postSp := ""
	if end < len(r) {
		postSp = " "
	}

	//println(start, end, len(r))
	return preSp + string(r[start:end]) + postSp
}

func VText(text string) *VNode {
	return &VNode{
		Type:     TextNode,
		Data:     trimSpace(text),
		Attrs:    make(map[string]interface{}),
		Children: []*VNode{},
	}
}

func VElem(tagName string, class string) *VNode {
	return &VNode{
		Type: ElementNode,
		Data: tagName,
		Attrs: Attributes{
			"class": class,
		},
	}
}

func VMustacheInfo(text string) *VNode {
	return &VNode{
		Type:     MustacheNode,
		Data:     trimSpace(text),
		Attrs:    make(map[string]interface{}),
		Children: []*VNode{},
	}
}

func VMustache(expr func() interface{}) *VNode {
	return &VNode{
		Type:     MustacheNode,
		Data:     "",
		Attrs:    make(map[string]interface{}),
		Children: []*VNode{},
		Binds: []BindFunc{
			func(vn *VNode) {
				vn.Data = fmt.Sprint(expr())
			},
		},
	}
}

func VPrep(node *VNode) *VNode {
	preprocessVNode(node)
	for _, c := range node.Children {
		VPrep(c)
	}

	return node
}

func VComponent(initFn func() (*VNode, func(node *VNode))) *VNode {
	node, udtFn := initFn()
	if node.Binds == nil {
		node.Binds = []BindFunc{}
	}

	node.Binds = append(node.Binds, udtFn)
	return node
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
	node.UpdateCond(nil)
}

func (node *VNode) UpdateCond(cond CondFn) {
	if cond != nil && !cond(node) {
		return
	}

	if node.Binds != nil {
		for _, cb := range node.Binds {
			//gopherjs:blocking
			cb(node)
		}
	}

	for _, c := range node.Children {
		c.UpdateCond(cond)
	}
}

func (node VNode) Attr(attr string) (v interface{}, ok bool) {
	v, ok = node.Attrs[strings.ToLower(attr)]
	return
}

func (node VNode) ChildElems() (l []*VNode) {
	l = []*VNode{}
	for _, item := range node.Children {
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
	return nodeDebug(&node, 0)
}

func nodeDebug(node *VNode, level int) (s string) {
	if node == nil {
		return "<nil>"
	}

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
		s += fmt.Sprintf(`{"%v"}`, node.Data)
	default:
		s += fmt.Sprintf("<%v:%v {%+v} [%v]>", node.Data, suffix, node.Attrs, node.ClassStr())
	}
	s += "\n"

	for _, c := range node.Children {
		s += nodeDebug(c, level+1)
	}

	return
}

func NodeWalkX(node *VNode, fn func(*VNode, int)) {
	for i, c := range node.Children {
		fn(node, i)
		NodeWalkX(c, fn)
	}
}

func NodeWalk(node *VNode, fn func(*VNode)) {
	fn(node)
	for _, c := range node.Children {
		NodeWalk(c, fn)
	}
}

func (node VNode) Clone() (clone *VNode) {
	return node.CloneWithCond(nil)
}

func (node VNode) CloneWithCond(cond CondFn) (clone *VNode) {
	clone = new(VNode)
	*clone = node
	preprocessVNode(clone)
	clone.Children = make([]*VNode, 0)
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
