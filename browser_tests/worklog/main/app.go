package main

import (
	"github.com/gowade/wade"
	"github.com/gowade/wade/driver"
	_ "github.com/gowade/wade/driver/jsdrv"

	. "github.com/gowade/wade/browser_tests/worklog"
)

func main() {
	worklog := &Worklog{
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
	}

	ctn := wade.FindContainer("#container")
	vnode := worklog.VDOMRender()
	driver.Render(vnode, nil, ctn)
}
