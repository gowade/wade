package core

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/phaikawl/wade/scope"
)

const (
	CompInner = "w-inner"
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

	ComponentView struct {
		Name      string // The tag name for the component
		Prototype ComponentPrototype
		Template  VNode
	}

	ComManager struct {
		compViews map[string]*componentView
	}

	componentInstance struct {
		model    ComponentPrototype
		origNode VNode
		realNode VNode
	}

	Empty struct{}

	// ComponentPrototype is the common interface for all component's prototypes
	ComponentPrototype interface {
		// Init is called on each instantiation of a component.
		//
		// You can modify the node's tree, the modified inner contents of the node
		// will be used to replace "<w-inner>" nodes instead of the original.
		Init(VNode) error

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

func (b BaseProto) Init(node VNode) error   { return nil }
func (b BaseProto) Update(node VNode) error { return nil }

func (c *componentInstance) Model() interface{} {
	return c.model
}

func dePtr(proto ComponentPrototype) reflect.Type {
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
	fieldName, has = cv.attrs[strings.ToLower(attr)]

	return
}

func (t *componentView) NewModel(node *VNode) (ComponentPrototype, error) {
	if t.Prototype == nil {
		return nil, nil
	}

	prototype := dePtr(t.Prototype)
	cptr := reflect.New(prototype)
	clone := cptr.Elem()
	if t.attrs != nil {
		for attr, fieldName := range t.attrs {
			if val, ok := node.Attr(attr); ok {
				field := clone.FieldByName(fieldName)
				if strVal, ok := val.(string); ok {
					if _, ok = field.Interface().(string); ok {
						field.Set(reflect.ValueOf(strVal))
						continue
					}

					n, err := fmt.Sscan(strVal, field.Addr().Interface())

					if n != 1 || err != nil {
						return nil, fmt.Errorf(`Cannot parse value "%v" to type "%v" for attribute "%v" of component "%v". Error: %v.`,
							val, field.Type().String(), attr, t.Name, err)
					}
				} else {
					if reflect.TypeOf(val).AssignableTo(field.Type()) {
						field.Set(reflect.ValueOf(val))
					} else {
						return nil, fmt.Errorf(`Incompatible type in prototype field assignment.`)
					}
				}
			}
		}
	}

	return cptr.Interface().(ComponentPrototype), nil
}

func (t *componentView) NewInstance(node *VNode) (inst *componentInstance, err error) {
	model, err := t.NewModel(node)
	if err != nil {
		return
	}

	orig := *node
	node.Children = t.template.Clone().Children
	err = model.Init(orig)
	orig.Data = "div"
	orig.Type = GroupNode

	inst = &componentInstance{
		model:    model,
		origNode: orig,
		realNode: *node,
	}

	return
}

func (ci *componentInstance) prepareInner(outerScope scope.Scope) {
	NodeWalk(&ci.realNode, func(node *VNode) {
		// replace <w-inner> elements with inner content
		if node.Type == ElementNode && node.Data == CompInner {
			ci.origNode.scope = &outerScope
			*node = ci.origNode.Clone()
		}
	})

	ci.realNode.addCallback(func() (err error) {
		err = ci.model.Update(ci.realNode)
		return
	})
}

func NewComManager() *ComManager {
	return &ComManager{
		compViews: make(map[string]*componentView),
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

func (tm *ComManager) Register(specs ...ComponentView) (ret error) {
	for _, ht := range specs {
		ct := &componentView{
			ComponentView: ht,
			attrs:         map[string]string{},
		}

		ct.template = ht.Template
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

		tm.compViews[strings.ToLower(ct.Name)] = ct
	}

	return ret
}

// GetHtmlTag checks if the element's tag is of a registered component
func (tm *ComManager) GetComponent(tagName string) (ct *componentView, ok bool) {
	ct, ok = tm.compViews[tagName]
	return
}
