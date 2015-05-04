package wade

import (
	"fmt"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gowade/wade/vdom"
	"github.com/gowade/wade/vdom/browser"
)

func Str(value interface{}) string {
	if s, ok := value.(string); ok {
		return s
	}

	return fmt.Sprint(value)
}

func MakeDOMInputEl(jso *js.Object) DOMInputEl {
	return DOMInputEl{jso}
}

type DOMInputEl struct{ *js.Object }

func (e DOMInputEl) Value() string {
	return e.Get("value").String()
}

func (e DOMInputEl) SetValue(value string) {
	e.Set("value", value)
}

func MakeCom(name string, children []vdom.Node) Com {
	return Com{
		Name:     name,
		Children: children,
	}
}

type VNodeHolder struct{ *vdom.Element }

type Com struct {
	Name     string
	Children []vdom.Node

	VNode *VNodeHolder

	InternalRefsHolder interface{} // please don't touch, this is for use by Fuel's generated code
}

func (c Com) InternalState() interface{} {
	return nil
}

func (c *Com) InternalComPtr() *Com {
	return c
}

type Component interface {
	Render(interface{}) *vdom.Element

	InternalState() interface{}
}

type RootComponent interface {
	Component
	InternalComPtr() *Com
}

func PerformDiff(a, b *vdom.Element, domNode *js.Object) {
	vdom.PerformDiff(a, b, browser.DomNode{domNode}, browser.Adapter)
}

func Render(com RootComponent, elemId string) {
	domNode := js.Global.Get("document").Call("getElementById", elemId)
	vnode := com.Render(nil)
	browser.PerformDiff(vnode, nil, domNode)

	c := com.InternalComPtr()
	c.VNode = &VNodeHolder{vnode}
}
