package jsdom

import (
	"strings"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gowade/wade/utils/dom"
)

func init() {
	if js.Global == nil || js.Global.Get("document") == js.Undefined {
		panic("jsdom package can only be imported in browser environment")
	}

	dom.SetDocument(Node{js.Global.Get("document")})
}

type Node struct {
	*js.Object
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
