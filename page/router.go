package page

import (
	urlrouter "github.com/naoina/kocha-urlrouter"
	_ "github.com/phaikawl/regrouter"
)

type (
	RouteHandler interface {
		UpdatePage(pm *PageManager, update pageUpdate) (found bool)
	}

	RouteEntry interface {
		Register(pm *PageManager, route string) RouteHandler
	}

	router struct {
		urlrouter.URLRouter
		routes          []urlrouter.Record
		notFoundHandler RouteHandler
	}

	Router struct {
		*router
		pm *PageManager
	}

	Redirecter struct {
		Url string
	}
)

func newRouter() *router {
	return &router{
		URLRouter: urlrouter.NewURLRouter("regexp"),
		routes:    []urlrouter.Record{},
	}
}

func (r Redirecter) Register(pm *PageManager, route string) RouteHandler {
	return r
}

func (r Redirecter) UpdatePage(pm *PageManager, update pageUpdate) (found bool) {
	return pm.updateUrl(r.Url, update.pushState, update.firstLoad)
}

func (r router) build() {
	err := r.URLRouter.Build(r.routes)
	if err != nil {
		panic(err)
	}
}

func (r router) Lookup(url string) (result RouteHandler, params []urlrouter.Param) {
	h, params := r.URLRouter.Lookup(url)
	if h != nil {
		result = h.(RouteHandler)
	}

	return
}

// Handle adds a route to the router
func (r Router) Handle(route string, action RouteEntry) {
	handler := action.Register(r.pm, route)
	r.router.routes = append(r.router.routes, urlrouter.NewRecord(route, handler))
}

func (r Router) Otherwise(action RouteEntry) {
	r.notFoundHandler = action.Register(r.pm, "")
}
