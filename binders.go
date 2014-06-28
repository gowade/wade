package wade

import (
	"reflect"
	"strings"

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
	tagname := strings.ToUpper(elem.Prop("tagName").(string))
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

//func eachBindFn(elem jq.JQuery, value interface{}, args, outputs []string) {

//}
