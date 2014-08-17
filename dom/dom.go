package dom

import (
	"errors"
	"fmt"
)

var (
	ErrorCantGetTagName    = errors.New("Not an element node, can't get tag name.")
	ErrorNoElementSelected = errors.New("No element selected.")
)

type (
	Dom interface {
		NewFragment(html string) Selection
		NewRootFragment(html string) Selection
	}

	Selection interface {
		TagName() (string, error)
		Children() Selection
		Contents() Selection
		First() Selection
		IsElement() bool
		Find(selector string) Selection
		Html() string
		Length() int
		Elements() []Selection
		Append(Selection)
		Remove()
		Clone() Selection
		ReplaceWith(Selection)
		OuterHtml() string
		Attr(attr string) (string, bool)
		Parents() Selection
		Is(selector string) bool
		Unwrap()
		Parent() Selection
		Dom
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
	str += " ("
	parents := sel.Parents().Elements()
	for j := len(parents) - 1; j >= 0; j-- {
		t, err := parents[j].TagName()
		if err == nil {
			str += t + ">"
		}
	}
	str += ")"

	return str
}

// ElementError returns an error with DebugInfo on the element
func ElementError(sel Selection, errstr string) error {
	return fmt.Errorf("Error on element {%v}: %v.", DebugInfo(sel), errstr)
}
