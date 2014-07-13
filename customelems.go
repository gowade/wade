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

func (t *CustomTag) TagContents(elem jq.JQuery) {
	elem.SetHtml(t.elem.Html())
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

// RegisterNew registers a new custom element tag.
// It selects the <welement> with id #elemid and registers a new tag with the given name
// The content and specifications of the tag is taken from #elemid.
// For example, if
//	wade.RegisterNew("errorlist", "t-errorlist", prototype)
// is called, the element welement#t-errorlist, like
//	<welement id="t-errorlist" attributes="Errors Subject">
//		<p>errors for <% Subject %></p>
//		<ul>
//			<li bind-each="Errors -> _, msg"><% msg %></li>
//		</ul>
//	</welement>
// will be selected and its contents
// will be used as the new tag <errorlist>'s contents.
// This new tag may be used like this
//	<errorlist attr-subject="Username" bind="Errors: Username.Errors"></errorlist>
// And if "Username.Errors" is {"Invalid.", "Not enough chars."}, something like this will
// be put in place of the above:
//	<p>errors for Username</p>
//	<ul>
//		<li>Invalid.</li>
//		<li>Not enough chars.</li>
//	</ul>
// The prototype parameter must not be a pointer, it is actually used like a type,
// It will be cloned, real instances of it will be created for each
// separate custom element.
func (tm *CustagMan) RegisterNew(tagName string, elemid string, prototype interface{}) {
	tagElem := tm.tcontainer.Find("#" + elemid)
	if tagElem.Length == 0 {
		panic(fmt.Sprintf("Welement with id #%v does not exist.", elemid))
	}
	if !tagElem.Is("welement") {
		panic(fmt.Sprintf("The element #%v to register new tag must be a welement.", elemid))
	}

	ptype := reflect.TypeOf(prototype)
	if ptype.Kind() != reflect.Struct {
		panic(fmt.Sprintf(`Wrong type for prototype of tag "%v", it must be a struct (non-pointer).`, tagName))
	}

	custag := &CustomTag{tagName, tagElem, prototype, nil}
	custag.prepareAttributes(ptype)
	tm.custags[strings.ToUpper(tagName)] = custag
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
