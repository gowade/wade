package vdom_test

import (
	"testing"

	. "github.com/gowade/wade/vdom"

	"github.com/gowade/wade/utils/htmlutils"
	"golang.org/x/net/html"

	"github.com/stretchr/testify/suite"
)

type DiffTestSuite struct {
	suite.Suite
}

type changeType string

const (
	Insert changeType = "INSERT"
	Update            = "UPDATE"
	Delete            = "DELETE"
)

func NewNode(src string) GoDomNode {
	return GoDomNode{htmlutils.FragmentFromString(src)}
}

type GoDomNode struct {
	*html.Node
}

func (n GoDomNode) Child(idx int) DomNode {
	for i, c := 0, n.FirstChild; ; c, i = c.NextSibling, i+1 {
		if i == idx {
			return GoDomNode{c}
		}
	}

	return nil
}

type change struct {
	typ     changeType
	node    Node
	domNode DomNode
}

type attrChange struct {
	remove  bool
	attr    string
	value   interface{}
	domNode DomNode
}

type modifier struct {
	changes     []change
	attrChanges []attrChange
}

func (m *modifier) record(c change) {
	m.changes = append(m.changes, c)
}

func (m *modifier) recordAC(c attrChange) {
	m.attrChanges = append(m.attrChanges, c)
}

func (m *modifier) Render(n Node, d DomNode) {
	m.record(change{Update, n, d})
}

func (m *modifier) Insert(n Node, d DomNode) {
	m.record(change{Insert, n, d})
}

func (m *modifier) Delete(d DomNode) {
	m.record(change{Delete, nil, d})
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
	PerformDiff(a, nil, nil, m1)
	s.Len(m1.changes, 1)
	s.Equal(change{Update, a, nil}, m1.changes[0])

	d, b := NewNode("<div><span>C</span><ul><li>A</li></ul></div>"),
		NewElement("div", Attributes{"title": "d"}, []Node{
			NewElement("span", nil, []Node{NewTextNode("C")}),
			NewElement("ul", Attributes{"disabled": true}, []Node{
				NewElement("li", nil, []Node{NewTextNode("A")}),
			}),
		})
	a = NewElement("div", nil, []Node{
		NewElement("span", nil, []Node{}),
		NewElement("ul", Attributes{"disabled": false, "value": "0"}, []Node{
			NewElement("notli", Attributes{"id": "11"}, []Node{NewTextNode("A")}),
			NewElement("li", nil, []Node{NewTextNode("B")}),
		})})

	m1 = newModifier()
	PerformDiff(a, b, d, m1)
	s.Equal(m1.changes[0].typ, Delete)
	s.Equal(m1.changes[0].domNode.(GoDomNode).Data, "C")

	s.Equal(m1.changes[1].typ, Update)
	s.Equal(m1.changes[1].node.(*Element).Tag, "notli")
	s.Equal(m1.changes[1].domNode.(GoDomNode).Data, "li")

	s.Equal(m1.changes[2].typ, Insert)
	s.Equal(m1.changes[2].node.(*Element).Children[0].(*TextNode).Data, "B")
	s.Equal(m1.changes[2].domNode.(GoDomNode).Data, "ul")

	s.Len(m1.changes, 3)

	// Test attribute changes
	s.Equal(m1.attrChanges[0].remove, true)
	s.Equal(m1.attrChanges[0].attr, "title")
	s.Equal(m1.attrChanges[0].domNode.(GoDomNode).Data, "div")

	s.Equal(m1.attrChanges[1].remove, true)
	s.Equal(m1.attrChanges[1].attr, "disabled")
	s.Equal(m1.attrChanges[1].value, nil)
	s.Equal(m1.attrChanges[1].domNode.(GoDomNode).Data, "ul")

	s.Equal(m1.attrChanges[2].remove, false)
	s.Equal(m1.attrChanges[2].attr, "value")
	s.Equal(m1.attrChanges[2].value, "0")
	s.Equal(m1.attrChanges[2].domNode.(GoDomNode).Data, "ul")

	s.Len(m1.attrChanges, 3)
}

func TestDiff(t *testing.T) {
	suite.Run(t, new(DiffTestSuite))
}
