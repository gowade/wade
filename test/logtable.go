package main

import (
	"strings"

	"github.com/hanleym/wade"
	"github.com/hanleym/wade/driver"
)

type LogTable struct {
	Args LogTableArgs
}

type LogTableArgs struct {
	FilterText string
	Projects   []Project
}

// START GENERATED //

func (component *LogTable) Render() driver.Element {
	rows := []interface{}{}
	for _, project := range component.Args.Projects {
		if !strings.Contains(strings.ToLower(project.Title), strings.ToLower(component.Args.FilterText)) {
			continue
		}
		rows = append(rows, wade.CreateElement(LogRowClass, LogRowArgs{
			Key:     project.ID,
			Project: project,
		}, nil))
	}
	return wade.CreateElement("div", nil, rows...)
}
