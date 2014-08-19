package icommon

import (
	"github.com/phaikawl/wade/dom"
)

const WrapperTag = "ww"

func WrapperUnwrap(elem dom.Selection) {
	elem.Find(WrapperTag).Unwrap()
}

func IsWrapperElem(elem dom.Selection) bool {
	return elem.Is(WrapperTag)
}
