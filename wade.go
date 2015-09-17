package wade

import (
	"fmt"
	gourl "net/url"

	"github.com/gowade/vdom"
	"github.com/gowade/wade/dom"
	"github.com/gowade/wade/driver"
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

type Component interface{}

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

func NewVNodeList(nodes ...interface{}) []vdom.VNode {
	var l []vdom.VNode
	for _, n := range nodes {
		if n == nil {
			continue
		}

		switch n := n.(type) {
		case []vdom.VNode:
			l = append(l, n...)
		case *vdom.VElement, vdom.VText:
			l = append(l, n.(vdom.VNode))
		default:
			panic("Invalid node type")
		}
	}

	return l
}
