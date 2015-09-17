package driver

import (
	gourl "net/url"

	"github.com/gowade/vdom"
	"github.com/gowade/wade/dom"
)

type EnvironmentType int

const (
	BrowserEnv EnvironmentType = iota
	ServerEnv
)

var (
	env         EnvironmentType = BrowserEnv
	routeDriver RouteDriver
	Render      func(newVdom, oldVdom vdom.VNode, domNode dom.Node)
)

func Init(router Router) {
	if routeDriver == nil {
		panic("Route Driver has not been set")
	}

	routeDriver.Init(router)
}

type Router interface {
	PathFromRoute(route string, params ...interface{}) string
	RouteByName(name string) (route string, ok bool)
	Render(url *gourl.URL)
	Build()
}

type RouteDriver interface {
	Init(Router)
	URL() *gourl.URL
	SetURL(url *gourl.URL, local bool)
}

func GetRouteDriver() RouteDriver {
	return routeDriver
}

func SetRouteDriver(drv RouteDriver) {
	routeDriver = drv
}

func Env() EnvironmentType {
	return env
}

func SetEnv(envType EnvironmentType) {
	env = envType
}
