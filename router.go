package wade

import (
	"bytes"
	"fmt"
	gourl "net/url"
	"path"
	"strings"

	urlrouter "github.com/naoina/kocha-urlrouter"
	_ "github.com/phaikawl/regRouter"
)

type (
	ControllerFunc func(*Context) error

	DefaultRouter struct {
		*defaultRouter
		nameMap      map[string]string
		errorHandler func(error)
	}

	defaultRouter struct {
		urlrouter.URLRouter
		routes          []urlrouter.Record
		notFoundHandler ControllerFunc
	}
)

func NewRouter() *DefaultRouter {
	return &DefaultRouter{
		nameMap:       map[string]string{},
		defaultRouter: newRouter(),
	}
}

func (r *DefaultRouter) RouteByName(routeName string) (route string, ok bool) {
	route, ok = r.nameMap[routeName]
	return
}

func (r *DefaultRouter) SetErrorHandler(handler func(error)) {
	r.errorHandler = handler
}

// Handle registers a handler for the given route, routeName is a unique name
// that you assign to a route, to be used in links for example
func (r *DefaultRouter) Handle(route string, routeName string, handler ControllerFunc) {
	if routeName == "" {
		panic("routeName cannot be empty")
	}

	if r.nameMap[routeName] != "" {
		panic(fmt.Errorf(`routeName "%v" is already taken`))
	}

	r.defaultRouter.handle(route, handler)
	r.nameMap[routeName] = route
}

func (r *DefaultRouter) SetNotFoundHandler(c ControllerFunc) {
	r.defaultRouter.setNotFoundHandler(c)
}

func (r *DefaultRouter) URLFromRoute(route string, params ...interface{}) string {
	routeparams := urlrouter.ParamNames(route)
	if len(routeparams) != len(params) {
		panic(fmt.Errorf(`Wrong number of parameters for route "%v". Expected %v, got %v.`,
			route, len(routeparams), len(params)))
	}

	var url bytes.Buffer
	var k, i int
	for {
		if i >= len(route) {
			break
		}

		if urlrouter.IsMetaChar(route[i]) && route[i:i+len(routeparams[k])] == routeparams[k] {
			param := routeparams[k]
			if k < len(params) && params[k] != nil {
				url.WriteString(fmt.Sprint(params[k]))
			}

			i += len(param)
			k++
		} else {
			url.WriteByte(route[i])
			i++
		}
	}

	if k != len(routeparams) || k != len(params) {
		panic(fmt.Errorf(`Wrong number of parameters for route "%v". Expected %v, got %v.`,
			route, len(routeparams), len(params)))
	}

	return url.String()
}

func (r *DefaultRouter) Lookup(path string) (interface{}, map[string]string) {
	cf, params := r.defaultRouter.lookup(path)
	var rp map[string]string
	for _, param := range params {
		rp[param.Name] = param.Value
	}

	return cf, rp
}

func (r *DefaultRouter) Render(url *gourl.URL) {
	tpath := path.Join("/", strings.TrimPrefix(url.Path, app.BasePath))

	handler, params := r.Lookup(tpath)
	if handler == nil {
		if r.notFoundHandler == nil {
			panic(fmt.Errorf(
				"No suitable handler can be found for %v. notFoundHandler has not been set.",
				tpath))
		}

		handler = r.notFoundHandler
	}

	cf := handler.(ControllerFunc)
	err := cf(&Context{
		router: r,
		URL:    url,
		Params: params,
	})

	if err != nil {
		if r.errorHandler == nil {
			panic(err)
		}

		r.errorHandler(err)
	}
}

func (r *DefaultRouter) Build() {
	r.defaultRouter.build()
}

func newRouter() *defaultRouter {
	return &defaultRouter{
		URLRouter: urlrouter.NewURLRouter("regexp"),
		routes:    []urlrouter.Record{},
	}
}

func (r *defaultRouter) build() {
	err := r.URLRouter.Build(r.routes)
	if err != nil {
		panic(err)
	}
}

func (r *defaultRouter) lookup(path string) (interface{}, []urlrouter.Param) {
	return r.URLRouter.Lookup(path)
}

// Handle adds a route to the Router
func (r *defaultRouter) handle(route string, c ControllerFunc) urlrouter.Record {
	record := urlrouter.NewRecord(route, c)
	r.routes = append(r.routes, record)
	return record
}

func (r *defaultRouter) setNotFoundHandler(c ControllerFunc) {
	r.notFoundHandler = c
}
