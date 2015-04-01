package wade

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/gowade/wade/vdom"
	"github.com/gowade/wade/vdom/browser"
)

var (
	cache *vdom.Element
)

func Render(elemId string, tree *vdom.Element) {
	elem := js.Global.Get("document").Call("getElementById", elemId)
	browser.PerformDiff(tree, cache, elem)
	cache = tree
}
