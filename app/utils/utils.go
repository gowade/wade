package utils

import (
	"fmt"

	"github.com/phaikawl/wade/app"
)

var Fmt = fmt.Sprintf

func Url(pageId string, params ...interface{}) string {
	return app.App().PageMgr.PageUrl(pageId, params...)
}
