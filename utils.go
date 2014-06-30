package wade

import (
	"reflect"
	"unicode"

	"github.com/phaikawl/wade/lib"
	"github.com/phaikawl/wade/services/http"
)

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

type FormResp lib.FormResp

func camelize(src string) string {
	res := []rune{}
	startW := true
	for _, c := range src {
		if c == '-' {
			startW = true
			continue
		}
		ch := c
		if startW {
			ch = unicode.ToUpper(c)
			startW = false
		}
		res = append(res, ch)
	}
	return string(res)
}

func SendFormTo(url string, data interface{}, valdErrs *Validated) *http.Response {
	req := http.Service().NewRequest(http.MethodPost, url)
	req.SetData(data)
	r := req.DoSync()
	err := r.DecodeDataTo(&valdErrs.Errors)
	if err != nil {
		panic(err.Error())
	}
	return r
}
