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
	SetClass(string, bool)
}

type Driver interface {
	ToInputEl(DOMNode) DOMInputEl
	ToFormEl(DOMNode) DOMFormEl
}

type DOMFormEl interface {
	DOMNode
	IsValid() bool
}

type DOMInputEl interface {
	DOMNode
	Value() string
	SetValue(string)
	Checked() bool
	SetChecked(bool)
}
