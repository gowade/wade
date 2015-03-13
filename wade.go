package wade

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/hanleym/wade/driver"
)

var Default driver.Driver

func CreateClass(specification driver.Specification) driver.Class {
	return Default.CreateClass(specification)
}

func CreateElement(kind interface{}, props interface{}, children ...interface{}) driver.Element {
	return Default.CreateElement(kind, props, children...)
}

func Render(element interface{}, object *js.Object) {
	Default.Render(element, object)
}
