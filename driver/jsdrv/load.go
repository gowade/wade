package jsdrv

import (
	//"github.com/gopherjs/gopherjs/js"

	_ "github.com/gowade/wade/dom/jsdom"
	"github.com/gowade/wade/driver"
	_ "github.com/gowade/wade/driver/jsdrv/shim"
	_ "github.com/gowade/wade/utils/http/jshttp"
)

func init() {
	driver.Render = Render

	driver.SetRouteDriver(getRouteDriver())
	driver.SetEnv(driver.BrowserEnv)
}
