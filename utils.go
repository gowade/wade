package wade

import (
	"unicode"

	"github.com/phaikawl/wade/icommon"
)

var (
	IsWrapperElem = icommon.IsWrapperElem
)

func camelize(src string) string {
	res := []rune{}
	startW := true
	for _, c := range src {
		if c == '-' {
			startW = true
			continue
		}
		ch := c
		if startW {
			ch = unicode.ToUpper(c)
			startW = false
		}
		res = append(res, ch)
	}
	return string(res)
}

type UrlInfo struct {
	path    string
	fullUrl string
}
