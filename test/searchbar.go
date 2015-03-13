package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/hanleym/wade"
	"github.com/hanleym/wade/driver"
)

type SearchBar struct {
	Args SearchBarArgs
	Refs SearchBarRefs
}

type SearchBarArgs struct {
	FilterText string
	OnSearch   func(filterText string)
}

type SearchBarRefs struct {
	FilterTextInput *js.Object
}

func (searchBar *SearchBar) HandleSearch() {
	searchBar.Args.OnSearch(searchBar.Refs.FilterTextInput.Call("getDOMNode").Get("value").String())
}

// START GENERATED //

func (component *SearchBar) Render() driver.Element {
	return wade.CreateElement("div", map[string]string{"class": "form-group"},
		wade.CreateElement("input", map[string]interface{}{
			"type":        "text",
			"class":       "form-control",
			"placeholder": "Search for a project...",
			"value":       component.Args.FilterText,
			"onChange":    component.HandleSearch,
			"ref":         "FilterTextInput",
		}),
	)
}
