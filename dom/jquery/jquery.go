package jquery

import (
	"fmt"
	"reflect"
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

	Event struct {
		jquery.Event
	}
)

func (e Event) Target() dom.Selection {
	return newSelection(gJQ(e.Event.Target))
}

func (e Event) CurrentTarget() dom.Selection {
	return newSelection(gJQ(e.Event.CurrentTarget))
}
func (e Event) RelatedTarget() dom.Selection {
	return newSelection(gJQ(e.Event.RelatedTarget))
}
func (e Event) DelegateTarget() dom.Selection {
	return newSelection(gJQ(e.Event.DelegateTarget))
}

func (e Event) PreventDefault() {
	e.Event.PreventDefault()
}

func (e Event) StopPropagation() {
	e.Event.StopPropagation()
}

func (e Event) KeyCode() int {
	return e.Event.KeyCode
}

func (e Event) Which() int {
	return e.Event.Which
}

func (e Event) MetaKey() bool {
	return e.Event.MetaKey
}

func (e Event) PageXY() (int, int) {
	return e.Event.PageX, e.Event.PageY
}

func (e Event) Type() string {
	return e.Event.Type
}

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

func (d Dom) NewEmptySelection() dom.Selection {
	return newSelection(gJQ())
}

func (d Dom) NewRootFragment() dom.Selection {
	return newSelection(gJQ(js.Global.Get(jquery.JQ).Call("parseHTML", "<wroot></wroot>")))
}

func (d Dom) NewTextNode(content string) dom.Selection {
	return newSelection(gJQ(js.Global.Get("document").Call("createTextNode", content)))
}

func (s Selection) TagName() (string, error) {
	if s.Length() == 0 {
		return "", dom.ErrorNoElementSelected
	}

	if !s.IsElement() {
		return "", dom.ErrorCantGetTagName
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
	r := s.NewRootFragment()
	r.Append(s.Clone())
	return r.Html()
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

func (s Selection) SetHtml(content string) {
	s.JQuery.SetHtml(content)
}

func (s Selection) SetVal(val string) {
	s.JQuery.SetVal(val)
}

func (s Selection) Val() string {
	return s.JQuery.Val()
}

func (s Selection) SetAttr(attr, value string) {
	s.JQuery.SetAttr(attr, value)
}

func (s Selection) RemoveAttr(attr string) {
	s.JQuery.RemoveAttr(attr)
}

func (s Selection) After(sel dom.Selection) {
	s.JQuery.After(sel.(Selection).JQuery)
}

func (s Selection) Next() dom.Selection {
	return newSelection(s.JQuery.Next())
}

func (s Selection) Exists() bool {
	return s.JQuery.Is("html") || s.JQuery.Is("wroot") ||
		s.JQuery.Parents("wroot").Length > 0 || s.JQuery.Parents("html").Length > 0
}

func (s Selection) Before(sel dom.Selection) {
	s.JQuery.Before(sel.(Selection).JQuery)
}

func (s Selection) Attrs() []dom.Attr {
	htmla := s.JQuery.Get(0).Get("attributes")
	attrs := make([]dom.Attr, htmla.Length())
	for i := 0; i < htmla.Length(); i++ {
		attr := htmla.Index(i)
		attrs[i].Name = attr.Get("name").Str()
		attrs[i].Value = attr.Get("value").Str()
	}

	return attrs
}

func (s Selection) Prev() dom.Selection {
	return newSelection(s.JQuery.Prev())
}

func (s Selection) On(eventname string, handler dom.EventHandler) {
	s.JQuery.On(eventname, func(event jquery.Event) {
		handler(Event{event})
	})
}

func (s Selection) Listen(event string, selector string, handler dom.EventHandler) {
	s.JQuery.On(event, selector, func(event jquery.Event) {
		handler(Event{event})
	})
}

func (s Selection) Hide() {
	s.JQuery.Hide()
}

func (s Selection) Show() {
	s.JQuery.Show()
}

func (s Selection) Filter(selector string) dom.Selection {
	return newSelection(s.JQuery.Filter(selector))
}

func (s Selection) AddClass(class string) {
	s.JQuery.AddClass(class)
}

func (s Selection) RemoveClass(class string) {
	s.JQuery.RemoveClass(class)
}

func (s Selection) Prepend(sel dom.Selection) {
	s.JQuery.Prepend(sel.(Selection).JQuery)
}

func (s Selection) ElemIndex() int {
	return s.JQuery.Underlying().Call("index").Int()
}

func (s Selection) IsTextNode() bool {
	return s.JQuery.Get(0).Get("nodeType").Int() == 3
}

func (s Selection) SetText(text string) {
	if s.IsElement() {
		s.JQuery.SetText(text)
	} else if s.IsTextNode() {
		s.JQuery.Get(0).Set("nodeValue", text)
	} else {
		js.Global.Get("console").Call("error", fmt.Sprintf("Cannot set text for this kind of node %v.", s.JQuery.Get(0).Get("nodeType").Int()))
	}
}

func (s Selection) Add(sel dom.Selection) dom.Selection {
	return newSelection(s.JQuery.Add(sel.(Selection).JQuery))
}

func (s Selection) Prop(prop string, recv interface{}) (ok bool) {
	p := s.JQuery.Underlying().Call("prop")
	if p.IsUndefined() {
		ok = false
		return
	}

	ok = true
	reflect.ValueOf(recv).Elem().Set(reflect.ValueOf(p.Interface()))
	return
}

func (s Selection) SetProp(prop string, value interface{}) {
	s.JQuery.SetProp(prop, value)
}

func (s Selection) Index() (n int) {
	e := s.JQuery.Get(0)
	for {
		if e = e.Get("previousSibling"); e.IsNull() {
			return
		}
		n++
	}
}
