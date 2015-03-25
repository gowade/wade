package browser

import (
	"fmt"

	"github.com/gopherjs/gopherjs/js"

	"github.com/gowade/wade/vdom"
)

var (
	document     = js.Global.Get("document")
	treeModifier = TreeModifier{}
)

func NewModifier() TreeModifier {
	return TreeModifier{}
}

func createElement(tag string) *js.Object {
	return document.Call("createElement", tag)
}

func createTextNode(data string) *js.Object {
	return document.Call("createTextNode", data)
}

type DomNode struct {
	*js.Object
}

func (d DomNode) Child(i int) vdom.DomNode {
	return DomNode{d.Get("childNodes").Index(i)}
}

type TreeModifier struct{}

func (m TreeModifier) renderElem(node vdom.Node) *js.Object {
	if !node.IsElement() {
		return createTextNode(node.NodeData())
	}

	e := node.(*vdom.Element)
	newElem := createElement(e.Tag)
	for attr, v := range e.Attrs {
		switch v := v.(type) {
		case bool:
			if v {
				newElem.Call("setAttribute", attr, attr)
			}
		case string:
			newElem.Call("setAttribute", attr, v)
		default:
			newElem.Call("setAttribute", attr, fmt.Sprint(v))
		}
	}

	for _, c := range e.Children {
		newElem.Call("appendChild", m.renderElem(c))
	}

	return newElem
}

func (m TreeModifier) Render(node vdom.Node, domNode vdom.DomNode) {
	d := domNode.(DomNode).Object
	if !node.IsElement() {
		d.Set("nodeValue", node.NodeData())
		return
	}

	d.Get("parentNode").Call("replaceChild", m.renderElem(node), d)
}

func (m TreeModifier) Insert(node vdom.Node, parent vdom.DomNode) {
	p := parent.(DomNode).Object
	p.Call("appendChild", m.renderElem(node))
}

func (m TreeModifier) Delete(domNode vdom.DomNode) {
	d := domNode.(DomNode).Object
	d.Get("parentNode").Call("removeChild", d)
}

func (m TreeModifier) RemoveAttr(dNode vdom.DomNode, attr string) {
	dNode.(DomNode).Call("removeAttribute", attr)
}

func (m TreeModifier) SetAttr(dNode vdom.DomNode, attr string, value interface{}) {
	d := dNode.(DomNode).Object

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

func PerformDiff(a, b *vdom.Element, root *js.Object) {
	if root.Get("childNodes").Get("length").Int() == 0 || b == nil {
		root.Call("appendChild", createElement(a.Tag))
	}

	vdom.PerformDiff(a, b, DomNode{root.Get("childNodes").Index(0)}, treeModifier)
}
