package main

import (
	"github.com/gowade/wade"
)

func main() {
	worklog := &Worklog{}
	worklog.State.Projects = []*Project{
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
	}

	wade.Render(wade.CreateComponent(worklog), "container")
}
