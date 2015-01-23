package jsdom

import (
	"fmt"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/phaikawl/wade/core"
	"github.com/phaikawl/wade/dom"
)

var (
	gDom      = Dom{}
	gJQ       = jquery.NewJQuery
	EventChan = make(chan jqEvent)
)

type (
	Selection struct {
		jquery.JQuery
		Dom
	}

	Dom struct{}

	jqEvent struct {
		jquery.Event
	}
)

func createEvent(orig js.Object) jqEvent {
	orig = js.Global.Get(jquery.JQ).Get("event").Call("fix", orig)
	//println(orig.Get("preventDefault"))
	return jqEvent{jquery.Event{Object: orig}}
}

func (e jqEvent) Target() dom.Selection {
	return NewSelection(gJQ(e.Event.Target))
}

func (e jqEvent) PreventDefault() {
	e.Event.PreventDefault()
}

func (e jqEvent) StopPropagation() {
	e.Event.StopPropagation()
}

func (e jqEvent) Which() int {
	return e.Event.Which
}

func (e jqEvent) Pos() (int, int) {
	return e.Event.PageX, e.Event.PageY
}

func (e jqEvent) Type() string {
	return e.Event.Type
}

func (e jqEvent) Js() js.Object {
	return e.Event.Object
}

func GetDom() dom.Dom {
	return gDom
}

func Document() dom.Selection {
	return NewSelection(gJQ(js.Global.Get("document")))
}

func NewSelection(jq jquery.JQuery) dom.Selection {
	return Selection{jq, gDom}
}

func (d Dom) NewFragment(html string) dom.Selection {
	return NewSelection(gJQ(html))
}

func (d Dom) NewEmptySelection() dom.Selection {
	return NewSelection(gJQ())
}

func (d Dom) NewRootFragment() dom.Selection {
	return NewSelection(gJQ(js.Global.Get(jquery.JQ).Call("parseHTML", "<wroot></wroot>")))
}

func (d Dom) NewTextNode(content string) dom.Selection {
	return NewSelection(gJQ(js.Global.Get("document").Call("createTextNode", content)))
}

func (s Selection) TagName() string {
	if !s.IsElement() {
		return ""
	}

	return s.JQuery.Underlying().Call("prop", "tagName").Call("toLowerCase").String()
}

func (s Selection) Children() dom.Selection {
	return NewSelection(s.JQuery.Children(""))
}

func (s Selection) Contents() dom.Selection {
	return NewSelection(s.JQuery.Contents())
}

func (s Selection) First() dom.Selection {
	return NewSelection(s.JQuery.First())
}

func (s Selection) IsElement() bool {
	return s.JQuery.Get(0).Get("nodeType").Int() == 1
}

func (s Selection) Find(selector string) dom.Selection {
	return NewSelection(s.JQuery.Find(selector))
}

func (s Selection) Length() int {
	return s.JQuery.Length
}

func (s Selection) Elements() []dom.Selection {
	list := make([]dom.Selection, s.Length())
	u := s.JQuery.Underlying()
	for i := 0; i < s.JQuery.Length; i++ {
		list[i] = NewSelection(gJQ(u.Index(i)))
	}

	return list
}

func (s Selection) Each(fn dom.EachFn) {
	u := s.JQuery.Underlying()
	for i := 0; i < s.JQuery.Length; i++ {
		fn(i, NewSelection(gJQ(u.Index(i))))
	}
}

func (s Selection) Append(c dom.Selection) {
	s.JQuery.Append(c.(Selection).JQuery)
}

func (s Selection) Remove() {
	s.JQuery.Remove()
}

func (s Selection) Clone() dom.Selection {
	return NewSelection(s.JQuery.Clone())
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
	if ret == js.Undefined {
		return "", false
	}

	return ret.String(), true
}

func (s Selection) Parents() dom.Selection {
	return NewSelection(s.JQuery.Parents())
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
	return NewSelection(s.JQuery.Parent())
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
	return NewSelection(s.JQuery.Next())
}

func (s Selection) Exists() bool {
	return s.JQuery.Closest("wroot").Length > 0 || s.JQuery.Closest("html").Length > 0
}

func (s Selection) Before(sel dom.Selection) {
	s.JQuery.Before(sel.(Selection).JQuery)
}

func (s Selection) Attrs() []dom.Attr {
	htmla := s.JQuery.Get(0).Get("attributes")
	attrs := make([]dom.Attr, htmla.Length())
	for i := 0; i < htmla.Length(); i++ {
		attr := htmla.Index(i)
		attrs[i].Name = attr.Get("name").String()
		attrs[i].Value = attr.Get("value").String()
	}

	return attrs
}

func (s Selection) Prev() dom.Selection {
	return NewSelection(s.JQuery.Prev())
}

func (s Selection) On(eventname string, handler dom.EventHandlerFn) {
	s.JQuery.On(eventname, func(event jquery.Event) {
		evt := jqEvent{event}
		handler(evt)

		select {
		case EventChan <- evt:
		default:
		}
	})
}

func (s Selection) Listen(event string, selector string, handler dom.EventHandlerFn) {
	s.JQuery.On(event, selector, func(event jquery.Event) {
		handler(jqEvent{event})
	})
}

func (s Selection) Hide() {
	s.JQuery.Hide()
}

func (s Selection) Show() {
	s.JQuery.Show()
}

func (s Selection) Filter(selector string) dom.Selection {
	return NewSelection(s.JQuery.Filter(selector))
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
		panic(fmt.Sprintf("Cannot set text for this kind of node %v.", s.JQuery.Get(0).Get("nodeType").Int()))
	}
}

func (s Selection) Add(sel dom.Selection) dom.Selection {
	return NewSelection(s.JQuery.Add(sel.(Selection).JQuery))
}

func (s Selection) Prop(prop string) (value interface{}, ok bool) {
	p := s.JQuery.Underlying().Call("prop", prop)
	if p == js.Undefined {
		ok = false
		return
	}

	value = p.Interface()
	return
}

func (s Selection) SetProp(prop string, value interface{}) {
	s.JQuery.SetProp(prop, value)
}

func (s Selection) Index() (n int) {
	e := s.JQuery.Get(0)
	for {
		if e = e.Get("previousSibling"); e == nil {
			return
		}
		n++
	}
}

func (s Selection) Underlying() js.Object {
	return s.JQuery.Underlying()
}

func (sel Selection) Render(vn *core.VNode) {
	Render(sel.Get(0), vn)
}

func (sel Selection) ToVNode() *core.VNode {
	return ToVNode(sel.Get(0))
}
