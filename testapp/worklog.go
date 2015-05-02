package main

import "github.com/gowade/wade"

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

func (this *Worklog) handleSearch() {
}

type SearchBar struct {
	wade.Com
	FilterText string
	Handler    func()
}

type LogTable struct {
	wade.Com
	FilterText string
	Projects   []*Project
}

type LogRow struct {
	wade.Com
	*Project
}
