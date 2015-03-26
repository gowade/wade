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

func (m TreeModifier) renderNode(node vdom.Node) *js.Object {
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
		newElem.Call("appendChild", m.renderNode(c))
	}

	return newElem
}

func (m TreeModifier) render(node vdom.Node, d *js.Object) {
	if !node.IsElement() {
		d.Set("nodeValue", node.NodeData())
		return
	}

	d.Get("parentNode").Call("replaceChild", m.renderNode(node), d)
}

func (m TreeModifier) Do(dNode vdom.DomNode, action vdom.Action) {
	d := dNode.(DomNode).Object

	switch action.Type {
	case vdom.Deletion:
		d.Call("removeChild", d.Get("childNodes").Index(action.Index))
	case vdom.Insertion:
		insertee := m.renderNode(action.Content)
		if action.Index == -1 {
			d.Call("appendChild", insertee)
		} else {
			d.Call("insertBefore", insertee, d.Get("childNodes").Index(action.Index))
		}
	case vdom.Move:
		d.Call("insertBefore", action.Element, d.Get("childNodes").Index(action.Index))
	case vdom.Update:
		m.render(action.Content, d)
	}
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
