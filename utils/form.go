package utils

import (
	"reflect"

	"github.com/pengux/check"
	"github.com/phaikawl/wade/libs/http"
)

type ErrorMap map[string]map[string]string

// Validated is the base for a model used in a form that is meant to be validated.
// It holds an Errors map that receives the error data.
//
// Must be embedded into a data struct to be used with ProcessForm.
type Validated struct {
	Errors ErrorMap
}

func (v *Validated) setErrors(m ErrorMap) {
	for k := range v.Errors {
		if mv, ok := m[k]; ok {
			v.Errors[k] = mv
		} else {
			v.Errors[k] = make(map[string]string)
		}
	}
}

// Init creates a map entry for each field in dataModel.
// This method must be run when creating a form model instance so that the
// Errors map has the necessary fields to be used for listing errors.
func (v *Validated) Init(dataModel interface{}) {
	m := make(ErrorMap)
	typ := reflect.TypeOf(dataModel)
	if typ.Kind() != reflect.Struct {
		panic("Validated data model passed to Init() must be a struct.")
	}
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		m[f.Name] = make(map[string]string)
	}
	v.Errors = m
}

type ErrorMapHolder interface {
	setErrors(ErrorMap)
}

// ProcessForm validates and sends the given data struct to a given url
// and puts validation errors into data's Errors field.
//
// Validated implements the ErrorMapHolder interface, so you just need to
// embed Validated into your form struct and pass it to this function.

func ProcessForm(httpClient *http.Client, url string, data interface{}, errdst ErrorMapHolder, validator check.Struct) (*http.Response, error) {
	if reflect.TypeOf(data).Kind() != reflect.Struct {
		panic("The dataModel given to ProcessForm must be a struct.")
	}
	errdst.setErrors(validator.Validate(data).ToMessages())
	resp, err := httpClient.POST(url, data)
	return resp, err
}
