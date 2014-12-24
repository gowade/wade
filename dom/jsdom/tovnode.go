package jsdom

import (
	"reflect"
	"strings"

	"github.com/gopherjs/gopherjs/js"

	"github.com/phaikawl/wade/core"
	"github.com/phaikawl/wade/dom"
	_ "github.com/phaikawl/wade/dom/jsdom/shim"
)

var (
	gMithril = mithril{js.Global.Get("Mithril")}
)

type (
	mithril struct {
		js.Object
	}
)

//type htmlNode struct{ js.Object }

//func (node htmlNode) ToVNode() (result []core.VNode) {
//	ni := node.Interface()
//	switch v := ni.(type) {
//	case string:
//		return dom.ParseMustaches(v)
//	}

//	attrs := map[string]interface{}{}
//	binds := []core.Bindage{}

//	if ao := node.Get("attrs"); !ao.IsUndefined() && !ao.IsNull() {
//		for attr, value := range ao.Interface().(map[string]interface{}) {
//			var bindType core.BindType
//			switch attr[0] {
//			case core.AttrBindPrefix:
//				bindType = core.AttrBind
//			case core.BinderBindPrefix:
//				bindType = core.BinderBind
//			default:
//				attrs[attr] = value
//				continue
//			}

//			binds = append(binds, core.Bindage{
//				Type: bindType,
//				Name: attr[1:],
//				Expr: value.(string),
//			})
//		}
//	}

//	n := core.VNode{
//		Data:     node.Get("tag").Str(),
//		Type:     core.ElementNode,
//		Attrs:    attrs,
//		Binds:    binds,
//		Children: []core.VNode{},
//	}

//	if _, isGrp := attrs[core.GroupAttrName]; isGrp { // has "!group" attribute
//		n.Type = core.GroupNode
//	}

//	if ca := node.Get("children"); !ca.IsUndefined() && !ca.IsNull() {
//		for i := 0; i < ca.Length(); i++ {
//			c := ca.Index(i)
//			n.Children = append(n.Children, htmlNode{c}.ToVNode()...)
//		}
//	}

//	return []core.VNode{core.VPrep(n)}
//}

func html2VNode(node js.Object) (result []core.VNode) {
	switch node.Get("nodeType").Int() {
	case 3:
		return dom.ParseMustaches(node.Get("nodeValue").Str())
	case 8:
		return []core.VNode{
			core.VPrep(core.VNode{Type: core.DataNode, Data: node.Get("nodeValue").Str()}),
		}
	case 1:
		attrs := map[string]interface{}{}
		binds := []core.Bindage{}
		if node.Get("hasAttributes").Bool() {
			jsAttrs := node.Get("attributes")
			for i := 0; i < jsAttrs.Length(); i++ {
				attr := jsAttrs.Index(i)
				key := attr.Get("name").Str()
				value := attr.Get("value").Str()

				var bindType core.BindType
				switch key[0] {
				case core.AttrBindPrefix:
					bindType = core.AttrBind
				case core.BinderBindPrefix:
					bindType = core.BinderBind
				default:
					attrs[key] = value
					continue
				}

				binds = append(binds, core.Bindage{
					Type: bindType,
					Name: key[1:],
					Expr: value,
				})
			}
		}

		tagName := strings.ToLower(node.Get("tagName").Str())
		if tagName == "template" {
			content := node.Get("content")
			if !content.IsUndefined() && !content.IsNull() {
				node = content
			}
		}

		n := core.VNode{
			Data:     tagName,
			Type:     core.ElementNode,
			Attrs:    attrs,
			Binds:    binds,
			Children: []core.VNode{},
		}

		if node.Get("hasChildNodes").Bool() {
			children := node.Get("childNodes")
			for i := 0; i < children.Length(); i++ {
				c := children.Index(i)
				n.Children = append(n.Children, html2VNode(c)...)
			}
		}

		return []core.VNode{core.VPrep(n)}
	}

	return []core.VNode{}
}

type Renderer struct {
	target js.Object
	vnode  *core.VNode
}

func (r Renderer) NewElementNode(vnode *core.VNode, children []dom.PlatformNode) dom.PlatformNode {
	attrs := make(map[string]interface{})
	for attr, value := range vnode.Attrs {
		if evtHandler, ok := value.(func(dom.Event)); ok {
			attrs[attr] = func(evt js.Object) {
				evtHandler(createEvent(evt))
				go func() {
					js.Global.Get("console").Call("profile")
					r.vnode.Update()
					js.Global.Get("console").Call("profileEnd")
					gMithril.Call("render", r.target, dom.Render(r.vnode, r).(js.Object))
				}()
			}

			continue
		}

		typ := reflect.TypeOf(value).Kind()
		if typ != reflect.Struct && typ != reflect.Ptr {
			attrs[attr] = value
		}
	}

	class := vnode.ClassStr()
	if class != "" {
		attrs["class"] = class
	}

	nChildren := make([]interface{}, len(children))
	for i := range children {
		nChildren[i] = children[i]
	}

	return gMithril.Invoke(vnode.Data, attrs, nChildren)
}

func (r Renderer) NewTextNode(vnode *core.VNode) dom.PlatformNode {
	return vnode.Data
}

func Render(target js.Object, vnode *core.VNode) {
	gMithril.Call("render", target, dom.Render(vnode, Renderer{target, vnode}).(js.Object))
}

func ToVNode(html js.Object) core.VNode {
	return html2VNode(html)[0]
}
