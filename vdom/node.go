package vdom

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
	OnMount()
	OnUnmount()

	InternalState() interface{}
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
		if t.ComRend != nil {
			t.ComRend.Key = t.Key
			return t.ComRend
		}

		return nil
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
