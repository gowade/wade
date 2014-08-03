package wade

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	jq "github.com/gopherjs/jquery"

	"github.com/phaikawl/wade/bind"
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

type CustomTag struct {
	name        string
	elem        jq.JQuery
	prototype   interface{}
	publicAttrs []string
}

func (tag *CustomTag) prepareAttributes(prototype reflect.Type) {
	tagElem := tag.elem
	publicAttrs := make([]string, 0)
	if attrs := tagElem.Attr("attributes"); attrs != "" {
		publicAttrs = strings.Split(attrs, " ")
		for _, attr := range publicAttrs {
			attr = strings.TrimSpace(attr)
			if isForbiddenAttr(attr) {
				panic(fmt.Sprintf(`Unable to register custom tag "%v", use of `+
					`"%v" as a public attribute is forbidden because it conflicts `+
					`with HTML's %v attribute.`, tag.name, attr, strings.ToLower(attr)))
			}
			if _, ok := prototype.FieldByName(attr); !ok {
				panic(fmt.Sprintf(`Attribute "%v" is not available in the model for custom tag "%v".`, attr, tag.name))
			}
		}
	}

	tag.publicAttrs = publicAttrs
}

func (t *CustomTag) TagContents(elem jq.JQuery, model interface{}) {
	contentElem := elem.Clone()
	elem.SetHtml(t.elem.Html())
	ce := &CustomElem{elem, contentElem}
	if im, ok := model.(CustomElemInit); ok {
		im.Init(ce)
	}
	elem.Find("wcontent").ReplaceWith(contentElem.Html())
}

func (t *CustomTag) NewModel(elem jq.JQuery) interface{} {
	if t.publicAttrs == nil {
		panic("Something is wrong, publicAttrs unset.")
	}

	prototype := reflect.TypeOf(t.prototype)
	cptr := reflect.New(prototype)
	clone := cptr.Elem()
	for _, attr := range t.publicAttrs {
		if val := elem.Attr(attr); val != "" {
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

type CustagMan struct {
	custags    map[string]*CustomTag
	tcontainer jq.JQuery
}

func newCustagMan(tcontainer jq.JQuery) *CustagMan {
	return &CustagMan{
		custags:    make(map[string]*CustomTag),
		tcontainer: tcontainer,
	}
}

type CustomElem struct {
	Elem    jq.JQuery
	Content jq.JQuery
}

type CustomElemInit interface {
	Init(*CustomElem)
}

func (tm *CustagMan) registerTags(tagElems []jq.JQuery, protoMap map[string]interface{}) error {
	for _, elem := range tagElems {
		tagname := elem.Attr("tagname")
		if tagname == "" {
			return fmt.Errorf("No tag name specified for the element with content:\n `%v`", elem.Get(0).Get("outerHTML").Str())
		}

		if prototype, ok := protoMap[tagname]; ok {
			p := reflect.ValueOf(prototype)
			if p.Kind() == reflect.Ptr {
				p = p.Elem()
			}

			if p.Kind() != reflect.Struct {
				return fmt.Errorf(`Custom tag prototype for "%v", type "%v" is not a struct or pointer to struct.`, tagname, p.Type().String())
			}

			custag := &CustomTag{tagname, elem, p.Interface(), nil}
			custag.prepareAttributes(p.Type())
			tm.custags[strings.ToUpper(tagname)] = custag
		} else {
			return fmt.Errorf(`No prototype is specified for the custom element tag "%v", there must be one.`, tagname)
		}
	}

	return nil
}

// GetCustomTag checks if the element's tag is of a registered custom tag
func (tm *CustagMan) GetCustomTag(elem jq.JQuery) (ct bind.CustomTag, ok bool) {
	ct, ok = tm.custags[strings.ToUpper(elem.Prop("tagName").(string))]
	return
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
