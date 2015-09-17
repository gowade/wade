package jsdrv

import (
	"github.com/gowade/wade/dom"
	//"github.com/gowade/wade/dom/jsdom"
	"github.com/gowade/vdom"
)

func Render(newVdom, oldVdom vdom.VNode, domNode dom.Node) {
	diff := vdom.Diff(oldVdom, newVdom)
	vdom.Patch(domNode.JS(), diff)
}
