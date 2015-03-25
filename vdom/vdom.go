package vdom

type Attributes map[string]interface{}

type EvtHandler func(Event)
type Event interface{}

type Node interface {
	IsElement() bool
	NodeData() string
}

type TextNode struct {
	Data string
}

func (t *TextNode) IsElement() bool {
	return false
}

func (t *TextNode) NodeData() string {
	return t.Data
}

func NewTextNode(data string) *TextNode {
	return &TextNode{Data: data}
}

type Element struct {
	Tag         string
	Attrs       Attributes
	Children    []Node
	EvtHandlers map[string]EvtHandler
	rendered    interface{}
}

func (t *Element) IsElement() bool {
	return true
}

func (t *Element) NodeData() string {
	return t.Tag
}

func NewElement(tag string, attrs Attributes, children []Node) *Element {
	return &Element{
		Tag:      tag,
		Attrs:    attrs,
		Children: children,
	}
}
