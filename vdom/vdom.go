package vdom

import "github.com/gopherjs/gopherjs/js"

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
	Render(interface{}) *Element
	InternalState() interface{}
}

type Element struct {
	Tag         string
	Attrs       Attributes
	Children    []Node
	EvtHandlers map[string]EvtHandler
	Component   Component
	rendCache   *Element
	oldElem     *Element
	Key         string

	domNode    *js.Object // the rendered node in DOM
	OnRendered func(*js.Object)
}

func (t *Element) Text() string {
	s := ""
	for _, c := range t.Children {
		if c != nil {
			s += c.Text()
		}
	}

	return s
}

func (t *Element) DOMNode() *js.Object {
	return t.domNode
}

func (t *Element) SetRenderedDOMNode(node *js.Object) {
	t.domNode = node
	if t.OnRendered != nil {
		t.OnRendered(t.domNode)
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
		if t.rendCache != nil {
			return t.rendCache
		}

		var state interface{}
		if t.oldElem != nil && t.oldElem.Component != nil {
			state = t.oldElem.Component.InternalState()
		}

		t.rendCache = t.Component.Render(state)
		t.rendCache.Key = t.Key
		return t.rendCache
	}

	return t
}

func NewComElement(comName, key string, com Component) *Element {
	return &Element{
		Tag:       comName,
		Key:       key,
		Component: com,
	}
}

func NewElement(tag, key string, attrs Attributes, children []Node) *Element {
	return &Element{
		Tag:      tag,
		Key:      key,
		Attrs:    attrs,
		Children: children,
	}
}
