package com

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/scope"
)

const (
	CompInnerTagName = "w-inner"
)

var (
	ForbiddenAttrs = [...]string{
		"id",
		"class",
	}
)

type (
	componentView struct {
		ComponentView
		attrs    map[string]string
		template VNode
	}

	TemplateProvider interface {
		Template(container dom.Selection) dom.Selection
	}

	ComponentView struct {
		Name      string // The tag name for the component
		Prototype Prototype
		Template  TemplateProvider
	}

	ViewManager struct {
		compViews       map[string]*componentView
		sourceContainer dom.Selection
	}

	componentInstance struct {
		model    Prototype
		origNode VNode
		realNode VNode
	}

	Empty struct{}

	// ComponentPrototype is the common interface for all component's prototypes
	ComponentPrototype interface {
		// Init is called on each instantiation of a component.
		//
		// You can return a VNode, it will replace the actual node displayed,
		// the new inner contents of the returned VNode will also be used
		// to replace "<w-inner>" nodes instead of the original.
		Init(VNode) (VNode, error)

		// Update is called whenever the virtual DOM is rerendered
		Update(VNode) error
	}

	BaseProto struct{}

	// StringView satisfies the TemplateProvider interface for a plain string
	StringTemplate string

	// HTMLTemplate is a TemplateProvider that gets the template HTML code
	// from a <template> element with the given Id
	HTMLTemplate struct {
		Id string
	}
)

func (t HTMLTemplate) Template(container dom.Selection) dom.Selection {
	tpl := container.Find("template#" + t.Id)
	tpl.Remove()

	return tpl
}

func (t StringTemplate) Template(container dom.Selection) dom.Selection {
	node := container.NewFragment("<node></node>")
	node.Append(container.NewFragment(string(t)))
	return node
}

func (b BaseProto) Init(node VNode) (VNode, error) { return node, nil }
func (b BaseProto) Update(node VNode) error        { return nil }

func (c *componentInstance) Model() interface{} {
	return c.model
}

func (c *componentInstance) Update() error {
	return c.model.Update(c)
}

func dePtr(proto Prototype) reflect.Type {
	if proto == nil {
		return reflect.TypeOf(Empty{})
	}

	p := reflect.TypeOf(proto)
	if p.Kind() == reflect.Ptr {
		return p.Elem()
	}

	return p
}

func (cv *componentView) prepareAttributes(prototype reflect.Type) error {
	for i := 0; i < prototype.NumField(); i++ {
		field := prototype.Field(i)
		hname := field.Tag.Get("html")
		if hname == "" {
			hname = field.Name
		}
		cv.attrs[strings.ToLower(hname)] = field.Name
	}

	return nil
}

func (cv *componentView) HasAttr(attr string) (has bool, fieldName string) {
	fieldName, has = cv.attrs[attr]

	return
}

func (ci *componentInstance) PrepareInner(outerScope *scope.Scope, bindFn func(VNode, *scope.Scope)) (err error) {
	NodeWalk(&ci.realNode, func(parent *VNode, i int) {
		node := parent.Children[i]
		// replace <w-inner> elements with inner content
		if node.Type == ElementNoder && node.Data == CompInnerTagName {
			ci.origNode.Type = GhostNode
			ci.origNode.scope = outerScope
			parent.Children[i] = NodeClone(ci.origNode)
		}
	})

	ci.realNode.rerenderCb = func(node VNode) {
		ci.model.Update(ci.realNode)
	}
}

func (t *componentView) NewModel(node VNode) Prototype {
	if t.Prototype == nil {
		return nil
	}

	prototype := dePtr(t.Prototype)
	cptr := reflect.New(prototype)
	clone := cptr.Elem()
	if t.attrs != nil {
		for attr, fieldName := range t.attrs {
			if val, ok := node.Attrs[attr]; ok {
				field := clone.FieldByName(fieldName)
				if _, ok := field.Interface().(string); ok {
					field.Set(val)
					continue
				}

				n, err := fmt.Sscan(val, field.Interface())

				if n != 1 || err != nil {
					panic(fmt.Sprintf(`Cannot parse value "%v" to type "%v" for attribute "%v" of component "%v". Error: %v.`,
						val, field.Type().String(), attr, t.Name, err))
				}
			}
		}
	}

	return cptr.Interface().(Prototype)
}

func (t *componentView) NewInstance(node VNode, outerScope *scope.Scope) (inst *componentInstance, err error) {
	model := t.NewModel(node)
	realNode := NodeClone(t.template)
	newOrig := model.Init(node)
	return &componentInstance{
		model:    model,
		origNode: newOrig,
		realNode: realNode,
	}

	return
}

func NewManager(sourceContainer dom.Selection) *Manager {
	return &Manager{
		components:      make(map[string]*Component),
		sourceContainer: sourceContainer,
	}
}

func isForbiddenAttr(attr string) bool {
	lattr := strings.ToLower(attr)
	for _, a := range ForbiddenAttrs {
		if a == lattr {
			return true
		}
	}
	return false
}

func (tm *Manager) RegisterComponents(specs []ComponentView) (ret error) {
	for _, ht := range specs {
		ct := &Component{
			ComponentView: ht,
			attrs:         map[string]string{},
		}

		if ht.Template == nil {
			panic(fmt.Errorf("No template available for component %v", ht.Name))
		}

		ct.template = ht.Template.Template(tm.sourceContainer).ToVNode()
		ct.template.Data = ht.Name

		prototype := ct.Prototype
		if prototype != nil {
			p := reflect.ValueOf(prototype)

			if p.Kind() == reflect.Ptr {
				p = p.Elem()
			}

			if p.Kind() != reflect.Struct {
				return fmt.Errorf(`Prototype for "%v" has type "%v", invalid, it must be a struct or pointer to struct instead.`, ct.Name, p.Type().String())
			}
		}

		err := ct.prepareAttributes(dePtr(prototype))
		if err != nil {
			ret = err
			continue
		}

		tm.components[strings.ToLower(ct.Name)] = ct
	}

	return ret
}

// GetHtmlTag checks if the element's tag is of a registered component
func (tm *Manager) GetComponent(elem dom.Selection) (ct *Component, ok bool) {
	tagname, err := elem.TagName()
	if err != nil {
		ok = false
		return
	}

	ct, ok = tm.components[tagname]
	return
}
