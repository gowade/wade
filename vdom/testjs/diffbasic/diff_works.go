package main

import (
	"github.com/gopherjs/gopherjs/js"

	. "github.com/gowade/wade/vdom"
	"github.com/gowade/wade/vdom/browser"
)

func main() {
	b := NewElement("div", "", nil, []Node{
		NewElement("span", "", nil, []Node{}),
		NewElement("ul", "", nil, []Node{
			NewElement("notli", "0", nil, []Node{NewTextNode("A")}),
			NewElement("li", "5", nil, []Node{NewTextNode("B")}),
			NewElement("li", "7", Attributes{"hidden": true}, []Node{NewTextNode("E")}),
			NewElement("li", "9", nil, []Node{NewTextNode("D")}),
		})})

	a := NewElement("div", "", nil, []Node{
		NewElement("span", "", nil, []Node{NewTextNode("C")}),
		NewElement("ul", "", nil, []Node{
			NewElement("li", "5", Attributes{"hidden": true}, []Node{NewTextNode("A")}),
			NewElement("li", "9", nil, []Node{NewTextNode("D")}),
			NewElement("li", "7", Attributes{"hidden": false}, []Node{NewTextNode("E")}),
		}),
	})

	root := js.Global.Get("document").Call("getElementById", "container")
	browser.PerformDiff(b, nil, root)
	browser.PerformDiff(a, b, root)
}
