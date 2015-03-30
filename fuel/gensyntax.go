package main

const (
	CreateElementOpener  = "vdom.NewElement"
	CreateTextNodeOpener = "vdom.NewTextNode"
	AttributeMapOpener   = "vdom.Attrs"
	ElementListOpener    = "[]vdom.Node"
	RenderFuncOpener     = "func Render() "
)

const Prelude = `package main

import "github.com/gowade/wade/vdom"
`
