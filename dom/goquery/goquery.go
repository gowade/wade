package goquery

import (
	"bytes"
	"encoding/json"
	"strings"

	"code.google.com/p/go.net/html"
	"code.google.com/p/go.net/html/atom"
	"github.com/PuerkitoBio/goquery"

	"github.com/phaikawl/wade/dom"
)

const (
	PropPrefix = "wade-rsvd-prop-"
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

func GetDom() Dom {
	return gDom
}

func newSelection(gq *goquery.Selection) dom.Selection {
	return Selection{gq, gDom}
}

func (d Dom) NewDocument(source string) dom.Selection {
	node, err := html.Parse(bytes.NewBufferString(strings.TrimSpace(source)))
	if err != nil {
		panic(err)
	}

	s := goquery.NewDocumentFromNode(node)

	return newSelection(s.Selection.Children().First())
}

func (d Dom) NewEmptySelection() dom.Selection {
	return d.NewFragment("")
}

func (d Dom) NewRootFragment() dom.Selection {
	return d.NewFragment("<wroot></wroot>")
}

func (d Dom) NewTextNode(content string) dom.Selection {
	node := d.NewFragment("a")
	node.SetText(content)
	return node
}

func parseHTML(source string) []*html.Node {
	nodes, err := html.ParseFragment(bytes.NewBufferString(strings.TrimSpace(source)), &html.Node{
		Type:     html.ElementNode,
		Data:     "body",
		DataAtom: atom.Body,
	})

	if err != nil {
		panic(err)
	}

	for _, node := range nodes {
		node.Parent = nil
		node.PrevSibling = nil
		node.NextSibling = nil
	}

	return nodes
}

func selFromNodes(nodes []*html.Node) dom.Selection {
	sel := goquery.NewDocumentFromNode(nodes[0]).Selection

	return newSelection(sel.AddNodes(nodes[1:]...))
}

func newFragment(source string) dom.Selection {
	nodes := parseHTML(source)
	if len(nodes) == 0 {
		empty := goquery.NewDocumentFromNode(nil)
		empty.Nodes = []*html.Node{}
		return newSelection(empty.Selection)
	}

	if nodes[0].Type == html.ErrorNode {
		panic("Parsing failed. Note that parsing html, head or body element *as fragment* will kill the parser. This may be the case.")
	}

	return selFromNodes(nodes)
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
	nnodes := make([]*html.Node, 0)
	i := 0
	for _, node := range s.Nodes {
		var pnodes []*html.Node
		if node.Type == html.TextNode {
			pnodes = s.NewTextNode(s.Text()).(Selection).Nodes
		} else {
			buf := bytes.NewBufferString("")
			html.Render(buf, node)
			pnodes = parseHTML(buf.String())
		}

		if len(pnodes) == 1 {
			if i == 0 {
				sel = goquery.NewDocumentFromNode(pnodes[0]).Selection
			} else {
				nnodes = append(nnodes, pnodes[0])
			}
			i++
		}
	}

	return newSelection(sel.AddNodes(nnodes...))
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

func (s Selection) Prepend(sel dom.Selection) {
	if s.Contents().Length() > 0 {
		s.Contents().First().Before(sel)
	} else {
		s.Append(sel)
	}
}

func (s Selection) After(sel dom.Selection) {
	s.operate(sel, func(dst, src *html.Node) {
		if dst.NextSibling != nil {
			dst.Parent.InsertBefore(src, dst.NextSibling)
		} else {
			dst.Parent.AppendChild(src)
		}
	})
}

func (s Selection) Before(sel dom.Selection) {
	s.operate(sel, func(dst, src *html.Node) {
		dst.Parent.InsertBefore(src, dst)
	})
}

func (s Selection) ReplaceWith(sel dom.Selection) {
	s.Before(sel)
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
	for _, elem := range s.Elements() {
		elem.ReplaceWith(elem.Contents())
	}
}

func (s Selection) SetHtml(content string) {
	s.Contents().Remove()
	s.Append(s.NewFragment(content))
}

func (s Selection) Val() string {
	attr, _ := s.First().Attr("value")
	return attr
}

func (s Selection) Attr(attr string) (string, bool) {
	return s.Selection.Attr(strings.ToLower(attr))
}

func (s Selection) SetAttr(name, value string) {
	name = strings.ToLower(name)
	for _, node := range s.Nodes {
		ok := false
		for i, attr := range node.Attr {
			if attr.Key == name {
				node.Attr[i] = html.Attribute{
					Key: name,
					Val: value,
				}

				ok = true
				break
			}
		}

		if !ok {
			node.Attr = append(node.Attr, html.Attribute{
				Key: name,
				Val: value,
			})
		}
	}
}

func (s Selection) SetVal(value string) {
	s.SetAttr("value", value)
}

func (s Selection) RemoveAttr(name string) {
	for _, node := range s.Nodes {
		for i, attr := range node.Attr {
			if attr.Key == name {
				node.Attr = append(node.Attr[:i], node.Attr[i+1:]...)
				break
			}
		}
	}
}

func (s Selection) Next() dom.Selection {
	nsnodes := make([]*html.Node, len(s.Nodes))
	for i, node := range s.Nodes {
		nsnodes[i] = node.NextSibling
	}

	return selFromNodes(nsnodes)
}

func (s Selection) Exists() bool {
	return s.Selection.Is("html") || s.Selection.Is("wroot") ||
		s.Selection.ParentsFiltered("wroot").Length() > 0 || s.Selection.ParentsFiltered("html").Length() > 0
}

func (s Selection) On(eventname string, handler dom.EventHandler) {
	//stub
}

func (s Selection) Attrs() []dom.Attr {
	aa := s.Selection.First().Nodes[0].Attr
	attrs := make([]dom.Attr, len(aa))
	for i, attr := range aa {
		attrs[i].Name = attr.Key
		attrs[i].Value = attr.Val
	}

	return attrs
}

func (s Selection) Prev() dom.Selection {
	nsnodes := make([]*html.Node, len(s.Nodes))
	for i, node := range s.Nodes {
		nsnodes[i] = node.PrevSibling
	}

	return selFromNodes(nsnodes)
}

func (s Selection) Listen(event string, selector string, handler dom.EventHandler) {
	//stub
}

func (s Selection) Hide() {
	//stub
}

func (s Selection) Show() {
	//stub
}

func (s Selection) AddClass(class string) {
	for _, elem := range s.Elements() {
		if !elem.HasClass(class) {
			cl, _ := elem.Attr("class")
			elem.SetAttr("class", cl+" "+class)
		}
	}
}

func (s Selection) RemoveClass(class string) {
	for _, elem := range s.Elements() {
		if elem.HasClass(class) {
			elClass, _ := elem.Attr("class")
			newCl := strings.Replace(" "+elClass+" ", " "+class+" ", " ", -1)
			elem.SetAttr("class", newCl)
		}
	}
}

func (s Selection) Filter(selector string) dom.Selection {
	return newSelection(s.Selection.Filter(selector))
}

func (s Selection) IsTextNode() bool {
	return s.Selection.Nodes[0].Type == html.TextNode
}

func (s Selection) SetText(text string) {
	for _, node := range s.Nodes {
		node.Data = text
	}
}

func (s Selection) Add(elem dom.Selection) dom.Selection {
	return newSelection(s.Selection.AddSelection(elem.(Selection).Selection))
}

func (s Selection) Prop(prop string, valueRecv interface{}) (ok bool) {
	str, ok := s.Attr(PropPrefix + prop)
	if ok {
		err := json.Unmarshal([]byte(str), valueRecv)
		if err != nil {
			ok = false
		}
	}

	return
}

func (s Selection) SetProp(prop string, value interface{}) {
	bytes, _ := json.Marshal(value)
	s.SetAttr(PropPrefix+prop, string(bytes[:]))
	if tf, ok := value.(bool); ok {
		if v, e := s.Attr(prop); !e || v == "reserved-prop-set" {
			if tf == true {
				s.SetAttr(prop, "reserved-prop-set")
			} else {
				s.RemoveAttr(prop)
			}
		}
	}
}

func (s Selection) ElemIndex() (n int) {
	if !s.IsElement() {
		return s.Index()
	}

	prev := s.Selection.PrevAll()
	prev.Each(func(_ int, e *goquery.Selection) {
		if e.Nodes[0].Type == html.ElementNode {
			n++
		}
	})

	return
}
