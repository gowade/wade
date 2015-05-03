package main

import (
	"strings"

	"github.com/gowade/wade"
)

type Project struct {
	wade.Com
	ID    int
	Title string
}

type WorklogState struct {
	FilterText string
	Projects   []*Project
}

type Worklog struct {
	wade.Com
	State *WorklogState `fuel:"state"`
}

func (this *Worklog) handleSearch(filterText string) {
	this.SetFilterText(strings.ToLower(filterText))
}

type SearchBar struct {
	wade.Com
	FilterText string
	OnSearch   func(string)
}

func (this *SearchBar) handleSearch() {
	this.OnSearch(this.Refs().filterTextInput.Get("value").String())
}

type LogTable struct {
	wade.Com
	FilterText string
	Projects   []*Project
}

func (this LogTable) filterCheck(text string) bool {
	return strings.Contains(strings.ToLower(text), this.FilterText)
}

type LogRow struct {
	wade.Com
	*Project
}
