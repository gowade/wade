package wade

import (
	"fmt"
	gourl "net/url"
	"path"

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

func (z Application) SetURLPath(newPath string) {
	url, err := gourl.Parse(path.Join(app.BasePath, newPath))
	if err != nil {
		panic(err)
	}

	driver.GetRouteDriver().SetURL(url, true)
}

func InitApp(basepath string, router driver.Router, container dom.Node) {
	if !path.IsAbs(basepath) {
		panic(fmt.Errorf(`application base path `+
			`must be a valid absolute path, got "%v"`, basepath))
	}

	app = Application{
		BasePath:  path.Clean(basepath),
		Router:    router,
		Container: container,
	}
	driver.Init(router)
	router.Build()

	url := driver.GetRouteDriver().URL()
	router.Render(url)
}

func Route(routeName string, params ...interface{}) string {
	if app.Router == nil {
		return "/"
	}

	route, ok := app.Router.RouteByName(routeName)
	if !ok {
		panic(fmt.Errorf(`there's no route named "%v"`, routeName))
	}

	return app.Router.PathFromRoute(route, params...)
}

func FindContainer(query string) dom.Node {
	ret := dom.GetDocument().Find(query)
	if len(ret) == 0 {
		panic(fmt.Errorf(`No DOM element found for query "%v"`, query))
	}

	return ret[0]
}
