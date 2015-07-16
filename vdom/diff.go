package vdom

import (
	"fmt"
	"sort"
	"strings"
)

var (
	domReady   bool
	domUpdated int
)

func InternalRenderLock() {
	domReady = false
}

func InternalRenderUnlock() {
	domReady = true
}

func InternalRenderLocked() bool {
	return !domReady
}

func SetUpdated(depth int) {
	if domUpdated < depth {
		domUpdated = depth
	}
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

func diffProps(a, b *Element, dNode DOMNode) (updated bool) {
	for attr, va := range a.Attrs {
		if IsEvent(attr) {
			continue
		}

		if vb, ok := b.Attrs[attr]; !ok || !equals(va, vb) {
			updated = true
			dNode.SetAttr(attr, va)
		}
	}

	for attr, _ := range b.Attrs {
		if _, ok := a.Attrs[attr]; !ok {
			updated = true
			dNode.RemoveAttr(attr)
		}
	}

	return
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
)

type Action struct {
	Type    ActionType
	Index   int
	Element DOMNode
	From    Node
	Content Node
}

type actionPriority []*Action

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
// The root node will not be replaced (even if its tag name isn't compatible)
func PerformDiff(a, b *Element, dNode DOMNode) {
	SetUpdated(-1)

	if b == nil {
		dNode.Clear()
	}

	if dNode == nil {
		panic("target DOM node is nil.")
	}

	var an, bn Node
	if a != nil {
		an = a
	}

	if b != nil {
		bn = b
	}

	performDiff(an, bn, dNode, true, 0)
}

func checkDomUpdated(el *Element, depth int) {
	if domUpdated >= depth && el.comref != nil {
		el.comref.OnUpdated()
	}
}

func do(action *Action, dNode DOMNode, depth int) {
	switch action.Type {
	case Deletion, Insertion:
		SetUpdated(depth)
	}

	dNode.Do(action)
}

func performDiff(an, bn Node, dNode DOMNode, root bool, depth int) {
	if bn == nil || an.IsElement() != bn.IsElement() || an.NodeData() != bn.NodeData() ||
		!dNode.Compat(bn) {

		SetUpdated(depth - 1)
		dNode.Render(an, root)
		return
	}

	if !an.IsElement() {
		return
	}

	ar, br := an.Render(), bn.Render()
	if ar == nil || br == nil {
		return
	}

	a, b := ar.(*Element), br.(*Element)
	a.SetRenderedDOMNode(b.DOMNode())
	updated := diffProps(a, b, dNode)
	if updated {
		SetUpdated(depth)
	}

	existing := make(map[string]*Action)
	keyedDiff := false
	var unkeyed []*Action

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
				existing[key] = &Action{
					Type:    Deletion,
					Index:   i,
					Element: dNode.Child(i),
					Content: bCh,
				}
			} else {
				unkeyed = append(unkeyed, &Action{
					Type:    Deletion,
					Index:   i,
					Element: dNode.Child(i),
					Content: bCh,
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
					existing[key] = &Action{Type: Insertion, Index: i, Content: aCh}
				} else {
					aCh.oldElem = action.Content.(*Element)
					existing[key] = &Action{
						Type:    Move,
						Index:   i,
						From:    action.Content,
						Element: action.Element,
						Content: aCh,
					}
				}
			} else {
				if len(unkeyed) > uki {
					action := unkeyed[uki]
					if ac.IsElement() && action.Content.IsElement() {
						ac.(*Element).oldElem = action.Content.(*Element)
					}

					*action = Action{
						Type:    Move,
						Index:   i,
						From:    action.Content,
						Element: action.Element,
						Content: ac,
					}
				} else {
					unkeyed = append(unkeyed,
						&Action{Type: Insertion, Index: i, Content: ac})
				}
				uki++
			}
		}

		actions := make([]*Action, len(existing)+len(unkeyed))
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
			do(action, dNode, depth)
			if action.Type == Move {
				performDiff(action.Content,
					action.From,
					action.Element, false, depth+1)
			}
		}

		checkDomUpdated(a, depth)
		return
	} // end keyed diff

	bd := make([]DOMNode, len(b.Children))
	bp := make([]int, len(b.Children))

	var bCh, aCh Node
	c := 0
	for i, bCh := range b.Children {
		if bCh != nil {
			bd[i] = dNode.Child(c)
			c++
		} else {
			bp[i] = c
		}
	}

	for i := len(b.Children) - 1; i >= 0; i-- {
		bCh = b.Children[i]

		if i < len(a.Children) {
			aCh = a.Children[i]
		} else {
			aCh = nil
		}

		if bCh != nil && aCh == nil {
			do(&Action{Type: Deletion, Element: bd[i]}, dNode, depth)
			if bCh.IsElement() {
				bCh.(*Element).Unmount()
			}
		} else if bCh == nil && aCh != nil {
			do(&Action{Type: Insertion, Index: bp[i], Content: aCh}, dNode, depth)
		}
	}

	for i, aCh := range a.Children {
		if i >= len(b.Children) {
			break
		}

		bCh = b.Children[i]
		if aCh != nil && bCh != nil {
			if aCh.IsElement() && bCh.IsElement() {
				aCh.(*Element).oldElem = bCh.(*Element)
			}

			performDiff(aCh, bCh, bd[i], false, depth+1)
		}
	}

	for i := len(b.Children); i < len(a.Children); i++ {
		if a.Children[i] != nil {
			do(&Action{
				Type:    Insertion,
				Index:   -1,
				Content: a.Children[i],
			}, dNode, depth)
		}
	}

	checkDomUpdated(a, depth)
}
