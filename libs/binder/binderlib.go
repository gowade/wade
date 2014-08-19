package binder

import (
	"reflect"
)

type IndexFunc func(i int, collection reflect.Value) (key interface{}, value reflect.Value)

func GetIndexFunc(value interface{}) IndexFunc {
	kind := reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.Slice:
		return func(i int, val reflect.Value) (interface{}, reflect.Value) {
			return i, val.Index(i)
		}
	case reflect.Map:
		return func(i int, val reflect.Value) (interface{}, reflect.Value) {
			key := val.MapKeys()[i]
			return key.String(), val.MapIndex(key)
		}
	}

	return nil
}
