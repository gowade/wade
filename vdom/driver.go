package vdom

type DOMNode interface {
	Child(int) DOMNode
	SetAttr(string, interface{})
	SetProp(string, interface{})
	RemoveAttr(string)
	Do(Action)
}

type Driver interface {
	PerformDiff(a, b Node, dNode DOMNode)
	ToInputEl(DOMNode) DOMInputEl
}

type DOMInputEl interface {
	Value() string
	SetValue(string)
}
