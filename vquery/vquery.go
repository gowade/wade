package vq

import (
	"reflect"
	"strings"

	"github.com/phaikawl/wade/core"
)

var (
	gParent = map[*core.VNode]*core.VNode{}
)

type Selection []*core.VNode

type M map[string]interface{}

type Selector struct {
	Id       string
	Tag      string
	Class    string
	Classes  []string
	HasAttrs []string
	Attrs    M
}

func (s Selector) Match(node *core.VNode) bool {
	if !node.IsElement() {
		if s.Id != "" || s.Tag != "" || s.Class != "" || s.Classes != nil || s.HasAttrs != nil || s.Attrs != nil {
			return false
		}
	}

	if s.Id != "" {
		if id, ok := node.Attr("id"); !ok || s.Id != id {
			return false
		}
	}

	if s.Tag != "" {
		if strings.ToLower(s.Tag) != strings.ToLower(node.Data) {
			return false
		}
	}

	if s.Class != "" {
		if !node.HasClass(s.Class) {
			return false
		}
	}

	if s.Classes != nil {
		for _, class := range s.Classes {
			if !node.HasClass(class) {
				return false
			}
		}
	}

	if s.Attrs != nil {
		for attr, wantv := range s.Attrs {
			if gotv, ok := node.Attrs[attr]; !ok || !reflect.DeepEqual(wantv, gotv) {
				return false
			}
		}
	}

	if s.HasAttrs != nil {
		for attr := range node.Attrs {
			if _, ok := s.Attrs[attr]; !ok {
				return false
			}
		}
	}

	return true
}

func New(vnode *core.VNode) Selection {
	core.NodeWalkX(vnode, func(parent *core.VNode, i int) {
		c := &parent.Children[i]
		gParent[c] = parent
	})

	return []*core.VNode{vnode}
}

func (s Selection) Find(selector Selector) Selection {
	result := []*core.VNode{}
	for _, node := range s {
		core.NodeWalk(node, func(node *core.VNode) {
			if selector.Match(node) {
				result = append(result, node)
			}
		})
	}

	return Selection(result)
}

func (s Selection) First() *core.VNode {
	return s[0]
}

func (s Selection) Children() Selection {
	children := []*core.VNode{}
	for _, e := range s {
		children = append(children, e.ChildElems()...)
	}

	return children
}

func (s Selection) Text() (text string) {
	for _, node := range s {
		text += node.Text()
	}

	return
}

func Parent(node *core.VNode) *core.VNode {
	return gParent[node]
}
