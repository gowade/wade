package jsdrv

import (
	"github.com/gowade/wade/utils/dom"
	"github.com/gowade/wade/utils/dom/jsdom"
	"github.com/gowade/wade/vdom"
	"github.com/gowade/wade/vdom/browser"
)

func Render(vnode vdom.Node, domNode dom.Node) {
	jsobj := domNode.(jsdom.Node).Object
	vdom.PerformDiff(vnode, nil, browser.DOMNode{jsobj})
}
