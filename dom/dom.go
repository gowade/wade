package dom

import (
	"fmt"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	"github.com/phaikawl/wade/core"
)

type (
	Dom interface {
		NewFragment(html string) Selection
		NewRootFragment() Selection
		NewEmptySelection() Selection
		NewTextNode(content string) Selection
	}

	Event interface {
		Target() Selection
		PreventDefault()
		StopPropagation()
		Type() string
		Js() js.Object
	}

	KeyEvent interface {
		Event
		Which() int
	}

	MouseEvent interface {
		Event
		Which() int
		Pos() (int, int)
	}

	EventHandlerFn func(Event)
	EachFn         func(i int, elem Selection)

	Attr struct {
		Name  string
		Value string
	}

	Selection interface {
		Dom
		TagName() string
		Filter(selector string) Selection
		Children() Selection
		Contents() Selection
		First() Selection
		IsElement() bool
		Find(selector string) Selection
		Html() string
		SetHtml(html string)
		Length() int
		Elements() []Selection
		Append(Selection)
		Prepend(Selection)
		Remove()
		Clone() Selection
		ReplaceWith(Selection)
		OuterHtml() string
		Attr(attr string) (string, bool)
		SetAttr(attr string, value string)
		RemoveAttr(attr string)
		Val() string
		SetVal(val string)
		Parents() Selection
		Is(selector string) bool
		Unwrap()
		Parent() Selection
		Next() Selection
		Prev() Selection
		Before(sel Selection)
		After(sel Selection)
		Exists() bool
		Attrs() []Attr
		Listen(event string, selector string, handler EventHandlerFn)
		Hide()
		Show()
		AddClass(class string)
		RemoveClass(class string)
		HasClass(class string) bool
		Text() string
		Index() int
		ElemIndex() int
		IsTextNode() bool
		SetText(text string)
		Add(element Selection) Selection
		Prop(prop string) (interface{}, bool)
		SetProp(prop string, value interface{})
		Underlying() js.Object
		Each(EachFn)
		Render(*core.VNode)
		ToVNode() *core.VNode
	}
)

func DebugHtml(sel Selection) (s string) {
	for _, elem := range sel.Elements() {
		s += debugHtml(elem, 0, false)
	}

	return
}

func debugHtml(elem Selection, idx int, inline bool) (s string) {
	indent := ""
	for i := 0; i < idx; i++ {
		indent += "  "
	}

	attrs := ""
	for _, attr := range elem.Attrs() {
		attrs += fmt.Sprintf(` %v="%v"`, attr.Name, attr.Value)
	}

	if elem.IsElement() {
		s += indent
		s += fmt.Sprintf("<%v%v>", elem.TagName(), attrs)
		inline := elem.Contents().Length() == 1 && elem.Contents().First().IsTextNode()
		if !inline {
			s += "\n"
		}

		for _, c := range elem.Contents().Elements() {
			s += debugHtml(c, idx+1, inline)
		}

		if !inline {
			s += indent
		}

		s += fmt.Sprintf("</%v>", elem.TagName())
		s += "\n"
	}

	if elem.IsTextNode() {
		text := strings.TrimSpace(elem.Text())
		if text == "" {
			return
		}

		if !inline {
			s += indent
		}
		s += "`" + text + "`"
		if !inline {
			s += "\n"
		}
	}

	return
}

// DebugInfo prints debug information for the element, including
// tag name, id and parent tree
func DebugInfo(sel Selection) string {
	sel = sel.First()
	tagname := sel.TagName()
	str := tagname
	if id, ok := sel.Attr("id"); ok {
		str += "#" + id
	}
	str += fmt.Sprintf(":%v", sel.ElemIndex()) + " ("
	parents := sel.Parents().Elements()
	for j := len(parents) - 1; j >= 0; j-- {
		t := parents[j].TagName()
		str += t + fmt.Sprintf(":%v", parents[j].ElemIndex()) + "/"
	}
	str += ")"

	return str
}

// ElementError returns an error with DebugInfo on the element
func ElementError(sel Selection, errstr string) error {
	return fmt.Errorf("Error on element {%v}: %v.", DebugInfo(sel), errstr)
}
