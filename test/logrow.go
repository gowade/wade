package main

import (
	"github.com/hanleym/wade"
	"github.com/hanleym/wade/driver"
)

type LogRow struct {
	Args LogRowArgs
}

type LogRowArgs struct {
	Key     int
	Project Project
}

// START GENERATED //

func (component *LogRow) Render() driver.Element {
	return wade.CreateElement("div", nil,
		wade.CreateElement("div", nil,
			wade.CreateElement("div", map[string]string{"class": "col-xs-7"},
				wade.CreateElement("h4", nil, component.Args.Project.Title),
			),
			wade.CreateElement("div", map[string]string{"class": "col-xs-2 text-right"}, "00:00:00"),
			wade.CreateElement("div", map[string]string{"class": "col-xs-3 text-right"}, "BTN"),
			wade.CreateElement("hr", nil, nil),
		),
	)
}
