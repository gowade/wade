package browser

import (
	"fmt"
	"strings"

	"github.com/gopherjs/gopherjs/js"

	wdrv "github.com/gowade/wade/driver"
	"github.com/gowade/wade/vdom"
)

var (
	document *js.Object
)

type DOMInputEl struct{ *js.Object }

func (e DOMInputEl) Checked() bool {
	return e.Get("checked").Bool()
}

func (e DOMInputEl) SetChecked(checked bool) {
	e.Set("checked", checked)
}

func (e DOMInputEl) Value() string {
	return e.Get("value").String()
}

func (e DOMInputEl) SetValue(value string) {
	e.Set("value", value)
}

func init() {
	if js.Global == nil || js.Global.Get("document") == js.Undefined {
		panic("This package is only available in browser environment.")
	}

	wdrv.SetVdomDriver(driver{})
	document = js.Global.Get("document")
}

func ElementById(id string) vdom.DOMNode {
	elem := document.Call("getElementById", id)
	if elem == js.Undefined || elem == nil {
		panic(fmt.Sprintf("No element with id %v found", id))
	}

	return DOMNode{elem}
}

type driver struct{}

func (d driver) ToInputEl(el vdom.DOMNode) vdom.DOMInputEl {
	return DOMInputEl{el.(DOMNode).Object}
}

func createElement(tag string) *js.Object {
	return document.Call("createElement", tag)
}

func createTextNode(data string) *js.Object {
	return document.Call("createTextNode", data)
}

type DOMNode struct {
	*js.Object
}

func (d DOMNode) Compat(node vdom.Node) bool {
	nt := d.Get("nodeType").Int()
	switch nt {
	case 1:
		return node.IsElement()
	case 3:
		return !node.IsElement()
	}

	return false
}

func (d DOMNode) Child(i int) vdom.DOMNode {
	c := d.Get("childNodes").Index(i)
	if c == nil || c == js.Undefined {
		c = createElement("div")
		d.Call("appendChild", c)
	}
	return DOMNode{c}
}

func render(node vdom.Node, d *js.Object) *js.Object {
	if !node.IsElement() {
		if d == nil {
			return createTextNode(node.NodeData())
		} else {
			d.Set("nodeValue", node.NodeData())
			return d
		}
	}

	oe := node.(*vdom.Element)
	if oe.Component != nil {
		vdom.InternalRenderLock()
		oe.Component.BeforeMount()
		vdom.InternalRenderUnlock()
	}

	nr := oe.Render()
	if nr == nil {
		d = document.Call("createComment", "")
		oe.SetRenderedDOMNode(DOMNode{d})
		return d
	}

	e := nr.(*vdom.Element)
	if d == nil || d.Get("nodeType").Int() != 1 {
		d = createElement(e.Tag)
	}

	for attr, v := range e.Attrs {
		if vdom.IsEvent(attr) {
			d.Set(strings.ToLower(attr), v)
			continue
		}

		switch v := v.(type) {
		case bool:
			if v {
				d.Call("setAttribute", attr, attr)
			}
		case string:
			d.Call("setAttribute", attr, v)
		default:
			d.Call("setAttribute", attr, fmt.Sprint(v))
		}
	}

	for _, c := range e.Children {
		if c != nil {
			cr := render(c, nil)
			if cr != nil {
				d.Call("appendChild", cr)
			}
		}
	}

	e.SetRenderedDOMNode(DOMNode{d})
	if oe != e {
		oe.SetRenderedDOMNode(DOMNode{d})
	}

	if oe.Component != nil {
		oe.Component.AfterMount()
	}

	return d
}

func (d DOMNode) JS() *js.Object {
	return d.Object
}

func (d DOMNode) Render(content vdom.Node, root bool) {
	if !root {
		d.Get("parentNode").Call("replaceChild", render(content, d.Object), d.Object)
	} else {
		if !content.IsElement() {
			panic("root render being called on a text node.")
		}

		render(content, d.Object)
	}
}

func (dNode DOMNode) Do(action *vdom.Action) {
	d := dNode.Object

	switch action.Type {
	case vdom.Deletion:
		if el, ok := action.Content.(*vdom.Element); ok {
			if el.Component != nil {
				el.Component.OnUnmount()
			}
		}

		d.Call("removeChild", action.Element.(DOMNode).Object)
	case vdom.Insertion:
		insertee := render(action.Content, nil)
		if action.Index == -1 {
			d.Call("appendChild", insertee)
		} else {
			d.Call("insertBefore", insertee, d.Get("childNodes").Index(action.Index))
		}
	case vdom.Move:
		d.Call("insertBefore", action.Element.(DOMNode).Object, d.Get("childNodes").Index(action.Index))
	}
}

func (dNode DOMNode) RemoveAttr(attr string) {
	dNode.Object.Call("removeAttribute", attr)
}

func (dNode DOMNode) SetProp(prop string, value interface{}) {
	dNode.Object.Set(prop, value)
}

func (d DOMNode) Clear() {
	var c *js.Object
	for {
		c = d.Get("lastChild")
		if c == nil {
			return
		}

		d.Call("removeChild", c)
	}
}

func (dNode DOMNode) SetAttr(attr string, value interface{}) {
	d := dNode.Object

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
