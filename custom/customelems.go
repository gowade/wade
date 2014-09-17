package custom

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/phaikawl/wade/dom"
)

var (
	ForbiddenAttrs = [...]string{
		"id",
		"class",
	}
)

type (
	HtmlTag struct {
		Name       string
		Html       string
		Prototype  TagPrototype
		Attributes []string
	}

	TagManager struct {
		custags map[string]HtmlTag
	}

	CustomElem struct {
		model    TagPrototype
		Elem     dom.Selection
		Contents dom.Selection
	}

	BaseProto struct{}

	Empty struct{}

	Ctl interface {
		Dom() dom.Dom
	}

	ContentsCtl interface {
		Ctl
		Contents() dom.Selection //returns the contents container
	}

	contentsCtlImpl struct {
		contents dom.Selection
	}

	ElemCtl interface {
		Ctl
		Element() dom.Selection // returns the element itself
	}

	TagPrototype interface {
		ProcessContents(ContentsCtl) error
		Update(ElemCtl) error
		Init() error
	}
)

func (c contentsCtlImpl) Contents() dom.Selection {
	return c.contents
}

func (c contentsCtlImpl) Dom() dom.Dom {
	return c.contents
}

func (b BaseProto) Init() error                           { return nil }
func (b BaseProto) ProcessContents(ctl ContentsCtl) error { return nil }
func (b BaseProto) Update(ctl ElemCtl) error              { return nil }

func (c *CustomElem) Model() interface{} {
	return c.model
}

func (c *CustomElem) ContentsCtn() dom.Selection {
	return c.Contents
}

func (c *CustomElem) Element() dom.Selection {
	return c.Elem
}

func (c *CustomElem) Update() error {
	return c.model.Update(c)
}

func (c *CustomElem) Dom() dom.Dom {
	return c.Elem
}

func dePtr(proto TagPrototype) reflect.Type {
	if proto == nil {
		return reflect.TypeOf(Empty{})
	}

	p := reflect.TypeOf(proto)
	if p.Kind() == reflect.Ptr {
		return p.Elem()
	}

	return p
}

func (tag HtmlTag) prepareAttributes(prototype reflect.Type) error {
	for _, attr := range tag.Attributes {
		attr = strings.TrimSpace(attr)
		if isForbiddenAttr(attr) {
			return fmt.Errorf(`Unable to register custom tag "%v", use of `+
				`"%v" as a public attribute is forbidden because it conflicts `+
				`with HTML's %v attribute.`, tag.Name, attr, strings.ToLower(attr))
		}
		if _, ok := prototype.FieldByName(attr); !ok {
			return fmt.Errorf(`Attribute "%v" is not available in the model for custom tag "%v".`, attr, tag.Name)
		}
	}

	return nil
}

func (tag HtmlTag) HasAttr(attr string) (has bool, realName string) {
	for _, a := range tag.Attributes {
		if strings.ToLower(a) == attr {
			has = true
			realName = a
			return
		}
	}

	return
}

func (ce *CustomElem) PrepareContents(contentBindFn func(dom.Selection, bool)) (err error) {
	contents := ce.Contents.Contents()
	if contents.Length() > 0 {
		for i, wc := range ce.Elem.Find("wcontents").Elements() {
			c := contents
			if i > 0 {
				c = c.Clone()
			}

			wc.ReplaceWith(c)
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

func (t HtmlTag) NewModel(elem dom.Selection) TagPrototype {
	if t.Prototype == nil {
		return nil
	}

	prototype := dePtr(t.Prototype)
	cptr := reflect.New(prototype)
	clone := cptr.Elem()
	if t.Attributes != nil {
		for _, attr := range t.Attributes {
			if val, ok := elem.Attr(attr); ok {
				field := clone.FieldByName(attr)
				var err error = nil
				var v interface{}
				ftype, _ := prototype.FieldByName(attr)
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

	return cptr.Interface().(TagPrototype)
}

func (t HtmlTag) NewElem(elem dom.Selection) (ce *CustomElem, err error) {
	contentElem := elem.NewFragment("<wroot></wroot>")
	contentElem.SetHtml(elem.Html())
	elem.SetHtml(t.Html)
	model := t.NewModel(elem)
	err = model.Init()
	if err != nil {
		return
	}

	ce = &CustomElem{
		model:    model,
		Elem:     elem,
		Contents: contentElem,
	}

	return
}

func NewTagManager() *TagManager {
	return &TagManager{
		custags: make(map[string]HtmlTag),
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

func (tm *TagManager) RegisterTags(customTags []HtmlTag) (ret error) {
	for _, ct := range customTags {
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

		tm.custags[strings.ToLower(ct.Name)] = ct
	}

	return ret
}

// GetHtmlTag checks if the element's tag is of a registered custom tag
func (tm *TagManager) GetTag(elem dom.Selection) (ct HtmlTag, ok bool) {
	if elem.Length() > 1 {
		panic("You are getting a custom tag for multiple elements, it's surely an error.")
	}

	tagname, err := elem.TagName()
	if err != nil {
		ok = false
		return
	}

	ct, ok = tm.custags[tagname]
	return
}

func (tm *TagManager) RedefTag(tagname string, html string) (err error) {
	tag, ok := tm.custags[tagname]
	if !ok {
		err = fmt.Errorf(`Custom tag "%v" has not been registered.`, tagname)
		return
	}

	tag.Html = html
	tm.custags[tagname] = tag
	return
}
