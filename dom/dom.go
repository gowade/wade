package dom

import (
	"github.com/gopherjs/gopherjs/js"
)

var (
	document        Document
	driver          Driver
	NewEventHandler func(EventHandler) interface{}
)

func SetDomDriver(drv Driver) {
	driver = drv
}

type EventHandler func(Event)

type Event interface {
	PreventDefault()
	StopPropagation()
	JS() *js.Object
}

func GetDocument() Document {
	if document == nil {
		panic(" document has not been set.")
	}
	return document
}

func CreateNode(native interface{}) Node {
	return driver.CreateNode(native)
}

type Document interface {
	Title() string
	SetTitle(title string)

	Node
}

func SetDocument(doc Document) {
	document = doc
}

type NodeType int

const (
	NopNode NodeType = iota
	ElementNode
	TextNode
)

type Node interface {
	Type() NodeType
	Find(query string) []Node
	Data() string
	Children() []Node

	SetAttr(string, interface{})
	SetProp(string, interface{})
	RemoveAttr(string)

	Clear()
	JS() *js.Object
	SetClass(string, bool)
}

type Driver interface {
	CreateNode(interface{}) Node
}

type FormEl interface {
	Node
	IsValid() bool
}

type InputEl interface {
	Node
	Value() string
	SetValue(string)
	Checked() bool
	SetChecked(bool)
}
