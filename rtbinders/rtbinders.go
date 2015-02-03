package rtbinders

import (
	"fmt"
	"reflect"

	"github.com/phaikawl/wade/core"
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/rt"
	"github.com/phaikawl/wade/utils"
)

type rtBinder struct {
	udtFn  core.BindFunc
	onceFn core.BindFunc
}

func newBinder(udtFunc core.BindFunc, onceFn ...core.BindFunc) (b rtBinder) {
	b.udtFn = udtFunc
	if len(onceFn) > 0 {
		b.onceFn = onceFn[0]
	}

	return
}

func RTBinder(b rtBinder) core.BindFunc {
	once := false
	return func(node *core.VNode) {
		if !once && b.onceFn != nil {
			b.onceFn(node)
			once = true
		}
		b.udtFn(node)
	}
}

func RTBinder_value(vFn func() interface{}, args []string) rtBinder {
	getValue := func() (rv reflect.Value) {
		rv = reflect.ValueOf(vFn())
		if rv.Kind() != reflect.Ptr {
			panic(`Value given to the "value" binder must be a pointer.`)
		}

		return rv
	}

	return newBinder(func(n *core.VNode) {
		n.SetAttr("value", utils.ToString(getValue().Elem().Interface()))
	}, func(n *core.VNode) {
		n.Attrs["onchange"] = func(evt dom.Event) {
			fmt.Sscan(evt.Target().Val(), getValue().Interface())
			go rt.App().Render()
		}
	})
}

var nodeOrigType = map[*core.VNode]core.NodeType{}

func RTBinder_if(vFn func() interface{}, args []string) rtBinder {
	return newBinder(func(n *core.VNode) {
		b := vFn().(bool)

		if b {
			if origType, ok := nodeOrigType[n]; ok {
				n.Type = origType
			}

		} else {
			if _, ok := nodeOrigType[n]; !ok {
				nodeOrigType[n] = n.Type
			}
			n.Type = core.DeadNode
		}
	})
}

func RTBinder_class(vFn func() interface{}, args []string) rtBinder {
	class := args[0]
	return newBinder(func(n *core.VNode) {
		n.SetClass(class, vFn().(bool))
	})
}
