package wade

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/phaikawl/wade/bind"
	"github.com/phaikawl/wade/dom"
)

const (
	ModelIdAttr = "data-wade-modelid"
)

var (
	ForbiddenAttrs = [...]string{
		"id",
		"class",
		"style",
		"title",
	}
)

type (
	CustomTag struct {
		name        string
		template    dom.Selection
		prototype   CustomElemProto
		publicAttrs []string
	}

	custagMan struct {
		custags map[string]*CustomTag
	}

	CustomElem struct {
		Dom      dom.Dom
		Template dom.Selection
		Contents dom.Selection
	}

	NoInit struct{}

	Empty struct{}

	CustomElemProto interface {
		Init(CustomElem) error
	}
)

func (ni NoInit) Init(ce CustomElem) error { return nil }

func dePtr(proto CustomElemProto) reflect.Type {
	if proto == nil {
		return reflect.TypeOf(Empty{})
	}

	p := reflect.TypeOf(proto)
	if p.Kind() == reflect.Ptr {
		return p.Elem()
	}

	return p
}

func (tag *CustomTag) prepareAttributes(prototype reflect.Type) error {
	publicAttrs := make([]string, 0)
	if attrs, ok := tag.template.Attr("attributes"); ok {
		publicAttrs = strings.Split(strings.TrimSpace(attrs), " ")
		for _, attr := range publicAttrs {
			attr = strings.TrimSpace(attr)
			if attr == "" {
				continue
			}
			if isForbiddenAttr(attr) {
				return fmt.Errorf(`Unable to register custom tag "%v", use of `+
					`"%v" as a public attribute is forbidden because it conflicts `+
					`with HTML's %v attribute.`, tag.name, attr, strings.ToLower(attr))
			}
			if _, ok := prototype.FieldByName(attr); !ok {
				return fmt.Errorf(`Attribute "%v" is not available in the model for custom tag "%v".`, attr, tag.name)
			}
		}
	}

	tag.publicAttrs = publicAttrs
	return nil
}

func (t *CustomTag) PrepareTagContents(elem dom.Selection, model interface{}, contentBindFn func(dom.Selection)) error {
	contentElem := elem.Clone()
	elem.SetHtml(t.template.Html())
	ce := CustomElem{
		Dom:      elem,
		Template: elem,
		Contents: contentElem,
	}
	err := model.(CustomElemProto).Init(ce)

	if err != nil {
		return err
	}

	contents := ce.Contents.Contents()
	if contents.Length() > 0 {
		elem.Find("wcontents").ReplaceWith(contents)
		contentBindFn(contents)
	} else {
		elem.Find("wcontents").Remove()
	}
	return nil
}

func (t *CustomTag) NewModel(elem dom.Selection) interface{} {
	if t.publicAttrs == nil {
		panic(fmt.Errorf("Something is wrong for %v, publicAttrs is not set.", t.name))
	}

	if t.prototype == nil {
		return nil
	}

	prototype := dePtr(t.prototype)
	cptr := reflect.New(prototype)
	clone := cptr.Elem()
	for _, attr := range t.publicAttrs {
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
				err = fmt.Errorf(`Unhandled type "%v", cannot use normal html to set the attribute "%v" of custom tag "%v".
consider using attribute binding instead.`, kind, attr, t.name)
			}

			if err != nil {
				panic(fmt.Sprintf(`Invalid value "%v" for attribute "%v" of custom tag "%v": type mismatch. Parse info: %v.`,
					val, attr, t.name, err))
			}

			field.Set(reflect.ValueOf(v).Convert(field.Type()))
		}
	}

	return cptr.Interface()
}

func newCustagMan() *custagMan {
	return &custagMan{
		custags: make(map[string]*CustomTag),
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

func (tm *custagMan) registerTags(tagElems []dom.Selection, protoMap map[string]CustomElemProto) (ret error) {
	for _, elem := range tagElems {
		tagname, ok := elem.Attr("tagname")
		if !ok {
			return fmt.Errorf(`No tag name specified (tagname="") for the element to register tag %v.`, tagname)
		}

		prototype, _ := protoMap[tagname]
		if prototype != nil {
			p := reflect.ValueOf(prototype)

			if p.Kind() == reflect.Ptr {
				p = p.Elem()
			}

			if p.Kind() != reflect.Struct {
				return fmt.Errorf(`Custom tag prototype for "%v" has type "%v", it must be a struct or pointer to struct instead.`, tagname, p.Type().String())
			}
		}

		custag := &CustomTag{tagname, elem, prototype, nil}
		err := custag.prepareAttributes(dePtr(prototype))
		if err != nil {
			ret = err
			continue
		}
		tm.custags[strings.ToLower(tagname)] = custag
	}

	return ret
}

// GetCustomTag checks if the element's tag is of a registered custom tag
func (tm *custagMan) GetCustomTag(elem dom.Selection) (ct bind.CustomTag, ok bool) {
	if elem.Length() > 1 {
		panic("You are getting a custom tag for multiple elements, it's likely an error.")
	}
	tagname, err := elem.TagName()
	if err != nil {
		return nil, false
	}
	ct, ok = tm.custags[tagname]
	return
}
