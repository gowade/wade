package wade

import (
	urlrouter "github.com/naoina/kocha-urlrouter"
	_ "github.com/phaikawl/regrouter"
)

type (
	RouteHandler interface {
		UpdatePage(pm *pageManager, update pageUpdate) (found bool)
	}

	RouteEntry interface {
		Register(pm *pageManager, route string) RouteHandler
	}

	Router struct {
		urlrouter.URLRouter
		pm              *pageManager
		routes          []urlrouter.Record
		notFoundHandler RouteHandler
	}

	Redirecter struct {
		Url string
	}
)

func (r Redirecter) Register(pm *pageManager, route string) RouteHandler {
	return r
}

func (r Redirecter) UpdatePage(pm *pageManager, update pageUpdate) (found bool) {
	return pm.updateUrl(r.Url, update.pushState, update.firstLoad)
}

func newRouter(pm *pageManager) *Router {
	return &Router{
		URLRouter: urlrouter.NewURLRouter("regexp"),
		pm:        pm,
		routes:    []urlrouter.Record{},
	}
}

func (r *Router) Build() {
	r.URLRouter.Build(r.routes)
}

func (r *Router) Handle(route string, entry RouteEntry) *Router {
	handler := entry.Register(r.pm, route)
	r.routes = append(r.routes, urlrouter.NewRecord(route, handler))

	return r
}

func (r *Router) Otherwise(entry RouteEntry) {
	r.notFoundHandler = entry.Register(r.pm, "")
}

func (r *Router) Lookup(url string) (result RouteHandler, params []urlrouter.Param) {
	h, params := r.URLRouter.Lookup(url)
	if h != nil {
		result = h.(RouteHandler)
	}

	return
}
