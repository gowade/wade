package icommon

import (
	"unicode"

	"github.com/phaikawl/wade/dom"
)

const WrapperTag = "ww"

func WrapperUnwrap(elem dom.Selection) {
	for _, e := range elem.Find(WrapperTag).Elements() {
		e.Unwrap()
	}
}

func IsWrapperElem(elem dom.Selection) bool {
	return elem.Is(WrapperTag)
}

func RemoveAllSpaces(src string) string {
	r := ""
	for _, c := range src {
		if !unicode.IsSpace(c) {
			r += string(c)
		}
	}

	return r
}
