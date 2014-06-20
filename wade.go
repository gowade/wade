package wade

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	jq "github.com/gopherjs/jquery"
)

var (
	gHistory = js.Global.Get("history")
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
	custags    map[string]interface{} //custom tags
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

func WadeUp(startPage, basePath string) *Wade {
	return &Wade{
		pm:      newPageManager(startPage, basePath),
		binding: newBindEngine(),
		custags: make(map[string]interface{}),
	}
}

func (wd *Wade) Pager() *PageManager {
	return wd.pm
}

func (wd *Wade) Start() {
	gJQ(js.Global.Get("document")).Ready(func() {
		pageHide(gJQ("wpage"))
		wd.pm.getReady()
		//for tagid, model := range wd.custags {
		//	mtype := reflect.TypeOf(model)
		//	if mtype.Kind() != reflect.Struct {
		//		panic(fmt.Sprintf("Wrong type for the model of tag #%v, it must be a struct (non-pointer).", tagid))
		//	}
		//	wd.prepare(tagid, mtype)
		//}

		wd.pm.bindPage(wd.binding)
		//for tagid, _ := range wd.custags {
		//	tagElem := gJQ("#" + tagid)
		//	elems := gJQ(tagid)
		//	elems.Each(func(i int, elem jq.JQuery) {
		//		modelId := int(elem.Data("modelId").(float64))
		//		elem.Append("<div>" + tagElem.Html() + "</div>")
		//		gRivets.Call("bind", elem.Children("").First().Underlying(), wd.elemModels[modelId])
		//	})
		//}
	})
}

func (wd *Wade) prepare(tagid string, model reflect.Type) {
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

	bindables := []string{}
	if bdbs := tagElem.Attr("bindables"); bdbs != "" {
		bindables = strings.Split(bdbs, " ")
		for _, bdb := range bindables {
			if _, ok := model.FieldByName(bdb); !ok {
				panic(fmt.Sprintf(`Bindable "%v" is not available in the model for custom tag "%v".`, bdb, tagid))
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
				switch ftype.Type.Kind() {
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
					err = fmt.Errorf(`Unhandled type for attribute "%v" of custom tag "%v".`, attr, tagid)
				}

				if err != nil {
					panic(fmt.Sprintf(`Invalid value "%v" for attribute "%v" of custom tag "%v": type mismatch. Parse info: %v.`,
						val, attr, tagid, err))
				}

				field.Set(reflect.ValueOf(v))
			}
		}

		for _, bdb := range bindables {
			f := clone.FieldByName(bdb)
			ftyp, _ := model.FieldByName(bdb)
			switch f.Type().Kind() {
			case reflect.Map:
				f.Set(reflect.MakeMap(ftyp.Type))
			}
		}

		wd.elemModels = append(wd.elemModels, cptr.Interface())
		elem.SetData("modelId", len(wd.elemModels)-1)
	})
}
