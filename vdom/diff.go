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
	Element DomNode
	From    Node
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
func PerformDiff(an, bn Node, dNode DomNode, m TreeModifier) {
	if bn == nil || an.IsElement() != bn.IsElement() || an.NodeData() != bn.NodeData() {
		m.Do(dNode, Action{Type: Update, Content: an})
		return
	}

	if !an.IsElement() {
		return
	}

	a, b := an.(*Element), bn.(*Element)

	a.SetRenderedDOMNode(b.DOMNode())
	diffProps(a, b, dNode, m)

	existing := make(map[string]Action)
	keyedDiff := false
	var unkeyed []Action

	for _, bCh := range b.Children {
		if bCh != nil && getKey(bCh) != "" {
			keyedDiff = true
		}
	}

	uki := 0
	if keyedDiff { // Algorithm inspired by Mithril.js

		offset := 0
		for ii, bCh := range b.Children {
			i := ii - offset
			if bCh == nil {
				offset++
				continue
			}

			if key := getKey(bCh); key != "" {
				existing[key] = Action{
					Type:    Deletion,
					Index:   i,
					Element: dNode.Child(i),
					Content: bCh.Render(),
				}
			} else {
				unkeyed = append(unkeyed, Action{
					Type:    Deletion,
					Index:   i,
					Element: dNode.Child(i),
					Content: bCh.Render(),
				})
			}
		}

		offset = 0
		for ii, ac := range a.Children {
			i := ii - offset
			if ac == nil {
				offset++
				continue
			}

			key := getKey(ac)
			if key != "" {
				aCh := ac.(*Element)
				if action, ok := existing[key]; !ok {
					existing[key] = Action{Type: Insertion, Index: i, Content: aCh.Render()}
				} else {
					aCh.oldElem = action.Content.(*Element)
					existing[key] = Action{
						Type:    Move,
						Index:   i,
						From:    action.Content,
						Element: action.Element,
						Content: aCh.Render(),
					}
				}
			} else {
				if len(unkeyed) > uki {
					action := &unkeyed[uki]
					if ac.IsElement() && action.Content.IsElement() {
						ac.(*Element).oldElem = action.Content.(*Element)
					}

					*action = Action{
						Type:    Move,
						Index:   i,
						From:    action.Content,
						Element: action.Element,
						Content: ac.Render(),
					}
				} else {
					unkeyed = append(unkeyed,
						Action{Type: Insertion, Index: i, Content: ac.Render()})
				}
				uki++
			}
		}

		actions := make([]Action, len(existing)+len(unkeyed))
		i := 0
		for k := range existing {
			actions[i] = existing[k]
			i++
		}

		for k := range unkeyed {
			actions[i] = unkeyed[k]
			i++
		}

		sort.Sort(actionPriority(actions))

		for _, action := range actions {
			m.Do(dNode, action)
			if action.Type == Move {
				PerformDiff(action.Content,
					action.From,
					action.Element, m)
			}
		}

		return
	} // end keyed diff

	bd := make([]DomNode, len(b.Children))
	bp := make([]int, len(b.Children))
	c := 0
	for i, bCh := range b.Children {
		if bCh != nil {
			bd[i] = dNode.Child(c)
			c++
		} else {
			bp[i] = c
		}
	}

	for i := len(bd) - 1; i >= 0; i-- {
		var aCh Node
		if i < len(a.Children) {
			aCh = a.Children[i]
		}

		if bd[i] != nil {
			if aCh != nil {
				bCh := b.Children[i]
				if aCh.IsElement() && bCh.IsElement() {
					bCh = bCh.Render()
					aCh.(*Element).oldElem = bCh.(*Element)
				}

				PerformDiff(aCh.Render(), bCh, bd[i], m)
			} else {
				m.Do(dNode, Action{Type: Deletion, Element: bd[i]})
			}
		} else if aCh != nil {
			m.Do(dNode, Action{Type: Insertion, Index: bp[i], Content: aCh})
		}
	}

	for i := len(b.Children); i < len(a.Children); i++ {
		if a.Children[i] != nil {
			m.Do(dNode, Action{
				Type:    Insertion,
				Index:   -1,
				Content: a.Children[i],
			})
		}
	}
}
