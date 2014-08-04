package bind

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/gopherjs/gopherjs/js"
	jq "github.com/gopherjs/jquery"
)

func elemError(elem jq.JQuery, errstr string) {
	msg := fmt.Sprintf(`Error while processing: "%v"`, elem.Clone().Wrap("<p>").Parent().Html())
	if len(msg) >= 200 {
		msg = msg[0:200] + "[...]"
	}
	println(msg)
	panic(errstr)
}

func jqExists(elem jq.JQuery) bool {
	return elem.Parents("html").Length > 0
}

func isValidExprChar(c rune) bool {
	return c == '`' || c == '.' || c == '_' || unicode.IsLetter(c) || unicode.IsDigit(c)
}

func jsGetType(obj js.Object) string {
	return js.Global.Get("Object").Get("prototype").Get("toString").Call("call", obj).Str()
}

func callFunc(fn reflect.Value, args []reflect.Value) (v reflect.Value, err error) {
	ftype := fn.Type()
	nin := ftype.NumIn()
	var ok bool
	if ftype.IsVariadic() {
		ok = len(args) >= nin-1
	} else {
		ok = nin == len(args)
	}

	if !ok {
		err = fmt.Errorf(`Invalid number of arguments.`)
		return
	}

	rets := fn.Call(args)
	if len(rets) == 1 {
		v = rets[0]
		return
	}

	return
}

// evaluateObj uses reflection to access a field (obj.field1.field2.field3) of the given model.
// It returns an evaluation of the field, and a bool which indicates whether the field is found
func evaluateObjField(query string, model reflect.Value) (*objEval, bool) {
	flist := strings.Split(query, ".")
	vals := make([]reflect.Value, len(flist)+1)
	o := model

	if o.Kind() == reflect.Ptr {
		o = o.Elem()
	}
	vals[0] = o

	for i, field := range flist {
		var found bool
		o, found = getReflectField(o, field)
		if !found {
			return nil, false
		}
		vals[i+1] = o
	}

	return &objEval{
		fieldRefl: vals[len(vals)-1],
		modelRefl: vals[len(vals)-2],
		field:     flist[len(flist)-1],
	}, true
}

// getReflectField returns the field value of an object, be it a struct instance
// or a map
func getReflectField(o reflect.Value, field string) (reflect.Value, bool) {
	var rv reflect.Value

	if o.Kind() == reflect.Ptr {
		o = o.Elem()
	}

	switch o.Kind() {
	case reflect.Struct:
		rv = o.FieldByName(field)
		if !rv.IsValid() && o.CanAddr() {
			rv = o.Addr().MethodByName(field)
		}
	case reflect.Map:
		rv = o.MapIndex(reflect.ValueOf(field))
		if rv.IsValid() {
			rv = reflect.ValueOf(rv.Interface())
		}
	default:
		return rv, false
	}

	if rv.IsValid() {
		return rv, true
	}

	return rv, false
}
