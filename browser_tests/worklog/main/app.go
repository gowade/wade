package main

import (
	"github.com/gowade/wade"

	. "github.com/gowade/wade/browser_tests/worklog"
)

func main() {
	worklog := &Worklog{
		State: &WState{
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

	r := wade.NewRouter()
	r.Handle("/", "home", func(c *wade.Context) error {
		return c.Render(worklog)
	})

	r.Handle("/hello/:name", "hello", func(c *wade.Context) error {
		return c.Render(&Hello{Name: c.Params.Get("name")})
	})

	basePath := "/"
	wade.InitApp(basePath, r, wade.FindContainer("#container"))
}
