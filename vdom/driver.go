package vdom

import (
	"github.com/gopherjs/gopherjs/js"
)

type DOMNode interface {
	Child(int) DOMNode
	SetAttr(string, interface{})
	SetProp(string, interface{})
	RemoveAttr(string)
	Do(*Action)
	Clear()
	Render(Node, bool)
	JS() *js.Object
	Compat(Node) bool
}

type Driver interface {
	ToInputEl(DOMNode) DOMInputEl
}

type DOMInputEl interface {
	Value() string
	SetValue(string)
	Checked() bool
	SetChecked(bool)
}
