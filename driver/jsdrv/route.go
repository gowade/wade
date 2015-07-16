package jsdrv

import (
	gourl "net/url"

	"github.com/gopherjs/gopherjs/js"

	"github.com/gowade/wade/driver"
)

func getRouteDriver() *routeDriver {
	hist := js.Global.Get("history")
	if hist == js.Undefined {
		panic("No HTML5 history object available.")
	}

	return &routeDriver{
		history: history{hist},
	}
}

type routeDriver struct {
	router driver.Router
	history
}

func (rd *routeDriver) url() string {
	location := rd.history.location()
	return location.Get("href").String()
}

func (rd *routeDriver) URL() *gourl.URL {
	url, err := gourl.Parse(rd.url())
	if err != nil {
		panic("This cannot happen, something is wrong.")
	}

	return url
}

func (rd *routeDriver) setURL(url *gourl.URL, local bool, pushState bool) {
	if !local {
		rd.history.redirectTo(url.String())
		return
	}

	if pushState {
		rd.history.pushState("", url.Path)
	}
	rd.router.Render(url)
}

func (rd *routeDriver) SetURL(url *gourl.URL, local bool) {
	rd.setURL(url, local, true)
}

func (rd *routeDriver) Init(router driver.Router) {
	rd.router = router

	rd.history.onPopState(func() {
		rd.setURL(rd.URL(), true, false)
	})
}

type history struct {
	*js.Object
}

func (h history) replaceState(title, path string) {
	h.Object.Call("replaceState", nil, title, path)
}

func (h history) pushState(title, path string) {
	h.Object.Call("pushState", nil, title, path)
}

func (h history) location() *js.Object {
	location := h.Get("location")
	if location == nil || location == js.Undefined {
		location = js.Global.Get("document").Get("location")
	}

	return location
}

func (h history) onPopState(fn func()) {
	js.Global.Get("window").Call("addEventListener", "popstate", fn)
}

func (h history) redirectTo(url string) {
	js.Global.Get("window").Get("location").Set("href", url)
}
