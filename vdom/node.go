package vdom

type Attributes map[string]interface{}

type EvtHandler func(Event)
type Event interface{}

type Node interface {
	NodeData() string
	Text() string
}

type TextNode struct {
	Data string
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

type Element struct {
	Tag      string
	Key      string
	Attrs    Attributes
	Children []Node
	//EvtHandlers map[string]EvtHandler
}

func (t *Element) Text() string {
	s := ""
	for _, c := range t.Children {
		s += c.Text()
	}

	return s
}

func (t *Element) NodeData() string {
	return t.Tag
}

func NewElement(tag, key string, attrs Attributes, children []Node) *Element {
	return &Element{
		Tag:      tag,
		Key:      key,
		Attrs:    attrs,
		Children: children,
	}
}
