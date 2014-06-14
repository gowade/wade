package wade

import (
	"fmt"
	"reflect"

	"github.com/gopherjs/gopherjs/js"
)

func getReflectField(obj interface{}, field string) (reflect.Value, error) {
	o := reflect.ValueOf(obj)
	if o.Kind() == reflect.Ptr {
		o = o.Elem()
	}
	var rv reflect.Value
	switch o.Kind() {
	case reflect.Struct:
		rv = o.FieldByName(field)
	case reflect.Map:
		rv = o.MapIndex(reflect.ValueOf(field))
	default:
		return rv, fmt.Errorf("Unhandled type for accessing %v.", field)
	}

	if !rv.IsValid() {
		return rv, fmt.Errorf("Unable to access %v, field not available.", field)
	}

	return rv, nil
}

func getField(obj interface{}, field string) (interface{}, error) {
	rv, err := getReflectField(obj, field)
	if !rv.IsValid() || err != nil {
		return nil, err
	}
	return rv.Interface(), err
}

func rivetsInstallDotAdapter() {
	dota := gRivets.Get("adapters").Get(".") //it is not Dota, it's DotAdapter :V
	dota.Set("read", func(obj interface{}, key string) interface{} {
		v, err := getField(obj, key)
		if err != nil {
			panic(err.Error())
		}
		return v
	})
	dota.Set("publish", func(obj interface{}, key string, value interface{}) {
		v, err := getReflectField(obj, key)
		if err != nil {
			panic(err.Error())
		}
		v.Set(reflect.ValueOf(value))
	})
	dota.Set("subscribe", func(jso js.Object, keypath string, callback func()) {
		obj := jso.Interface()
		callbacks := dota.Call("weakReference", obj).Get("callbacks")
		cbs := callbacks.Get(keypath)
		if cbs.IsNull() || cbs.IsUndefined() {
			callbacks.Set(keypath, make([]interface{}, 0))
			cbs = callbacks.Get(keypath)

			v, err := getField(obj, keypath)
			if err != nil {
				panic(err.Error())
			}

			jso.Set("get", func() interface{} {
				return v
			})
			jso.Set("set", func(newValue interface{}) {
				if !reflect.DeepEqual(v, newValue) {
					v = newValue
					for i := 0; i < cbs.Length(); i++ {
						cbs.Index(i).Invoke()
					}
					dota.Call("observeMutations", newValue, jso.Get(dota.Get("id").Str()).Int(), keypath)
				}
			})
		}

		for i := 0; i < cbs.Length(); i++ {
			if interface{}(callback) != cbs.Index(i).Interface() {
				cbs.Call("push", callback)
			}
		}

		v, err := getField(obj, keypath)
		if err != nil {
			panic(err.Error())
		}
		dota.Call("observeMutations", v, jso.Get(dota.Get("id").Str()).Int(), keypath)
	})
}
