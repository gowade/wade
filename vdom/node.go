package vdom

import ()

type Attributes map[string]interface{}

type EvtHandler func(Event)
type Event interface{}

type Node interface {
	IsElement() bool
	NodeData() string
	Render() Node
	Text() string
}

type TextNode struct {
	Data string
}

func (t *TextNode) IsElement() bool {
	return false
}

func (t *TextNode) Text() string {
	return t.Data
}

func (t *TextNode) NodeData() string {
	return t.Data
}

func (t *TextNode) Render() Node {
	return t
}

func NewTextNode(data string) *TextNode {
	return &TextNode{Data: data}
}

type Component interface {
	SetVNode(node *Element)
	Render(interface{}) *Element

	BeforeMount()
	AfterMount()
	OnUnmount()
	OnUpdated()

	InternalState() interface{}
	InternalInitState(interface{})
	InternalUnmount()
	InternalUnmounted() bool
}

type Element struct {
	Tag         string
	Attrs       Attributes
	Children    []Node
	EvtHandlers map[string]EvtHandler
	Component   Component
	Key         string

	comref     Component
	domNode    DOMNode // the rendered node in DOM
	OnRendered func(DOMNode)

	ComRend *Element
	oldElem *Element
}

func (t *Element) Text() string {
	s := ""
	for _, c := range t.Children {
		if c != nil {
			r := c.Render()
			if r != nil {
				s += r.Text()
			}
		}
	}

	return s
}

func (t *Element) DOMNode() DOMNode {
	if t.Component != nil && t.ComRend != nil {
		return t.ComRend.DOMNode()
	}
	return t.domNode
}

func (t *Element) SetRenderedDOMNode(node DOMNode) {
	t.domNode = node
	if t.OnRendered != nil {
		t.OnRendered(t.domNode)
	}
}

func (t *Element) Unmount() {
	if t.Component != nil {
		t.Component.InternalUnmount()
	}
}

func (t *Element) IsElement() bool {
	return true
}

func (t *Element) NodeData() string {
	return t.Tag
}

func (t *Element) Render() Node {
	if t.Component != nil {
		if t.ComRend != nil {
			return t.ComRend
		}

		var state interface{}
		if t.oldElem != nil && !t.Component.InternalUnmounted() && t.oldElem.Component != nil {
			state = t.oldElem.Component.InternalState()
		}

		t.ComRend = t.Component.Render(state)
		t.ComRend.comref = t.Component
		if t.ComRend != nil {
			t.ComRend.Key = t.Key
			if t.ComRend.Component != nil {
				t.ComRend = t.ComRend.Render().(*Element)
			}
			return t.ComRend
		}

		return nil
	}

	return t
}

func NewComElement(comName, key string, com Component, initFn func(interface{})) *Element {
	el := &Element{
		Tag:       comName,
		Key:       key,
		Component: com,
	}
	com.InternalInitState(nil)
	initFn(com)
	com.SetVNode(el)

	return el
}

func NewElement(tag, key string, attrs Attributes, children []Node) *Element {
	return &Element{
		Tag:      tag,
		Key:      key,
		Attrs:    attrs,
		Children: children,
	}
}

func Debug(node Node) {
	debug(">>>", node, 0)
}

func debug(prefix string, node Node, depth int) {
	var sp string
	for i := 0; i < depth; i++ {
		sp += "  "
	}

	if e, ok := node.(*Element); ok {
		e = e.Render().(*Element)
		println(prefix+sp+e.Tag, e.Attrs["class"], e.Attrs["id"], e.Attrs)
		for _, c := range e.Children {
			debug("", c, depth+1)
		}
	} else {
		println(sp + `"` + node.NodeData() + `"`)
	}
}
