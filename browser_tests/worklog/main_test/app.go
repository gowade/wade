package main

import (
	"strings"

	. "github.com/phaikawl/jasmine"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"

	"github.com/gowade/wade"
	"github.com/gowade/wade/vdom"
	"github.com/gowade/wade/vdom/browser"

	. "github.com/gowade/wade/browser_tests/worklog"
)

var JQ = jquery.NewJQuery

func main() {
	Describe("Test Worklog", func() {
		worklog := &Worklog{
			State: &WorklogState{
				Projects: []*Project{
					{
						ID:    0,
						Title: "ABC",
					},
					{
						ID:    1,
						Title: "XYZ",
					},
					{
						ID:    3,
						Title: "The Great Project",
					},
					{
						ID:    4,
						Title: "Project Epic",
					},
				},
			},
		}

		ctn := JQ("<div/>").AppendTo(JQ("body"))
		vnode := wade.Render(worklog)
		vdom.PerformDiff(vnode, nil, browser.DOMNode{ctn.Get(0)})

		logTable := ctn.Find(".logtable")
		It("Should not render component tags", func() {
			Expect(ctn.Find("logtable").Length).ToBe(0)
		})

		It("Should show the page and update when an item is changed", func() {
			Expect(js.Global.Get("document").Get("title").String()).ToBe("Worklog")

			Expect(strings.HasPrefix(logTable.Find("h4").Text(), "ABCXYZ")).ToBe(true)
			Expect(ctn.Find("h4").Length).ToBe(4)

			worklog.State.Projects[0].Title = "ABC 2"
			worklog.Rerender()
			Expect(ctn.Find("h4").First().Text()).ToBe("ABC 2")
		})

		It("Should filter worklog items when the search bar is changed.", func() {
			searchBar := ctn.Find("input")
			searchBar.SetVal("&&&")
			searchBar.Trigger("change")
			Expect(ctn.Find(".row").Length).ToBe(0)

			searchBar.SetVal("EPI")
			searchBar.Trigger("change")
			Expect(logTable.Find("h4").Text()).ToBe("PROJECT EPIC")
		})
	})
}
