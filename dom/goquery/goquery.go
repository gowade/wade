package goquery

import (
	"bytes"
	"strings"

	"code.google.com/p/go.net/html"
	"code.google.com/p/go.net/html/atom"
	"github.com/PuerkitoBio/goquery"

	"github.com/phaikawl/wade/dom"
)

var (
	gDom = Dom{}
)

type (
	Dom struct{}

	Selection struct {
		*goquery.Selection
		Dom
	}

	operateFunc func(dst, src *html.Node)
)

func GetDom() dom.Dom {
	return gDom
}

func newSelection(gq *goquery.Selection) dom.Selection {
	return Selection{gq, gDom}
}

func (d Dom) NewDocument(source string) dom.Selection {
	node, err := html.Parse(bytes.NewBufferString(source))
	if err != nil {
		panic(err)
	}

	s := goquery.NewDocumentFromNode(node)

	return newSelection(s.Selection.Children().First())
}

func (d Dom) NewRootFragment(source string) dom.Selection {
	return d.NewFragment(source)
}

func parseHTML(source string) []*html.Node {
	nodes, err := html.ParseFragment(bytes.NewBufferString(source), &html.Node{
		Type:     html.ElementNode,
		Data:     "body",
		DataAtom: atom.Body,
	})

	if err != nil {
		panic(err)
	}

	if len(nodes) == 0 {
		panic("Source string is empty or cannot be parsed.")
	}

	for _, node := range nodes {
		node.Parent = nil
		node.PrevSibling = nil
		node.NextSibling = nil
	}

	return nodes
}

func newFragment(source string) dom.Selection {
	nodes := parseHTML(source)
	if strings.TrimSpace(nodes[0].Data) == "" {
		panic("Parsing failed. Note that parsing html, head or body element *as fragment* will kill the parser. This may be the case.")
	}

	sel := goquery.NewDocumentFromNode(nodes[0]).Selection
	sel.AddNodes(nodes[1:]...)

	return newSelection(sel)
}

func (d Dom) NewFragment(source string) dom.Selection {
	return newFragment(source)
}

func (s Selection) firstNode() *html.Node {
	if len(s.Nodes) == 0 {
		panic("The selection has 0 nodes.")
	}

	node := s.Nodes[0]
	if node.Type == html.DocumentNode {
		return node.FirstChild
	}

	return node
}

func (s Selection) First() dom.Selection {
	return newSelection(s.Selection.First())
}

func (s Selection) Children() dom.Selection {
	return newSelection(s.Selection.Children())
}

func (s Selection) Contents() dom.Selection {
	return newSelection(s.Selection.Contents())
}

func (s Selection) IsElement() bool {
	return s.firstNode().Type == html.ElementNode
}

func (s Selection) TagName() (string, error) {
	if len(s.Nodes) == 0 {
		return "", dom.ErrorNoElementSelected
	}

	node := s.firstNode()
	if !s.IsElement() {
		return "", dom.ErrorCantGetTagName
	}

	return strings.ToLower(node.Data), nil
}

func (s Selection) Find(selector string) dom.Selection {
	return newSelection(s.Selection.Find(selector))
}

func (s Selection) Html() string {
	contents, _ := s.Selection.Html()
	return contents
}

func (s Selection) Elements() []dom.Selection {
	list := make([]dom.Selection, s.Length())
	s.Selection.Each(func(i int, elem *goquery.Selection) {
		list[i] = newSelection(elem)
	})

	return list
}

func (s Selection) Remove() {
	for _, node := range s.Nodes {
		if node.Parent != nil {
			node.Parent.RemoveChild(node)
		}
	}
}

func (s Selection) Clone() dom.Selection {
	var sel *goquery.Selection
	for i, node := range s.Nodes {
		buf := bytes.NewBufferString("")
		html.Render(buf, node)
		nn := parseHTML(buf.String())[0]
		if i == 0 {
			sel = goquery.NewDocumentFromNode(nn).Selection
		} else {
			sel.AddNodes(nn)
		}
	}

	return newSelection(sel)
}

func (s Selection) operate(sel dom.Selection, opFunc operateFunc) {
	sel.Remove()
	for i, node := range s.Nodes {
		var cont dom.Selection
		if i == len(s.Nodes)-1 {
			cont = sel
		} else {
			cont = sel.Clone()
		}

		for _, cnode := range cont.(Selection).Nodes {
			opFunc(node, cnode)
		}
	}
}

func (s Selection) Append(sel dom.Selection) {
	s.operate(sel, func(dst, src *html.Node) {
		dst.AppendChild(src)
	})
}

func (s Selection) ReplaceWith(sel dom.Selection) {
	s.operate(sel, func(dst, src *html.Node) {
		if dst.Parent == nil {
			panic("Element has no parent, cannot perform replace.")
		}
		dst.Parent.InsertBefore(src, dst)
	})
	s.Remove()
}

func (s Selection) OuterHtml() string {
	output := bytes.NewBufferString("")
	for _, node := range s.Nodes {
		html.Render(output, node)
	}

	return output.String()
}

func (s Selection) Parents() dom.Selection {
	return newSelection(s.Selection.Parents())
}

func (s Selection) Parent() dom.Selection {
	return newSelection(s.Selection.Parent())
}

func (s Selection) Unwrap() {
	s.ReplaceWith(s.Contents())
}
