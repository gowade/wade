package main

import (
	"github.com/gopherjs/jquery"
	. "github.com/phaikawl/busterjs"

	. "github.com/gowade/wade/utils/testutils"
	. "github.com/gowade/wade/vdom"
	"github.com/gowade/wade/vdom/browser"
)

var JQ = jquery.NewJQuery

func testBasic() {
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

	root := JQ("body")
	It("should show the right elements", func() {
		browser.PerformDiff(b, nil, root.Get(0))
		Expect(root.Find("ul").Children("").Eq(0).Prop("tagName")).ToEqual("NOTLI")
		Expect(SpacesRemoved(root.Text())).ToEqual("ABED")
		Expect(SpacesRemoved(root.Find("li:visible").Text())).ToEqual("BD")

		browser.PerformDiff(a, b, root.Get(0))
		Expect(SpacesRemoved(root.Text())).ToEqual("CADE")
		Expect(SpacesRemoved(root.Find("li:visible").Text())).ToEqual("DE")
	})
}

func main() {
	Describe("basic list diff", testBasic)
}
