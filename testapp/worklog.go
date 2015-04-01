package main

//import "github.com/gowade/wade"

type Project struct {
	ID    int
	Title string
}

type Worklog struct {
	State struct {
		FilterText string
		Projects   []*Project
	}
}

type SearchBar struct {
	Args struct {
		FilterText string
		Handler    func()
	}
}

type LogTable struct {
	Args struct {
		FilterText string
		Projects   []*Project
	}
}

type LogRow struct {
	Args struct {
		Project
	}
}
