package main

import (
	"github.com/hanleym/wade"
	"github.com/hanleym/wade/driver"
)

type Worklog struct {
	State WorklogState
}

type WorklogState struct {
	FilterText string
	Projects   []Project
}

func (worklog *Worklog) HandleSearch(filterText string) {
	worklog.State.FilterText = filterText
	//rerender
}

// START GENERATED //

func GetWorklogClass(name string) driver.Class {
	if _, found := Classes[name]; !found {
		Classes[name] = wade.CreateClass(&Worklog{})
	}
	return Classes[name]
}

func (component *Worklog) Render() driver.Element {
	return wade.CreateElement("div", nil,
		wade.CreateElement("h2", nil, "Worklog"),
		wade.CreateElement(SearchBarClass, SearchBarArgs{
			FilterText: component.State.FilterText,
			OnSearch:   component.HandleSearch,
		}),
		wade.CreateElement(LogTableClass, LogTableArgs{
			FilterText: component.State.FilterText,
			Projects:   component.State.Projects,
		}),
	)
}
