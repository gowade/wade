package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gopherjs/gopherjs/js"

	. "github.com/gowade/wade/vdom"
	"github.com/gowade/wade/vdom/browser"
)

func main() {
	rand.Seed(time.Now().Unix())
	n := 1000
	list := make([]Node, n)
	for i := range list {
		labels := make([]Node, 5)
		for j := 0; j < 5; j++ {
			hidden := false
			if rand.Intn(2) == 1 {
				hidden = true
			}
			labels[j] = NewElement("p", Attributes{
				"hidden": hidden,
			}, []Node{
				NewTextNode(fmt.Sprint(rand.Intn(1000))),
			})
		}

		list[i] = NewElement("li", Attributes{
			"key": fmt.Sprint(rand.Intn(1000)),
		}, labels)
	}

	b := NewElement("div", nil, []Node{
		NewElement("span", nil, []Node{}),
		NewElement("ul", nil, list)})

	a := NewElement("div", nil, []Node{
		NewElement("span", nil, []Node{NewTextNode("C")}),
		NewElement("ul", nil, list),
	})

	root := js.Global.Get("document").Call("getElementById", "container")
	browser.PerformDiff(b, nil, root)
	js.Global.Get("console").Call("profile")
	browser.PerformDiff(a, b, root)
	js.Global.Get("console").Call("profileEnd")
}
