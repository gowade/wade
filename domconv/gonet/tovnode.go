package gonet

import (
	"bytes"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/phaikawl/wade/core"
	"github.com/phaikawl/wade/utils"
)

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

func createElement(tagName string) *html.Node {
	node := createNode()
	node.Type = html.ElementNode
	node.Data = tagName
	node.DataAtom = atom.Lookup([]byte(tagName))

	return node
}

func createTextNode(text string) *html.Node {
	node := createNode()
	node.Type = html.TextNode
	node.Data = text

	return node
}

func renderRec(v core.VNode) []*html.Node {
	children := []*html.Node{}
	for _, c := range v.Children {
		rchild := renderRec(c)
		if rchild != nil {
			children = append(children, rchild...)
		}
	}

	if v.Type == core.GroupNode {
		return children
	}

	var n *html.Node
	switch v.Type {
	case core.TextNode, core.MustacheNode:
		n = createTextNode(v.Data)
	case core.ElementNode:
		n = createElement(v.Data)
		for name, val := range v.Attrs() {
			var value string

			switch v := val.(type) {
			case string:
				value = v
			case bool:
				if !v {
					continue
				}

				value = ""
			default:
				value = utils.ToString(val)
			}

			n.Attr = append(n.Attr, html.Attribute{
				Key: name,
				Val: value,
			})
		}

		for _, c := range children {
			n.AppendChild(c)
		}
	case core.DeadNode:
		return []*html.Node{}
	case core.DataNode:
		n = createNode()
		n.Type = html.CommentNode
		n.Data = v.Data
	default:
		panic("Invalid type of node")
	}

	return []*html.Node{n}
}

func Render(r *html.Node, v core.VNode) {
	n := renderRec(v)
	*r = *n[0]
}

func getAttr(n *html.Node, attrName string) *html.Attribute {
	if n == nil {
		return nil
	}

	for i, a := range n.Attr {
		if a.Key == attrName {
			return &n.Attr[i]
		}
	}

	return nil
}

func tovnodeRec(node *html.Node) (result []core.VNode) {
	switch node.Type {
	case html.TextNode:
		return parseMustaches(node.Data)
	case html.CommentNode:
		return []core.VNode{
			core.V(core.DataNode, node.Data, core.NoAttr(), core.NoBind(), []core.VNode{}),
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

		n := core.VElem(node.Data, attrs, binds, []core.VNode{})

		if getAttr(node, core.GroupAttrName) != nil { // has "!group" attribute
			n.Type = core.GroupNode
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n.Children = append(n.Children, tovnodeRec(c)...)
		}

		return []core.VNode{n}
	}

	return []core.VNode{}
}

func ToVNode(r *html.Node) core.VNode {
	return tovnodeRec(r)[0]
}

var (
	MustacheRegex = regexp.MustCompile("{{([^{}]+)}}")
)

func parseMustaches(text string) []core.VNode {
	matches := MustacheRegex.FindAllStringSubmatch(text, -1)

	if matches == nil {
		return []core.VNode{core.VText(text)}
	}

	nodes := []core.VNode{}
	splitted := MustacheRegex.Split(text, -1)

	for i, m := range matches {
		if splitted[i] != "" {
			nodes = append(nodes, core.VText(splitted[i]))
		}
		nodes = append(nodes, core.VMustache(m[1]))
	}

	if splitted[len(splitted)-1] != "" {
		nodes = append(nodes, core.VText(splitted[len(splitted)-1]))
	}

	return nodes
}
