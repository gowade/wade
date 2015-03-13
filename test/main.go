package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/hanleym/wade"
	"github.com/hanleym/wade/driver"
	_ "github.com/hanleym/wade/react"
)

// START GENERATED //
var Classes = map[string]driver.Class{}

// END GENERATED //

var WorklogClass = wade.CreateClass(&Worklog{})
var SearchBarClass = wade.CreateClass(&SearchBar{})
var LogTableClass = wade.CreateClass(&LogTable{})
var LogRowClass = wade.CreateClass(&LogRow{})

func main() {
	wade.Render(GetWorklogClass("Worklog"), js.Global.Get("document").Call("getElementById", "worklog"))
}
