package dom

import (
	"errors"
	"fmt"

	"github.com/gopherjs/gopherjs/js"
)

var (
	ErrorCantGetTagName    = errors.New("Not an element node, can't get tag name.")
	ErrorNoElementSelected = errors.New("No element selected.")
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
		CurrentTarget() Selection
		DelegateTarget() Selection
		RelatedTarget() Selection
		PreventDefault()
		StopPropagation()
		KeyCode() int
		Which() int
		MetaKey() bool
		PageXY() (int, int)
		Type() string
	}

	EventHandler func(Event)

	Attr struct {
		Name  string
		Value string
	}

	Selection interface {
		Dom
		TagName() (string, error)
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
		On(Event string, handler EventHandler)
		Listen(event string, selector string, handler EventHandler)
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
		Prop(prop string, recv interface{}) bool
		SetProp(prop string, value interface{})
		Underlying() js.Object
	}
)

// DebugInfo prints debug information for the element, including
// tag name, id and parent tree
func DebugInfo(sel Selection) string {
	sel = sel.First()
	tagname, _ := sel.TagName()
	str := tagname
	if id, ok := sel.Attr("id"); ok {
		str += "#" + id
	}
	str += fmt.Sprintf(":%v", sel.ElemIndex()) + " ("
	parents := sel.Parents().Elements()
	for j := len(parents) - 1; j >= 0; j-- {
		t, err := parents[j].TagName()
		if err == nil {
			str += t + fmt.Sprintf(":%v", parents[j].ElemIndex()) + "/"
		}
	}
	str += ")"

	return str
}

// ElementError returns an error with DebugInfo on the element
func ElementError(sel Selection, errstr string) error {
	return fmt.Errorf("Error on element {%v}: %v.", DebugInfo(sel), errstr)
}

func GetElemCounterpart(elem Selection, container Selection) Selection {
	container.Parents().Length()
	parents := elem.Parents().Elements()
	tree := make([]int, 0)
	i := len(parents) - 2
	if elem.Exists() {
		i -= container.Parents().Length()
	}

	for ; i >= 0; i-- {
		tree = append(tree, parents[i].Index())
	}

	if elem.Index() != -1 {
		tree = append(tree, elem.Index())
	}

	e := container
	for _, t := range tree {
		e = e.Children().Elements()[t]
	}

	return e
}
