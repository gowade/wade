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

	vdomProvider interface {
		ToVNode(*VNode, templateConverter)
	}

	ComponentView struct {
		Name      string // The tag name for the component
		Prototype ComponentPrototype
		Template  vdomProvider
	}

	ComManager struct {
		templateConv templateConverter
		compViews    map[string]*componentView
	}

	templateConverter interface {
		FromString(html string) VNode
		FromHTMLTemplate(dst *VNode, templateId string) VNode
	}

	componentInstance struct {
		model    ComponentPrototype
		origNode VNode
		realNode *VNode
	}

	empty struct{}

	// ComponentPrototype is the common interface for all component's prototypes
	ComponentPrototype interface {
		// ProcessInner is called before instantiation of a component to process
		// the component's inner content (passed  in as a VNode).
		//
		// You can modify the node's tree, the modified inner content of the node
		// will be used to replace "<w-inner>" nodes instead of the original.
		//
		// Inner content is *moved* to the real node for the first <w-inner>,
		// cloned for other <w-inner>.
		ProcessInner(inner VNode)

		// Inner is called on instantiation of a component to perform initializations
		Init(realNode VNode)

		// Update is called whenever the virtual DOM is rerendered
		Update(realNode VNode)
	}

	BaseProto struct{}
)

type (
	StringTemplate struct {
		HTML string
	}

	HTMLTemplate struct {
		TemplateId string
	}
)

func (t VNode) ToVNode(template *VNode, conv templateConverter) {
	*template = t
}

func (t StringTemplate) ToVNode(template *VNode, conv templateConverter) {
	*template = conv.FromString(t.HTML)
	return
}

func (t HTMLTemplate) ToVNode(template *VNode, conv templateConverter) {
	*template = conv.FromHTMLTemplate(template, t.TemplateId)
}

func (b BaseProto) ProcessInner(node VNode) {}
func (b BaseProto) Init(node VNode)         {}
func (b BaseProto) Update(node VNode)       {}

func (c *componentInstance) Model() interface{} {
	return c.model
}

func dePtr(proto ComponentPrototype) reflect.Type {
	if proto == nil {
		return reflect.TypeOf(empty{})
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
	orig.Data = "group"
	orig.Type = GroupNode
	orig.Attrs = map[string]interface{}{}

	inst = &componentInstance{
		model:    model,
		origNode: orig,
		realNode: node,
	}

	return
}

func (ci *componentInstance) prepareInner(outerScope scope.Scope) {
	ci.model.ProcessInner(ci.origNode)

	i := 0
	NodeWalk(ci.realNode, func(node *VNode) {
		// replace <w-inner> elements with inner content
		if node.Type == ElementNode && node.Data == CompInner {
			ci.origNode.scope = &outerScope
			if i > 0 {
				*node = ci.origNode.Clone()
			} else {
				*node = ci.origNode
				i++
			}
		}
	})

	ci.model.Init(*ci.realNode)
	ci.realNode.addCallback(func() (err error) {
		ci.model.Update(*ci.realNode)
		return
	})
}

func NewComManager(tempConv templateConverter) *ComManager {
	return &ComManager{
		templateConv: tempConv,
		compViews:    make(map[string]*componentView),
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

		ht.Template.ToVNode(&ct.template, tm.templateConv)
		if ct.template.Type != GroupNode {
			ct.template = VNode{
				Type:     GroupNode,
				Data:     "component",
				Children: []VNode{ct.template},
			}
		}

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
