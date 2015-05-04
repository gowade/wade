package vdom

import (
	"fmt"
	"sort"
	"strings"
)

type TreeModifier interface {
	SetAttr(DomNode, string, interface{})
	SetProp(DomNode, string, interface{})
	RemoveAttr(DomNode, string)
	Do(DomNode, Action)
}

type DomNode interface {
	Child(int) DomNode
}

func IsEvent(attr string) bool {
	return strings.HasPrefix(strings.ToLower(attr), "on")
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

	panic(fmt.Sprintf("Unhandled HTML attribute type %T", x))
	return false
}

func diffProps(a, b *Element, dNode DomNode, m TreeModifier) {
	for attr, va := range a.Attrs {
		if IsEvent(attr) {
			m.SetProp(dNode, strings.ToLower(attr), va)
			continue
		}

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

func getKey(node Node) string {
	if e, ok := node.(*Element); ok {
		return e.Key
	}

	return ""
}

type ActionType int

const (
	Deletion  ActionType = 0
	Insertion            = 1
	Move                 = 2
	Update               = 3
)

type Action struct {
	Type    ActionType
	Index   int
	From    int
	Element DomNode
	Content Node
}

type actionPriority []Action

func (a actionPriority) Len() int      { return len(a) }
func (a actionPriority) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a actionPriority) Less(i, j int) bool {
	if a[i].Type == a[j].Type {
		return a[i].Index < a[j].Index
	}

	return a[i].Type < a[j].Type
}

// PerformDiff calculates and performs operations on the DOM tree dNode
// to transform an old tree representation (b) to the new tree (a)
func PerformDiff(a, b *Element, dNode DomNode, m TreeModifier) {
	if b == nil || a.Tag != b.Tag {
		m.Do(dNode, Action{Type: Update, Content: a})
		return
	}

	a.SetRenderedDOMNode(b.DOMNode())
	diffProps(a, b, dNode, m)

	existing := make(map[string]Action)
	keyedDiff := false
	for i, bCh := range b.Children {
		key := getKey(bCh)
		if key != "" {
			keyedDiff = true
			existing[key] = Action{Type: Deletion, Index: i, Element: dNode.Child(i), Content: bCh}
		}
	}

	if keyedDiff { // Algorithm inspired by Mithril.js
		var unkeyed []Action
		for i, ac := range a.Children {
			aCh, ok := ac.(*Element)
			if !ok {
				unkeyed = append(unkeyed, Action{Type: Insertion, Index: i, Content: aCh})
				continue
			}

			key := aCh.Key
			if key != "" {
				if action, ok := existing[key]; !ok {
					existing[key] = Action{Type: Insertion, Index: i, Content: aCh.Render()}
				} else {
					aCh.oldElem = action.Content.(*Element)
					existing[key] = Action{
						Type:    Move,
						Index:   i,
						From:    action.Index,
						Element: dNode.Child(action.Index),
					}
				}
			} else {
				unkeyed = append(unkeyed, Action{Type: Insertion, Index: i, Content: aCh.Render()})
			}
		}

		actions := make([]Action, len(existing))
		i := 0
		for _, action := range existing {
			actions[i] = action
			i++
		}

		sort.Sort(actionPriority(actions))

		for _, action := range actions {
			m.Do(dNode, action)
			if action.Type == Move {
				PerformDiff(a.Children[action.Index].(*Element).Render(),
					b.Children[action.From].(*Element).Render(),
					dNode.Child(action.Index), m)
			}
		}

		for _, action := range unkeyed {
			m.Do(dNode, action)
		}

		return
	} // end keyed diff

	i := 0
	for c := 0; i < len(a.Children); i++ {
		if a.Children[i] == nil && b.Children[i] != nil {
			m.Do(dNode, Action{Type: Deletion, Index: i - c, Element: dNode.Child(i - c)})
			c++
		}
	}

	for i = 0; i < len(a.Children); i++ {
		aCh := a.Children[i]
		if aCh == nil {
			continue
		}

		ae, ok := aCh.(*Element)
		if i <= len(b.Children)-1 {
			bCh := b.Children[i]
			if bCh != nil && nodeCompat(aCh, bCh) {
				if aCh.IsElement() {
					be := bCh.(*Element)
					ae.oldElem = be
					PerformDiff(ae.Render(), be.Render(), dNode.Child(i), m)
				}
				continue
			}
		}

		ar := aCh
		if ok {
			ar = ae.Render()
		}

		if i > len(b.Children)-1 {
			m.Do(dNode, Action{Type: Insertion, Index: -1, Content: ar})
			continue
		} else {
			m.Do(dNode.Child(i), Action{Type: Update, Content: ar})
		}
	}

	for ii := i; i < len(b.Children); i++ {
		m.Do(dNode, Action{Type: Deletion, Index: i, Element: dNode.Child(ii)})
	}
}
