package wade

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	jq "github.com/gopherjs/jquery"
	"github.com/phaikawl/wade/services/http"
)

var (
	gHistory js.Object
	gJQ      = jq.NewJQuery
)

const (
	AttrPrefix = "attr-"
	BindPrefix = "bind-"
)

type Wade struct {
	pm         *PageManager
	binding    *binding
	elemModels []interface{}
	custags    map[string]*CustomTag
}

type CustomTag struct {
	meid  string //id of the model welement used to declare the tag content
	model interface{}
}

type ErrorMap map[string]map[string]interface{}

type Validated struct {
	Errors ErrorMap
}

type ErrorsBinding struct {
	Errors *ErrorMap
}

func (v *Validated) Init(dataModel interface{}) {
	m := make(ErrorMap)
	typ := reflect.TypeOf(dataModel)
	if typ.Kind() != reflect.Struct {
		panic("Validated data model passed to Init() must be a struct.")
	}
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		m[f.Name] = make(map[string]interface{})
	}
	v.Errors = m
}

func WadeUp(startPage, basePath string, initFn func(*Wade)) *Wade {
	gHistory = js.Global.Get("history")
	wd := &Wade{
		pm:      newPageManager(startPage, basePath),
		binding: newBindEngine(nil),
		custags: make(map[string]*CustomTag),
	}
	wd.binding.wade = wd
	wd.Init()
	initFn(wd)
	return wd
}

func (wd *Wade) Pager() *PageManager {
	return wd.pm
}

func (wd *Wade) htmlImport(parent jq.JQuery, origin string) {
	parent.Find("import").Each(func(i int, elem jq.JQuery) {
		src := elem.Attr("src")
		req := http.NewRequest(http.MethodGet, origin+src)
		html := req.DoSync()
		elem.Hide()
		elem.Append(html)
		wd.htmlImport(elem, origin)
	})
}

func (wd *Wade) Init() {
	origin := js.Global.Get("document").Get("location").Get("origin").Str()
	wd.htmlImport(gJQ("body"), origin)
}

func (wd *Wade) modelForCustomElem(elem jq.JQuery) interface{} {
	modelId := int(elem.Data("modelId").(float64))
	return wd.elemModels[modelId]
}

func (wd *Wade) Start() {
	gJQ(js.Global.Get("document")).Ready(func() {
		pageHide(gJQ("wpage"))
		wd.pm.getReady()
		for _, tag := range wd.custags {
			mtype := reflect.TypeOf(tag.model)
			if mtype.Kind() != reflect.Struct {
				panic(fmt.Sprintf("Wrong type for the model of tag #%v, it must be a struct (non-pointer).", tag.meid))
			}
			wd.prepareCustomTags(tag.meid, mtype)
		}
		wd.pm.bindPage(wd.binding)
		for tagName, tag := range wd.custags {
			tagElem := gJQ("#" + tag.meid)
			elems := gJQ(tagName)
			elems.Each(func(i int, elem jq.JQuery) {
				elem.Append(tagElem.Html())
				wd.binding.Bind(elem, wd.modelForCustomElem(elem))
			})
		}
	})

	gJQ("import").Show()
}

func (wd *Wade) RegisterNewTag(tagid string, model interface{}) {
	tagElem := gJQ("#" + tagid)
	if tagElem.Length == 0 {
		panic(fmt.Sprintf("Welement with id #%v does not exist.", tagid))
	}
	if tagElem.Prop("tagName") != "WELEMENT" {
		panic(fmt.Sprintf("The element #%v to register new tag must be a welement.", tagid))
	}
	wd.custags[strings.ToUpper(tagid)] = &CustomTag{tagid, model}
}

func (wd *Wade) prepareCustomTags(tagid string, model reflect.Type) {
	tagElem := gJQ("#" + tagid)
	publicAttrs := []string{}
	if attrs := tagElem.Attr("attributes"); attrs != "" {
		publicAttrs = strings.Split(attrs, " ")
		for _, attr := range publicAttrs {
			if _, ok := model.FieldByName(attr); !ok {
				panic(fmt.Sprintf(`Attribute "%v" is not available in the model for custom tag "%v".`, attr, tagid))
			}
		}
	}

	elems := gJQ(tagid)
	elems.Each(func(idx int, elem jq.JQuery) {
		cptr := reflect.New(model)
		clone := cptr.Elem()
		for _, attr := range publicAttrs {
			if val := elem.Attr(AttrPrefix + attr); val != "" {
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
consider using attribute binding instead.`, kind, attr, tagid)
				}

				if err != nil {
					panic(fmt.Sprintf(`Invalid value "%v" for attribute "%v" of custom tag "%v": type mismatch. Parse info: %v.`,
						val, attr, tagid, err))
				}

				field.Set(reflect.ValueOf(v))
			}
		}

		wd.elemModels = append(wd.elemModels, cptr.Interface())
		elem.SetData("modelId", len(wd.elemModels)-1)
	})
}
