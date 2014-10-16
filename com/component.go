package com

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/scope"
)

var (
	ForbiddenAttrs = [...]string{
		"id",
		"class",
	}
)

type (
	Component struct {
		Spec
		template string
		attrs    map[string]string
		manager  *Manager
	}

	TemplateProvider interface {
		Template(container dom.Selection) string
	}

	Spec struct {
		Name      string
		Prototype Prototype
		Template  TemplateProvider
	}

	Manager struct {
		components      map[string]*Component
		sourceContainer dom.Selection
	}

	Element struct {
		model    Prototype
		Elem     dom.Selection
		Contents dom.Selection
	}

	Empty struct{}

	Ctl interface {
		Dom() dom.Dom
	}

	ContentsData interface {
		Ctl
		Contents() dom.Selection //returns the contents container
	}

	contentsCtlImpl struct {
		contents dom.Selection
	}

	ElemData interface {
		Ctl
		Element() dom.Selection // returns the element itself
	}

	Prototype interface {
		Init(parentScope *scope.Scope) error
		ProcessContents(ContentsData) error
		Update(ElemData) error
	}

	ComponentIniter interface {
		ComponentInit(proto Prototype)
	}

	BaseProto struct{}

	StringTemplate   string
	DeclaredTemplate struct {
		Id string
	}
)

func (t DeclaredTemplate) Template(container dom.Selection) string {
	tpl := container.Find("template#" + t.Id)
	tpl.Remove()

	return tpl.Html()
}

func (t StringTemplate) Template(_ dom.Selection) string {
	return string(t)
}

func (b BaseProto) Init(parentScope *scope.Scope) error    { return nil }
func (b BaseProto) ProcessContents(ctl ContentsData) error { return nil }
func (b BaseProto) Update(ctl ElemData) error              { return nil }

func (c contentsCtlImpl) Contents() dom.Selection {
	return c.contents
}

func (c contentsCtlImpl) Dom() dom.Dom {
	return c.contents
}

func (c *Element) Model() interface{} {
	return c.model
}

func (c *Element) ContentsCtn() dom.Selection {
	return c.Contents
}

func (c *Element) Element() dom.Selection {
	return c.Elem
}

func (c *Element) Update() error {
	return c.model.Update(c)
}

func (c *Element) Dom() dom.Dom {
	return c.Elem
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

func (tag *Component) prepareAttributes(prototype reflect.Type) error {
	for i := 0; i < prototype.NumField(); i++ {
		field := prototype.Field(i)
		hname := field.Tag.Get("html")
		if hname == "" {
			hname = field.Name
		}
		tag.attrs[strings.ToLower(hname)] = field.Name
	}

	return nil
}

func (tag *Component) HasAttr(attr string) (has bool, fieldName string) {
	fieldName, has = tag.attrs[attr]

	return
}

func (ce *Element) PrepareContents(contentBindFn func(dom.Selection, bool)) (err error) {
	contents := ce.Contents.Contents()
	if contents.Length() > 0 {
		for i, wc := range ce.Elem.Find("wcontents").Elements() {
			c := contents
			if i > 0 {
				c = c.Clone()
			}

			wc.ReplaceWith(c)
			//gopherjs:blocking
			contentBindFn(c, false)

			err = ce.model.ProcessContents(contentsCtlImpl{c})
			if err != nil {
				return
			}
		}
	} else {
		ce.Elem.Find("wcontents").Remove()
	}

	err = ce.model.Update(ce)
	if err != nil {
		return
	}

	return
}

func (t *Component) NewModel(elem dom.Selection) Prototype {
	if t.Prototype == nil {
		return nil
	}

	prototype := dePtr(t.Prototype)
	cptr := reflect.New(prototype)
	clone := cptr.Elem()
	if t.attrs != nil {
		for attr, fieldName := range t.attrs {
			if val, ok := elem.Attr(attr); ok {
				field := clone.FieldByName(fieldName)
				var err error = nil
				var v interface{}
				ftype, _ := prototype.FieldByName(fieldName)
				kind := ftype.Type.Kind()
				switch kind {
				case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
					v, err = strconv.Atoi(val)
				case reflect.Uint, reflect.Uint16, reflect.Uint32:
					var m uint32
					var n uint64
					n, err = strconv.ParseUint(val, 10, 32)
					m = uint32(n)
					v = m
				case reflect.Float32:
					v, err = strconv.ParseFloat(val, 32)
				case reflect.Bool:
					v, err = strconv.ParseBool(val)
				case reflect.String:
					v = val
				default:
					err = fmt.Errorf(`Unhandled type "%v", cannot use html to set the attribute "%v" of custom tag "%v".
	consider using field binding syntax instead.`, kind, attr, t.Name)
				}

				if err != nil {
					panic(fmt.Sprintf(`Invalid value "%v" for attribute "%v" of custom tag "%v": type mismatch. Parse info: %v.`,
						val, attr, t.Name, err))
				}

				field.Set(reflect.ValueOf(v).Convert(field.Type()))
			}
		}
	}

	return cptr.Interface().(Prototype)
}

func (t *Component) NewElem(elem dom.Selection, initer ComponentIniter, parentScope *scope.Scope) (ce *Element, err error) {
	contentElem := elem.NewFragment("<wroot></wroot>")
	contentElem.SetHtml(elem.Html())
	elem.SetHtml(t.template)
	model := t.NewModel(elem)

	if initer != nil {
		initer.ComponentInit(model)
	}

	err = model.Init(parentScope)
	if err != nil {
		return
	}

	ce = &Element{
		model:    model,
		Elem:     elem,
		Contents: contentElem,
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

func (tm *Manager) RegisterComponents(specs []Spec) (ret error) {
	for _, ht := range specs {
		ct := &Component{
			Spec:    ht,
			attrs:   map[string]string{},
			manager: tm,
		}

		if ht.Template == nil {
			panic(fmt.Errorf("No template available for component %v", ht.Name))
		}

		ct.template = ht.Template.Template(tm.sourceContainer)

		prototype := ct.Prototype
		if prototype != nil {
			p := reflect.ValueOf(prototype)

			if p.Kind() == reflect.Ptr {
				p = p.Elem()
			}

			if p.Kind() != reflect.Struct {
				return fmt.Errorf(`Custom tag prototype for "%v" has type "%v", it must be a struct or pointer to struct instead.`, ct.Name, p.Type().String())
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
