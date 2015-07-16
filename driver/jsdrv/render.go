package jsdrv

import (
	"github.com/gowade/wade/utils/dom"
	"github.com/gowade/wade/utils/dom/jsdom"
	"github.com/gowade/wade/vdom"
	"github.com/gowade/wade/vdom/browser"
)

func Render(newVdom, oldVdom *vdom.Element, domNode dom.Node) {
	var container vdom.DOMNode
	if oldVdom != nil {
		oldVdom = oldVdom.Render().(*vdom.Element)
		//vdom.Debug(oldVdom)
		//vdom.Debug(newVdom)
		container = oldVdom.DOMNode()
	} else {
		jsobj := domNode.(jsdom.Node).Object
		container = browser.DOMNode{jsobj}
	}

	vdom.PerformDiff(newVdom, oldVdom, container)
}
