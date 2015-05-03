package main

import (
	"github.com/gowade/wade"
)

func main() {
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

	wade.Render(worklog, "container")

	worklog.State.Projects[0].Title = "Oh Yeah"
	worklog.Rerender()
	worklog.Rerender()
	worklog.Rerender()
	worklog.Rerender()
	worklog.Rerender()
	worklog.Rerender()
}
