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

// this function is copied from ssa/interp
// equals returns true iff x and y are equal according to Go's
// linguistic equivalence relation for type t.
// In a well-typed program, the dynamic types of x and y are
// guaranteed equal.
func equals(x, y interface{}) bool {
	switch x := x.(type) {
	case bool:
		return x == y.(bool)
	case int:
		return x == y.(int)
	case int8:
		return x == y.(int8)
	case int16:
		return x == y.(int16)
	case int32:
		return x == y.(int32)
	case int64:
		return x == y.(int64)
	case uint:
		return x == y.(uint)
	case uint8:
		return x == y.(uint8)
	case uint16:
		return x == y.(uint16)
	case uint32:
		return x == y.(uint32)
	case uint64:
		return x == y.(uint64)
	case uintptr:
		return x == y.(uintptr)
	case float32:
		return x == y.(float32)
	case float64:
		return x == y.(float64)
	case complex64:
		return x == y.(complex64)
	case complex128:
		return x == y.(complex128)
	case string:
		return x == y.(string)
	}

	panic(fmt.Sprintf("comparing uncomparable type %T", x))
}

func diffProps(a, b *Element, dNode DomNode, m TreeModifier) {
	for attr, va := range a.Attrs {
		if vb, ok := b.Attrs[attr]; !ok || !equals(va, vb) {
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
