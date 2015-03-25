package vdom

import (
	"fmt"
)

type TreeModifier interface {
	Render(Node, DomNode)
	SetAttr(DomNode, string, interface{})
	RemoveAttr(DomNode, string)
	Insert(Node, DomNode)
	Delete(DomNode)
}

type DomNode interface {
	Child(int) DomNode
}

func nodeCompat(a, b Node) bool {
	aie, bie := a.IsElement(), b.IsElement()
	if aie != bie {
		return false
	}

	if aie {
		return a.(*Element).Tag == b.(*Element).Tag
	}

	return a.(*TextNode).Data == b.(*TextNode).Data
}

func equal(va, vb interface{}) bool {
	switch va.(type) {
	case string:
		return va == vb.(string)
	case bool:
		return va == vb.(bool)
	case int:
		return va == vb.(int)
	case int64:
		return va == vb.(int64)
	case int32:
		return va == vb.(int32)
	case float32:
		return va == vb.(float32)
	case float64:
		return va == vb.(float64)
	}

	panic(fmt.Sprintf("Unsupported attribute type %T", va))
	return false
}

func diffProps(a, b *Element, dNode DomNode, m TreeModifier) {
	for attr, va := range a.Attrs {
		if vb, ok := b.Attrs[attr]; !ok || !equal(va, vb) {
			m.SetAttr(dNode, attr, va)
		}
	}

	for attr, _ := range b.Attrs {
		if _, ok := a.Attrs[attr]; !ok {
			m.RemoveAttr(dNode, attr)
		}
	}
}

// PerformDiff calculates and performs operations on the DOM tree dNode
// to transform an old tree representation (b) to the new tree (a)
func PerformDiff(a, b *Element, dNode DomNode, m TreeModifier) {
	if b == nil || a.Tag != b.Tag {
		m.Render(a, dNode)
		return
	}

	diffProps(a, b, dNode, m)

	i := 0
	for ; i < len(a.Children); i++ {
		aCh := a.Children[i]
		if i > len(b.Children)-1 {
			m.Insert(aCh, dNode)
			continue
		}

		bCh := b.Children[i]
		if nodeCompat(aCh, bCh) && aCh.IsElement() {
			PerformDiff(aCh.(*Element), bCh.(*Element), dNode.Child(i), m)
		} else {
			m.Render(aCh, dNode.Child(i))
		}
	}

	for ; i < len(b.Children); i++ {
		m.Delete(dNode.Child(i))
	}
}
