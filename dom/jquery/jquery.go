package jquery

import (
	"strings"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/phaikawl/wade/dom"
)

var (
	gDom = Dom{}
	gJQ  = jquery.NewJQuery
)

type (
	Selection struct {
		jquery.JQuery
		Dom
	}

	Dom struct{}
)

func GetDom() dom.Dom {
	return gDom
}

func Document() dom.Selection {
	return newSelection(gJQ(js.Global.Get("document")))
}

func newSelection(jq jquery.JQuery) dom.Selection {
	return Selection{jq, gDom}
}

func (d Dom) NewFragment(html string) dom.Selection {
	return newSelection(gJQ(html))
}

func (d Dom) NewRootFragment(html string) dom.Selection {
	return newSelection(gJQ(js.Global.Get(jquery.JQ).Call("parseHTML", html)))
}

func (s Selection) TagName() (string, error) {
	if s.Length() == 0 {
		return "", dom.ErrorNoElementSelected
	}

	tn := s.JQuery.First().Prop("tagName")
	if tag, ok := tn.(string); ok {
		return strings.ToLower(tag), nil
	}

	return "", dom.ErrorCantGetTagName
}

func (s Selection) Children() dom.Selection {
	return newSelection(s.JQuery.Children(""))
}

func (s Selection) Contents() dom.Selection {
	return newSelection(s.JQuery.Contents())
}

func (s Selection) First() dom.Selection {
	return newSelection(s.JQuery.First())
}

func (s Selection) IsElement() bool {
	return s.JQuery.Get(0).Get("nodeType").Int() == 1
}

func (s Selection) Find(selector string) dom.Selection {
	return newSelection(s.JQuery.Find(selector))
}

func (s Selection) Length() int {
	return s.JQuery.Length
}

func (s Selection) Elements() []dom.Selection {
	list := make([]dom.Selection, s.Length())
	s.JQuery.Each(func(i int, elem jquery.JQuery) {
		list[i] = newSelection(elem)
	})

	return list
}

func (s Selection) Append(c dom.Selection) {
	s.JQuery.Append(c.(Selection).JQuery)
}

func (s Selection) Remove() {
	s.JQuery.Remove()
}

func (s Selection) Clone() dom.Selection {
	return newSelection(s.JQuery.Clone())
}

func (s Selection) ReplaceWith(sel dom.Selection) {
	s.JQuery.ReplaceWith(sel.(Selection).JQuery)
}

func (s Selection) OuterHtml() string {
	return gJQ("<div>").Append(s.Clone()).Html()
}

func (s Selection) Attr(attr string) (string, bool) {
	ret := s.JQuery.Underlying().Call("attr", attr)
	if ret.IsUndefined() {
		return "", false
	}

	return ret.Str(), true
}

func (s Selection) Parents() dom.Selection {
	return newSelection(s.JQuery.Parents())
}

func (s Selection) Is(selector string) bool {
	return s.JQuery.Is(selector)
}

func (s Selection) Unwrap() {
	if s.Children().Length() == 0 {
		return
	}

	s.Children().First().(Selection).JQuery.Unwrap()
}

func (s Selection) Parent() dom.Selection {
	return newSelection(s.JQuery.Parent())
}
