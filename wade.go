package wade

import (
	"fmt"
	gourl "net/url"

	"github.com/gowade/wade/driver"
	"github.com/gowade/wade/utils/dom"
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
		ComponentName: name,
		Children:      children,
	}
}

type VNodeHolder struct{ *vdom.Element }

type Com struct {
	ComponentName string
	Children      []vdom.Node
	Attrs         vdom.Attributes

	VNode *vdom.Element

	InternalRefsHolder interface{} // please don't touch, this is for use by Fuel's generated code
	unmounted          bool
}

// OnRender is called every time the component's Render() method is called.
func (c *Com) OnInvoke() {}

// PrepareMount is called once before the component or its ancestor is rendered into the real DOM
func (c *Com) BeforeMount() {}

// PrepareMount is called once after the component or its ancestor is rendered into the real DOM
func (c *Com) AfterMount() {}

// DidUpdate is called when the component or its descendants have been updated in the real DOM
func (c *Com) OnUpdated() {}

// OnUnmount is called whenever the component or its ancestor is removed from the real DOM
func (c *Com) OnUnmount() {}

func (c *Com) InternalInitState(interface{}) {}

func (c *Com) InternalState() interface{} {
	return nil
}

func (c *Com) InternalComPtr() *Com {
	return c
}

func (c *Com) InternalUnmount() {
	c.OnUnmount()
	c.unmounted = true
}

func (c *Com) SetVNode(node *vdom.Element) {
	c.VNode = node
}

func (c *Com) InternalUnmounted() bool {
	return c.unmounted
}

type Component interface {
	Render(interface{}) *vdom.Element

	InternalState() interface{}
	InternalComPtr() *Com
	InternalInitState(interface{})

	OnInvoke()
	BeforeMount()
	AfterMount()
	OnUpdated()
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

func MergeMaps(m1, m2 map[string]interface{}) map[string]interface{} {
	if m1 == nil && m2 == nil {
		return nil
	}

	m := make(map[string]interface{})
	if m1 != nil {
		for k, v := range m1 {
			m[k] = v
		}
	}

	if m2 != nil {
		for k, v := range m2 {
			m[k] = v
		}
	}

	return m
}

func If(cond bool, v string) string {
	if cond {
		return v
	}

	return ""
}

func WrapEvt(handler func(dom.Event)) interface{} {
	return dom.NewEventHandler(handler)
}

func QueryEscape(str string) string {
	return gourl.QueryEscape(str)
}
