package main

import (
	"github.com/gopherjs/gopherjs/js"

	. "github.com/gowade/wade/vdom"
	"github.com/gowade/wade/vdom/browser"
)

func main() {
	b := NewElement("div", nil, []Node{
		NewElement("span", nil, []Node{}),
		NewElement("ul", nil, []Node{
			NewElement("notli", Attributes{"key": 0}, []Node{NewTextNode("A")}),
			NewElement("li", Attributes{"key": 5}, []Node{NewTextNode("B")}),
			NewElement("li", Attributes{"key": 7, "hidden": true}, []Node{NewTextNode("E")}),
			NewElement("li", Attributes{"key": 9}, []Node{NewTextNode("D")}),
		})})

	a := NewElement("div", nil, []Node{
		NewElement("span", nil, []Node{NewTextNode("C")}),
		NewElement("ul", nil, []Node{
			NewElement("li", Attributes{"key": 5, "hidden": true}, []Node{NewTextNode("A")}),
			NewElement("li", Attributes{"key": 9}, []Node{NewTextNode("D")}),
			NewElement("li", Attributes{"key": 7, "hidden": false}, []Node{NewTextNode("E")}),
		}),
	})

	root := js.Global.Get("document").Call("getElementById", "container")
	browser.PerformDiff(b, nil, root)
	browser.PerformDiff(a, b, root)
}
