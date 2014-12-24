package gonet

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/phaikawl/wade/core"
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/utils"
)

var (
	gNodeMap = map[*html.Node]*core.VNode{}
)

func GetVNode(node *html.Node) *core.VNode {
	return gNodeMap[node]
}

func parseHtml(src string) (*html.Node, error) {
	nodes, err := html.ParseFragment(bytes.NewBufferString(strings.TrimSpace(src)), &html.Node{
		Type:     html.ElementNode,
		Data:     "body",
		DataAtom: atom.Body,
	})

	return nodes[0], err
}

func createNode() *html.Node {
	node, _ := parseHtml("d")
	return node
}

type htmlNode html.Node

func (n *htmlNode) ToVNode() (result []core.VNode) {
	node := n.node()
	switch node.Type {
	case html.TextNode:
		return dom.ParseMustaches(node.Data)
	case html.CommentNode:
		return []core.VNode{
			core.VPrep(core.VNode{Type: core.DataNode, Data: node.Data}),
		}
	case html.ElementNode:
		attrs := map[string]interface{}{}
		binds := []core.Bindage{}
		for _, attr := range node.Attr {
			var bindType core.BindType
			switch attr.Key[0] {
			case core.AttrBindPrefix:
				bindType = core.AttrBind
			case core.BinderBindPrefix:
				bindType = core.BinderBind
			default:
				attrs[attr.Key] = attr.Val
				continue
			}

			binds = append(binds, core.Bindage{
				Type: bindType,
				Name: attr.Key[1:],
				Expr: attr.Val,
			})
		}

		n := core.VNode{
			Data:     node.Data,
			Type:     core.ElementNode,
			Attrs:    attrs,
			Binds:    binds,
			Children: []core.VNode{},
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n.Children = append(n.Children, (*htmlNode)(c).ToVNode()...)
		}

		return []core.VNode{core.VPrep(n)}
	default:
		panic(fmt.Errorf(`Unhandled node type %v when
		converting HTML to VNode", node.Type`))
	}

	return []core.VNode{}
}

func (n *htmlNode) node() *html.Node {
	return (*html.Node)(n)
}

func record(node *html.Node, vnode *core.VNode) dom.PlatformNode {
	if gNodeMap != nil {
		gNodeMap[node] = vnode
	}

	vnode.Rendered = node
	return (*htmlNode)(node)
}

type Renderer struct{}

func (r Renderer) NewElementNode(vnode *core.VNode, children []dom.PlatformNode) dom.PlatformNode {
	node := createNode()
	node.Type = html.ElementNode
	node.Data = vnode.Data
	node.DataAtom = atom.Lookup([]byte(vnode.Data))

	renderAttrs(vnode, node)

	for _, c := range children {
		node.AppendChild(c.(*htmlNode).node())
	}

	return record(node, vnode)
}

func (r Renderer) NewTextNode(vnode *core.VNode) dom.PlatformNode {
	node := createNode()
	node.Type = html.TextNode
	node.Data = vnode.Data

	return record(node, vnode)
}

func renderAttrs(v *core.VNode, n *html.Node) {
	var classAttr *html.Attribute
	for name, val := range v.Attrs {
		var value string

		switch v := val.(type) {
		case string:
			value = v
		case bool:
			if !v {
				continue
			}

			value = ""
		case int, int32, int64, float32, float64:
			value = utils.ToString(val)
		default:
			continue
		}

		n.Attr = append(n.Attr, html.Attribute{
			Key: name,
			Val: value,
		})

		if name == "class" {
			classAttr = &n.Attr[len(n.Attr)-1]
		}
	}

	if clsStr := v.ClassStr(); clsStr != "" {
		if classAttr != nil {
			classAttr.Val = clsStr
		} else {
			n.Attr = append(n.Attr, html.Attribute{
				Key: "class",
				Val: clsStr,
			})
		}
	}
}

func Render(node *html.Node, v *core.VNode) {
	gNodeMap = make(map[*html.Node]*core.VNode)
	ptr := dom.Render(v, Renderer{}).(*htmlNode).node()
	*node = *ptr
}

func ToVNode(node *html.Node) core.VNode {
	return (*htmlNode)(node).ToVNode()[0]
}
