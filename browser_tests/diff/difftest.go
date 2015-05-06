package main

import (
	//"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	. "github.com/phaikawl/jasmine"

	. "github.com/gowade/wade/utils/testutils"
	. "github.com/gowade/wade/vdom"
	"github.com/gowade/wade/vdom/browser"
)

var JQ = jquery.NewJQuery

func testKeyed() {
	b := NewElement("div", "", nil, []Node{
		NewElement("span", "", nil, []Node{}),
		NewElement("ul", "", nil, []Node{
			nil,
			NewElement("notli", "", nil, []Node{NewTextNode("A")}),
			nil,
			NewElement("li", "5", nil, []Node{NewTextNode("B")}),
			NewElement("li", "7", Attributes{"hidden": true}, []Node{NewTextNode("E")}),
			NewElement("li", "", nil, []Node{NewTextNode("X")}),
			NewElement("li", "9", nil, []Node{NewTextNode("D")}),
		})})

	a := NewElement("div", "", nil, []Node{
		NewElement("span", "", nil, []Node{NewTextNode("C")}),
		NewElement("ul", "", nil, []Node{
			nil,
			NewElement("li", "5", Attributes{"hidden": true}, []Node{NewTextNode("A")}),
			NewElement("li", "9", nil, []Node{NewTextNode("D")}),
			nil,
			nil,
			NewElement("li", "", nil, []Node{NewTextNode("X")}),
			nil,
			NewElement("li", "", nil, []Node{NewTextNode("Y")}),
			NewElement("li", "70", Attributes{"hidden": false}, []Node{NewTextNode("E")}),
		}),
	})

	qroot := JQ("<div/>")
	JQ("body").Append(qroot)
	root := qroot.Get(0)
	It("should show the right elements", func() {
		browser.PerformDiff(b, nil, root)
		Expect(qroot.Find("ul").Children("").Eq(0).Prop("tagName")).ToBe("NOTLI")
		Expect(SpacesRemoved(qroot.Text())).ToBe("ABEXD")
		Expect(SpacesRemoved(qroot.Find("li:visible").Text())).ToBe("BXD")

		browser.PerformDiff(a, b, root)
		Expect(SpacesRemoved(qroot.Text())).ToBe("CADXYE")
		Expect(SpacesRemoved(qroot.Find("li:visible").Text())).ToBe("DXYE")
	})
}

func testUnkeyed() {
	b := NewElement("div", "", nil, []Node{
		NewElement("span", "", nil, []Node{}),
		NewElement("ul", "", nil, []Node{
			nil,
			NewElement("li", "", nil, []Node{NewTextNode("A")}),
			nil,
			nil,
			nil,
			NewElement("li", "", nil, []Node{NewTextNode("B")}),
			nil,
			nil,
			NewElement("li", "", nil, []Node{NewTextNode("E")}),
			nil,
			nil,
		})})

	a := NewElement("div", "", nil, []Node{
		NewElement("span", "", nil, []Node{NewTextNode("C")}),
		NewElement("ul", "", nil, []Node{
			nil,
			nil,
			nil,
			NewElement("li", "", nil, []Node{NewTextNode("A")}),
			NewElement("li", "", nil, []Node{NewTextNode("D")}),
			nil,
			nil,
			NewElement("li", "", nil, []Node{NewTextNode("X")}),
			nil,
			NewElement("li", "", nil, []Node{NewTextNode("Y")}),
			NewElement("li", "", nil, []Node{NewTextNode("E")}),
			nil,
			nil,
		}),
	})

	qroot := JQ("<div/>")
	JQ("body").Append(qroot)
	root := qroot.Get(0)
	It("should show the right elements", func() {
		browser.PerformDiff(b, nil, root)
		Expect(SpacesRemoved(qroot.Text())).ToBe("ABE")

		browser.PerformDiff(a, b, root)
		Expect(qroot.Find("ul").Children("").Eq(0).Prop("tagName")).ToBe("LI")
		Expect(SpacesRemoved(qroot.Text())).ToBe("CADXYE")
	})
}

func main() {
	Describe("keyed diff", testKeyed)
	Describe("unkeyed diff", testUnkeyed)
}
