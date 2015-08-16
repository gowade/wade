package wade

import (
	"fmt"
	gourl "net/url"

	"github.com/gowade/wade/driver"
	//"github.com/gowade/wade/vdom"
)

// RouteParams holds the values of named parameters for a route
type RouteParams map[string]string

// ScanTo uses fmt.Sscan to scan the value of the given named parameter to a pointer.
func (rp RouteParams) ScanTo(dest interface{}, param string) {
	v, ok := rp[param]
	if !ok {
		panic(fmt.Errorf("ScanTo: No parameter with such name %v.", param))
	}

	fmt.Sscan(v, dest)
}

// Get returns the string value of the given named parameter
func (rp RouteParams) Get(param string) string {
	return rp[param]
}

// Context provides access to page data and page operations inside a controller function
type Context struct {
	router *DefaultRouter
	Params RouteParams
	URL    *gourl.URL
}

func (c *Context) GoToRoute(routeName string, params ...interface{}) error {
	route, ok := c.router.RouteByName(routeName)
	if !ok {
		return fmt.Errorf(`there's no route named "%v"`, routeName)
	}

	app.SetURLPath(c.router.PathFromRoute(route, params))
	return nil
}

func (c *Context) GoToRemoteURL(destURL string) error {
	url, err := gourl.Parse(destURL)
	if err != nil {
		return err
	}

	driver.GetRouteDriver().SetURL(url, false)
	return nil
}

func (c *Context) Render(component Component) error {
	//var oldVdom *vdom.Element
	//if c.router.currentComponent != nil {
	//oldVdom = c.router.currentComponent.InternalComPtr().VNode
	//}

	//vnode := Render(component)
	////driver.Render(vnode.Render().(*vdom.Element), oldVdom, app.Container)
	//c.router.currentComponent = component
	//vdom.InternalRenderUnlock()
	return nil
}
