package jsdrv

import (
	//"github.com/gopherjs/gopherjs/js"

	"github.com/gowade/wade/driver"
	_ "github.com/gowade/wade/driver/jsdrv/shim"
	_ "github.com/gowade/wade/utils/dom/jsdom"
	_ "github.com/gowade/wade/utils/http/jshttp"
	_ "github.com/gowade/wade/vdom/browser"
)

func init() {
	driver.Render = Render
	driver.SetRouteDriver(getRouteDriver())
	driver.SetEnv(driver.BrowserEnv)
}
