package wade

import (
	"fmt"

	"github.com/gowade/wade/vdom"
)

var domDriver vdom.Driver

func SetDOMDriver(d vdom.Driver) {
	domDriver = d
}

func DOM() vdom.Driver {
	if domDriver == nil {
		panic("DOM driver not set.")
	}

	return domDriver
}

func Str(value interface{}) string {
	if s, ok := value.(string); ok {
		return s
	}

	return fmt.Sprint(value)
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

	VNode *vdom.Element

	InternalRefsHolder interface{} // please don't touch, this is for use by Fuel's generated code
	unmounted          bool
}

func (c *Com) InternalState() interface{} {
	return nil
}

func (c *Com) InternalComPtr() *Com {
	return c
}

func (c *Com) InternalUnmount() {
	c.unmounted = true
}

func (c *Com) InternalUnmounted() bool {
	return c.unmounted
}

type Component interface {
	Render(interface{}) *vdom.Element

	InternalState() interface{}
}

type RootComponent interface {
	Component
	InternalComPtr() *Com
}

func Render(com RootComponent, d vdom.DOMNode) {
	vnode := com.Render(nil)
	c := com.InternalComPtr()
	c.VNode = vnode

	vdom.PerformDiff(vnode, nil, d)
}
