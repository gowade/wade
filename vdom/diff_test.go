package vdom_test

import (
	"testing"

	. "github.com/gowade/wade/vdom"

	"github.com/stretchr/testify/suite"
)

type DiffTestSuite struct {
	suite.Suite
}

type GoDomNode struct {
	Node
}

func (n GoDomNode) Child(idx int) DomNode {
	return GoDomNode{n.Node.(*Element).Children[idx]}
}

type attrChange struct {
	remove  bool
	attr    string
	value   interface{}
	domNode DomNode
}

type change struct {
	action Action
	dNode  GoDomNode
}

func (c change) affectedNode() GoDomNode {
	return c.dNode.Child(c.action.Index).(GoDomNode)
}

type modifier struct {
	changes     []change
	attrChanges []attrChange
}

func (m *modifier) recordAC(c attrChange) {
	m.attrChanges = append(m.attrChanges, c)
}

func (m *modifier) Do(d DomNode, action Action) {
	change := change{action: action}
	if d != nil {
		change.dNode = d.(GoDomNode)
	}
	m.changes = append(m.changes, change)
}

func (m *modifier) SetAttr(d DomNode, attr string, v interface{}) {
	if b, ok := v.(bool); ok && b == false {
		m.RemoveAttr(d, attr)
		return
	}

	m.recordAC(attrChange{false, attr, v, d})
}

func (m *modifier) RemoveAttr(d DomNode, attr string) {
	m.recordAC(attrChange{true, attr, nil, d})
}

func newModifier() *modifier {
	return &modifier{make([]change, 0), make([]attrChange, 0)}
}
func (s *DiffTestSuite) TestDiff() {
	m1 := newModifier()
	a := NewElement("div", nil, nil)
	PerformDiff(a, nil, GoDomNode{NewElement("div", nil, nil)}, m1)
	s.Len(m1.changes, 1)
	s.Equal(m1.changes[0].action.Type, Update)
	s.Equal(m1.changes[0].dNode.NodeData(), "div")

	b := NewElement("div", Attributes{"title": "d"}, []Node{
		NewElement("span", nil, []Node{NewTextNode("C")}),
		NewElement("ul", Attributes{"disabled": true}, []Node{
			NewElement("li", nil, []Node{NewTextNode("A")}),
		}),
	})
	d := GoDomNode{b}
	a = NewElement("div", nil, []Node{
		NewElement("span", nil, []Node{}),
		NewElement("ul", Attributes{"disabled": false, "value": "0"}, []Node{
			NewElement("notli", Attributes{"id": "11"}, []Node{NewTextNode("A")}),
			NewElement("li", nil, []Node{NewTextNode("B")}),
		})})

	m1 = newModifier()
	PerformDiff(a, b, d, m1)
	s.Equal(m1.changes[0].action, Action{Type: Deletion, Index: 0})
	s.Equal(m1.changes[0].affectedNode().NodeData(), "C")

	s.Equal(m1.changes[1].action.Type, Update)
	s.Equal(m1.changes[1].action.Content.NodeData(), "notli")
	s.Equal(m1.changes[1].dNode.NodeData(), "li")

	s.Equal(m1.changes[2].action.Type, Insertion)
	s.Equal(m1.changes[2].action.Content.(*Element).Children[0].NodeData(), "B")

	s.Len(m1.changes, 3)

	// Test attribute diffing
	s.Equal(m1.attrChanges[0].remove, true)
	s.Equal(m1.attrChanges[0].attr, "title")
	s.Equal(m1.attrChanges[0].domNode.(GoDomNode).NodeData(), "div")

	s.Equal(m1.attrChanges[1].remove, true)
	s.Equal(m1.attrChanges[1].attr, "disabled")
	s.Equal(m1.attrChanges[1].value, nil)
	s.Equal(m1.attrChanges[1].domNode.(GoDomNode).NodeData(), "ul")

	s.Equal(m1.attrChanges[2].remove, false)
	s.Equal(m1.attrChanges[2].attr, "value")
	s.Equal(m1.attrChanges[2].value, "0")
	s.Equal(m1.attrChanges[2].domNode.(GoDomNode).NodeData(), "ul")

	s.Len(m1.attrChanges, 3)
}

func (s *DiffTestSuite) TestKeyedDiff() {
	m1 := newModifier()
	b := NewElement("div", nil, []Node{
		NewElement("ul", nil, []Node{
			NewElement("li", Attributes{"key": 1}, nil),
			NewElement("li", Attributes{"key": 2}, nil),
			NewElement("li", Attributes{"key": 3}, nil),
			NewElement("li", Attributes{"key": 4}, nil),
		}),
	})
	d := GoDomNode{b}
	a := NewElement("div", nil, []Node{
		NewElement("ul", nil, []Node{
			NewElement("li", Attributes{"key": 0}, nil),
			NewElement("li", Attributes{"key": 4}, nil),
			NewElement("li", nil, nil),
			NewElement("li", Attributes{"key": 2}, nil),
			NewElement("li", Attributes{"key": 5}, nil),
		}),
	})

	PerformDiff(a, b, d, m1)

	s.Equal(m1.changes[0].action, Action{Type: Deletion, Index: 0})
	s.Equal(m1.changes[1].action, Action{Type: Deletion, Index: 2})

	s.Equal(m1.changes[2].action.Type, Insertion)
	s.Equal(m1.changes[2].action.Index, 0)
	s.Equal(m1.changes[2].action.Content.(*Element).Attrs["key"], 0)

	s.Equal(m1.changes[3].action.Type, Insertion)
	s.Equal(m1.changes[3].action.Index, 4)
	s.Equal(m1.changes[3].action.Content.(*Element).Attrs["key"], 5)

	s.Equal(m1.changes[4].action.Type, Move)
	s.Equal(m1.changes[4].action.Index, 1)
	s.Equal(m1.changes[4].action.From, 3)

	s.Equal(m1.changes[5].action.Type, Move)
	s.Equal(m1.changes[5].action.Index, 3)
	s.Equal(m1.changes[5].action.From, 1)

	// unkeyed
	s.Equal(m1.changes[6].action.Type, Insertion)
	s.Equal(m1.changes[6].action.Index, 2)
}

func TestDiff(t *testing.T) {
	suite.Run(t, new(DiffTestSuite))
}
