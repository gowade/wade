package wade

import (
	"reflect"

	jq "github.com/gopherjs/jquery"
)

func htmlUpdateFn(elem jq.JQuery, value interface{}, arg []string) {
	elem.SetHtml(value.(string))
}

func valueUpdateFn(elem jq.JQuery, value interface{}, arg []string) {
	println(value)
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

func eventBindFn(elem jq.JQuery, value interface{}, arg []string) {
	fnt := reflect.TypeOf(value)
	if fnt.Kind() != reflect.Func {
		panic("what used in event bind must be a function.")
	}
	if fnt.NumIn() > 0 {
		panic("function used in event bind must have no parameter.")
	}
	if len(arg) > 1 {
		panic("Too many dash arguments to event bind.")
	}
	elem.On(arg[0], value)
}
