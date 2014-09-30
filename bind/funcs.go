package bind

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gopherjs/gopherjs/js"
)

func toString(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

func jsGetType(obj js.Object) string {
	return js.Global.Get("Object").Get("prototype").Get("toString").Call("call", obj).Str()
}

func callFunc(fn reflect.Value, args []reflect.Value) (v reflect.Value, err error) {
	rets := fn.Call(args)
	if len(rets) >= 1 {
		v = rets[0]
		return
	}

	return
}

// evaluateObj uses reflection to access a field (obj.field1.field2.field3) of the given model.
// It returns an evaluation of the field, and a bool which indicates whether the field is found
func evaluateObjField(query string, model reflect.Value) (oe *objEval, ok bool, err error) {
	flist := strings.Split(query, ".")
	vals := make([]reflect.Value, len(flist)+1)
	o := model

	if o.Kind() == reflect.Ptr {
		o = o.Elem()
	}
	vals[0] = o

	for i, field := range flist {
		o, ok, err = getReflectField(o, field)
		if err != nil {
			return
		}

		if !ok {
			return
		}

		vals[i+1] = o
	}

	ok = true
	oe = &objEval{
		fieldRefl: vals[len(vals)-1],
		modelRefl: vals[len(vals)-2],
		field:     flist[len(flist)-1],
	}

	return
}

// getReflectField returns the field value of an object, be it a struct instance
// or a map
func getReflectField(o reflect.Value, field string) (rv reflect.Value, ok bool, err error) {
	if o.Kind() == reflect.Ptr && o.IsNil() {
		err = fmt.Errorf("Accessing field of a nil pointer")
		return
	}

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

	case reflect.Slice:
		var num int
		_, err = fmt.Sscan(field, &num)
		rv = o.Index(num)

	default:
		return
	}

	ok = rv.IsValid()

	return
}
