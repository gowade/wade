package utils

import (
	"fmt"

	"github.com/phaikawl/wade/rt"
)

var Fmt = fmt.Sprintf

func Url(pageId string, params ...interface{}) string {
	return rt.App().PageMgr.PageUrl(pageId, params...)
}
