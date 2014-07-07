package wade

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	jq "github.com/gopherjs/jquery"
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
	name  string
	elem  jq.JQuery
	model interface{}
}

type CustagMan struct {
	elemModels []interface{}
	custags    map[string]*CustomTag
	tcontainer jq.JQuery
}

func newCustagMan(tcontainer jq.JQuery) *CustagMan {
	return &CustagMan{
		custags:    make(map[string]*CustomTag),
		elemModels: make([]interface{}, 0),
		tcontainer: tcontainer,
	}
}

func (tm *CustagMan) ModelForElem(elem jq.JQuery) interface{} {
	mi := elem.Attr(ModelIdAttr)
	if mi == "" {
		panic("no modelId assigned for the element, something's wrong?")
	}
	modelId, err := strconv.Atoi(mi)
	if err != nil {
		panic("wrong format for internal element attribute " + ModelIdAttr + ", something's wrong?")
	}
	return tm.elemModels[modelId]
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
	tm.custags[strings.ToUpper(tagName)] = &CustomTag{tagName, tagElem, prototype}
}

// IsCustomElem checks if the element's tag is of a registered custom tag
func (tm *CustagMan) IsCustomElem(elem jq.JQuery) bool {
	_, ok := tm.custags[strings.ToUpper(elem.Prop("tagName").(string))]
	return ok
}

func (tm *CustagMan) prepare() {
	for _, tag := range tm.custags {
		mtype := reflect.TypeOf(tag.model)
		if mtype.Kind() != reflect.Struct {
			panic(fmt.Sprintf("Wrong type for the model of tag #%v, it must be a struct (non-pointer).", tag.name))
		}
		tm.prepareTag(tag, mtype)
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

func (tm *CustagMan) prepareTag(tag *CustomTag, model reflect.Type) {
	tagElem := tag.elem
	publicAttrs := []string{}
	if attrs := tagElem.Attr("attributes"); attrs != "" {
		publicAttrs = strings.Split(attrs, " ")
		for _, attr := range publicAttrs {
			attr = strings.TrimSpace(attr)
			if isForbiddenAttr(attr) {
				panic(fmt.Sprintf(`Unable to register custom tag "%v", use of `+
					`"%v" as a public attribute is forbidden because it conflicts `+
					`with HTML's %v attribute.`, tag.name, attr, strings.ToLower(attr)))
			}
			if _, ok := model.FieldByName(attr); !ok {
				panic(fmt.Sprintf(`Attribute "%v" is not available in the model for custom tag "%v".`, attr, tag.name))
			}
		}
	}

	elems := tm.tcontainer.Find(tag.name)
	elems.Each(func(idx int, elem jq.JQuery) {
		cptr := reflect.New(model)
		clone := cptr.Elem()
		for _, attr := range publicAttrs {
			if val := elem.Attr(attr); val != "" {
				field := clone.FieldByName(attr)
				var err error = nil
				var v interface{}
				ftype, _ := model.FieldByName(attr)
				kind := ftype.Type.Kind()
				switch kind {
				case reflect.Int:
					v, err = strconv.Atoi(val)
				case reflect.Uint:
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
					if kind == reflect.Map {
						v = reflect.MakeMap(ftype.Type)
					}
					err = fmt.Errorf(`Unhandled type "%v", cannot use normal html to set the attribute "%v" of custom tag "%v".
consider using attribute binding instead.`, kind, attr, tag.name)
				}

				if err != nil {
					panic(fmt.Sprintf(`Invalid value "%v" for attribute "%v" of custom tag "%v": type mismatch. Parse info: %v.`,
						val, attr, tag.name, err))
				}

				field.Set(reflect.ValueOf(v))
			}
		}

		tm.elemModels = append(tm.elemModels, cptr.Interface())
		elem.SetAttr(ModelIdAttr, strconv.Itoa(len(tm.elemModels)-1))
	})
}
