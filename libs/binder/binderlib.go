package binder

import (
	"fmt"
	"reflect"
)

type (
	Item struct {
		Key   reflect.Value
		Value reflect.Value
	}
)

func GetLoopList(val interface{}) ([]Item, error) {
	v := reflect.ValueOf(val)
	list := make([]Item, 0)
	switch v.Kind() {
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			list = append(list, Item{reflect.ValueOf(i), v.Index(i)})
		}

		return list, nil
	case reflect.Map:
		for _, key := range v.MapKeys() {
			list = append(list, Item{key, v.MapIndex(key)})
		}
		return list, nil
	}

	return nil, fmt.Errorf("Wrong type for collection, the value passed to NewCollection must be a pointer to slice or pointer to map.")
}
