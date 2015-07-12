package wade

import (
	"fmt"

	"github.com/gowade/wade/driver"
	"github.com/gowade/wade/vdom"
)

type M map[string]interface{}

type AppMode int

const (
	DevelopmentMode AppMode = iota
	ProductionMode
)

var (
	mode AppMode = DevelopmentMode
)

func ClientSide() bool {
	return driver.Env() == driver.BrowserEnv
}

func DevMode() bool {
	return mode == DevelopmentMode
}

func SetMode(appMode AppMode) {
	mode = appMode
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

// OnRender is called every time the component's Render() method is called.
// Could be used for initialization of field values, please don't call anything
// costly in this method.
func (c *Com) OnInvoke() {}

// OnMount is called whenever the component or its ancestor is rendered into the real DOM
func (c *Com) OnMount() {}

// OnUnmount is called whenever the component or its ancestor is removed from the real DOM
func (c *Com) OnUnmount() {}

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
	InternalComPtr() *Com

	OnInvoke()
	OnMount()
	OnUnmount()
}

func Render(com Component) vdom.Node {
	vnode := com.Render(nil)
	c := com.InternalComPtr()
	c.VNode = vnode

	return vnode
}

func VdomDrv() vdom.Driver {
	return driver.Vdom()
}
