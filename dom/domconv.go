package dom

import (
	"fmt"
	"regexp"

	"github.com/phaikawl/wade/core"
)

var (
	MustacheRegex = regexp.MustCompile("{{([^{}]+)}}")
)

type (
	PlatformNode interface{}

	Renderer interface {
		NewTextNode(vnode *core.VNode) PlatformNode
		NewElementNode(vnode *core.VNode, children []PlatformNode) PlatformNode
	}
)

func renderRec(v *core.VNode, renderer Renderer) []PlatformNode {
	children := []PlatformNode{}
	for i := range v.Children {
		c := &v.Children[i]
		rchilds := renderRec(c, renderer)

		if rchilds != nil {
			children = append(children, rchilds...)
		}
	}

	if v.Type == core.GroupNode {
		return children
	}

	//fmt.Printf("%v %v {%v}\n", utils.NoSp(v.Data), v.Type, v.Attrs)

	var n PlatformNode
	switch v.Type {
	case core.TextNode, core.MustacheNode:
		n = renderer.NewTextNode(v)
	case core.ElementNode:
		n = renderer.NewElementNode(v, children)
	case core.DataNode, core.DeadNode:
		return []PlatformNode{}
	case core.NotsetNode:
		panic(fmt.Errorf("node type not set for this node (nodeType=0)."))
	default:
		panic(fmt.Errorf("Invalid node type %v", v.Type))
	}

	return []PlatformNode{n}
}

func Render(v *core.VNode, renderer Renderer) PlatformNode {
	v.Prep()
	n := renderRec(v, renderer)
	return n[0]
}

func ParseMustaches(text string) []core.VNode {
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
