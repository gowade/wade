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

func NewNode(src string) GoNode {
	return GoNode{htmlutils.FragmentFromString(src)}
}

type GoNode struct {
	*html.Node
}

func (n GoNode) Children() []DomNode {
	chs := make([]DomNode, 0)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		chs = append(chs, GoNode{c})
	}

	return chs
}

func (n GoNode) AppendChild(c DomNode) {
	n.Node.AppendChild(c.(GoNode).Node)
}

func (n GoNode) Remove() {
	n.Node.Parent.RemoveChild(n.Node)
}

type change struct {
	typ     changeType
	node    Node
	domNode DomNode
}

type modifier struct {
	changes []change
}

func (m *modifier) record(c change) {
	m.changes = append(m.changes, c)
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

func newModifier() *modifier {
	return &modifier{make([]change, 0)}
}
func (s *DiffTestSuite) TestDiff() {
	m1 := newModifier()
	a := NewElement("div", nil, nil)
	PerformDiff(a, nil, nil, m1)
	s.Len(m1.changes, 1)
	s.Equal(change{Update, a, nil}, m1.changes[0])

	d, b := NewNode("<div><span>C</span><ul><li>A</li></ul></div>"),
		NewElement("div", nil, []Node{
			NewElement("span", nil, []Node{NewTextNode("C")}),
			NewElement("ul", nil, []Node{
				NewElement("li", nil, []Node{NewTextNode("A")}),
			}),
		})
	a = NewElement("div", nil, []Node{
		NewElement("span", nil, []Node{}),
		NewElement("ul", nil, []Node{
			NewElement("notli", nil, []Node{NewTextNode("A")}),
			NewElement("li", nil, []Node{NewTextNode("B")}),
		})})

	m1 = newModifier()
	PerformDiff(a, b, d, m1)
	s.Equal(m1.changes[0].typ, Delete)
	s.Equal(m1.changes[0].domNode.(GoNode).Data, "C")

	s.Equal(m1.changes[1].typ, Update)
	s.Equal(m1.changes[1].node.(*Element).Tag, "notli")
	s.Equal(m1.changes[1].domNode.(GoNode).Data, "li")

	s.Equal(m1.changes[2].typ, Insert)
	s.Equal(m1.changes[2].node.(*Element).Children[0].(*TextNode).Data, "B")
	s.Equal(m1.changes[2].domNode.(GoNode).Data, "ul")

	s.Len(m1.changes, 3)
}

func TestDiff(t *testing.T) {
	suite.Run(t, new(DiffTestSuite))
}
