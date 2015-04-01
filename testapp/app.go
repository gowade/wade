package main

import (
	"github.com/gowade/wade"
)

func main() {
	wade.Render(createWorklog(nil, nil, nil), "container")
	wade.Render(createWorklog(nil, nil, nil), "container")
	wade.Render(createWorklog(nil, nil, nil), "container")
	wade.Render(createWorklog(nil, nil, nil), "container")
}
