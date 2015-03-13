package react

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/hanleym/wade"
	"github.com/hanleym/wade/driver"
)

type React struct {
	*js.Object
}

type Class *js.Object

type Element *js.Object

func (react *React) CreateClass(specification driver.Specification) driver.Class {
	return react.Call("createClass", js.M{
		"render": specification.Render(),
	})
}

func (react *React) CreateElement(kind interface{}, props interface{}, children ...interface{}) driver.Element {
	params := make([]interface{}, len(children)+2)
	for i := range children {
		params[i+2] = children[i]
	}

	switch kind := kind.(type) {
	case driver.Class:
		params[0] = kind
		params[1] = props
	case string:
		params[0] = kind
		params[1] = props
	default:
		panic("BAD KIND")
	}

	element := driver.Element(react.Call("createElement", params...))

	return &element
}

// func (react *React) RenderToHTML(element Element) string {
// 	val := react.Call("renderToString", element)
// 	if val == nil {
// 		return ""
// 	}
// 	return val.String()
// }

// func (react *React) RenderToStaticMarkup(element Element) string {
// 	val := react.Call("renderToStaticMarkup", element)
// 	if val == nil {
// 		return ""
// 	}
// 	return val.String()
// }

func (react *React) Render(element interface{}, node *js.Object) {
	react.Call("render", element, node)
}

func init() {
	wade.Default = &React{js.Global.Get("React")}
}
