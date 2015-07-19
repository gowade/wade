package driver

import (
	gourl "net/url"

	"github.com/gowade/wade/utils/dom"
	"github.com/gowade/wade/vdom"
)

type EnvironmentType int

const (
	BrowserEnv EnvironmentType = iota
	ServerEnv
)

var (
	env         EnvironmentType = BrowserEnv
	vdomDriver  vdom.Driver
	routeDriver RouteDriver
	Render      func(newVdom, oldVdom *vdom.Element, domNode dom.Node)
)

func Init(router Router) {
	if vdomDriver == nil {
		panic("DOM Driver has not been set.")
	}

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

func Vdom() vdom.Driver {
	if vdomDriver == nil {
		panic("DOM driver not set.")
	}

	return vdomDriver
}

func SetVdomDriver(d vdom.Driver) {
	vdomDriver = d
}

func Env() EnvironmentType {
	return env
}

func SetEnv(envType EnvironmentType) {
	env = envType
}
