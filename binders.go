package wade

import (
	"fmt"
	"reflect"

	jq "github.com/gopherjs/jquery"
)

func defaultBinders() map[string]DomBinder {
	return map[string]DomBinder{
		"value": DomBinder{
			update: valueUpdateFn,
			watch:  valueWatchFn,
			bind:   nil,
		},
		"html": DomBinder{
			update: htmlUpdateFn,
			watch:  nil,
			bind:   nil,
		},
		"on": DomBinder{
			update: nil,
			watch:  nil,
			bind:   eventBindFn,
		},
	}
}

func htmlUpdateFn(elem jq.JQuery, value interface{}, args []string) {
	elem.SetHtml(value.(string))
}

func valueUpdateFn(elem jq.JQuery, value interface{}, args []string) {
	elem.SetVal(value.(string))
}

func valueWatchFn(elem jq.JQuery, ufn ModelUpdateFn) {
	tagname := elem.Prop("tagName")
	if tagname != "INPUT" && tagname != "TEXTAREA" && tagname != "SELECT" {
		panic("Can only watch for changes on html input, textarea and select.")
	}

	elem.On(jq.CHANGE, func(evt jq.Event) {
		ufn(elem.Val())
	})
}

func eventBindFn(elem jq.JQuery, value interface{}, args, outputs []string) {
	fnt := reflect.TypeOf(value)
	if fnt.Kind() != reflect.Func {
		panic("what used in event bind must be a function.")
	}
	if fnt.NumIn() > 0 {
		panic("function used in event bind must have no parameter.")
	}
	if len(args) > 1 {
		panic("Too many dash arguments to event bind.")
	}
	elem.On(args[0], value)
}

func eachBindFn(elem jq.JQuery, collection interface{}, args, outputs []string) {
	cv := reflect.ValueOf(collection)
	ki := cv.Type().Kind()
	var indexFn func(i int) reflect.Value
	var keys []reflect.Value
	switch ki {
	case reflect.Slice:
		indexFn = func(i int) reflect.Value {
			return cv.Index(i)
		}
	case reflect.Map:
		keys = cv.MapKeys()
		indexFn = func(i int) reflect.Value {
			return cv.MapIndex(keys[i])
		}
	default:
		panic(fmt.Sprintf("Wrong kind %v of target for the each binder, must be a slice or map.", ki.String()))
	}

	cl := elem.Clone()
	cl.RemoveAttr(BindPrefix + "each")
	marker := elem.Before("<!-- wade each -->")
	for i := 0; i < cv.Len(); i++ {
		v := indexFn(i)
		cl.Clone()
	}
}
