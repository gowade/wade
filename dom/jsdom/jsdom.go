package jsdom

import (
	"fmt"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gowade/wade/dom"
)

type Event struct{ *js.Object }

func (e Event) JS() *js.Object {
	return e.Object
}

func (e Event) PreventDefault() {
	e.Call("preventDefault")
}

func (e Event) StopPropagation() {
	e.Call("stopPropagation")
}

func newEvent(evt *js.Object) dom.Event {
	return Event{evt}
}

func newEventHandler(handler dom.EventHandler) interface{} {
	return func(evt *js.Object) {
		handler(newEvent(evt))
	}
}

type driver struct{}

func (d driver) CreateNode(native interface{}) dom.Node {
	node := Node{native.(*js.Object)}
	switch node.Data() {
	case "input":
		return InputEl{node}
	case "form":
		return FormEl{node}
	}

	return node
}

func init() {
	if js.Global == nil || js.Global.Get("document") == js.Undefined {
		panic("jsdom package can only be imported in browser environment")
	}

	dom.SetDocument(Document{Node{js.Global.Get("document")}})
	dom.SetDomDriver(driver{})
	dom.NewEventHandler = newEventHandler
}

type Node struct {
	*js.Object
}

func (z Node) JS() *js.Object {
	return z.Object
}

func (z Node) Type() dom.NodeType {
	ntype := z.Get("nodeType").Int()
	switch ntype {
	case 1:
		return dom.ElementNode
	case 3:
		return dom.TextNode
	default:
		return dom.NopNode
	}
}

func (z Node) Data() string {
	switch z.Type() {
	case dom.ElementNode:
		return strings.ToLower(z.Get("tagName").String())
	default:
		return z.Get("nodeValue").String()
	}
}

func nodeList(jslist *js.Object) []dom.Node {
	n := jslist.Length()
	l := make([]dom.Node, 0, n)
	for i := 0; i < n; i++ {
		l = append(l, Node{jslist.Index(i)})
	}

	return l
}

func (z Node) Clear() {
	var c *js.Object
	for {
		c = z.Get("lastChild")
		if c == nil {
			return
		}

		z.Call("removeChild", c)
	}
}

func (z Node) Children() []dom.Node {
	cs := z.Get("childNodes")
	if cs == nil {
		panic("childNodes not available for this node")
	}

	return nodeList(cs)
}

func (z Node) Find(query string) []dom.Node {
	qselector := z.Get("querySelectorAll")
	if qselector == nil {
		panic("querySelectorAll not available for this node")
	}

	return nodeList(z.Call("querySelectorAll", query))
}

func (d Node) SetAttr(attr string, value interface{}) {
	var vstr string
	switch v := value.(type) {
	case bool:
		if !v {
			if d.Call("hasAttribute", attr).Bool() {
				d.Call("removeAttribute", attr)
			}

			return
		} else {
			vstr = attr
		}

	case string:
		vstr = v
	default:
		vstr = fmt.Sprint(v)
	}

	d.Call("setAttribute", attr, vstr)
}

func (z Node) RemoveAttr(attr string) {
	z.Call("removeAttribute", attr)
}

func (z Node) SetProp(prop string, value interface{}) {
	z.Set(prop, value)
}

func (z Node) SetClass(class string, val bool) {
	if val {
		z.Call("addClass", class)
	} else {
		z.Call("removeClass", class)
	}
}

type Document struct {
	Node
}

func (z Document) Title() string {
	return z.Get("title").String()
}

func (z Document) SetTitle(title string) {
	z.Set("title", title)
}
