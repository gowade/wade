package driver

import "github.com/gopherjs/gopherjs/js"

type Specification interface {
	Render() Element
}

type Class interface{}

type Element interface{}

type Driver interface {
	CreateClass(specification Specification) Class
	CreateElement(kind interface{}, props interface{}, children ...interface{}) Element
	Render(interface{}, *js.Object)
}
