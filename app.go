package wade

import (
	"fmt"

	"github.com/gowade/wade/driver"
	"github.com/gowade/wade/utils/dom"
)

var (
	app Application
)

type Application struct {
	BasePath  string
	Router    driver.Router
	Container dom.Node
}

func App() Application {
	return app
}

func InitApp(basepath string, router driver.Router, container dom.Node) {
	app = Application{
		BasePath:  basepath,
		Router:    router,
		Container: container,
	}
	driver.Init(router)
	router.Build()

	url := driver.GetRouteDriver().URL()
	router.Render(url)
}

func FindContainer(query string) dom.Node {
	ret := dom.Document().Find(query)
	if len(ret) == 0 {
		panic(fmt.Errorf(`No DOM element found for query "%v"`, query))
	}

	return ret[0]
}
