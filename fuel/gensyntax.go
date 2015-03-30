package main

const (
	CreateElementOpener  = "vdom.NewElement"
	CreateTextNodeOpener = "vdom.NewTextNode"
	AttributeMapOpener   = "vdom.Attributes"
	ElementListOpener    = "[]vdom.Node"
	RenderFuncOpener     = "func Render() "
)

const Prelude = `package template
import (
	"fmt"
	"github.com/gowade/wade/vdom"
)
`
