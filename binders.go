package wade

import (
	"fmt"
	"reflect"
	"strings"

	jq "github.com/gopherjs/jquery"
)

func defaultBinders() map[string]DomBinder {
	return map[string]DomBinder{
		"value": &ValueBinder{},
		"html":  &HtmlBinder{},
		"on":    &EventBinder{},
		"each":  new(EachBinder),
	}
}

type ValueBinder struct{}

func toString(value interface{}) string {
	return fmt.Sprintf("%v", value)
}

func (b *ValueBinder) Bind(binding *Binding, elem jq.JQuery, value interface{}, arg, outputs []string) {
}
func (b *ValueBinder) Update(elem jq.JQuery, value interface{}, args, outputs []string) {
	elem.SetVal(toString(value))
}
func (b *ValueBinder) Watch(elem jq.JQuery, ufn ModelUpdateFn) {
	tagname := strings.ToUpper(elem.Prop("tagName").(string))
	if tagname != "INPUT" && tagname != "TEXTAREA" && tagname != "SELECT" {
		panic("Can only watch for changes on html input, textarea and select.")
	}

	elem.On(jq.CHANGE, func(evt jq.Event) {
		ufn(elem.Val())
	})
}
func (b *ValueBinder) BindInstance() DomBinder { return b }

type HtmlBinder struct{}

func (b *HtmlBinder) Bind(binding *Binding, elem jq.JQuery, value interface{}, arg, outputs []string) {
}
func (b *HtmlBinder) Update(elem jq.JQuery, value interface{}, args, outputs []string) {
	elem.SetHtml(toString(value))
}
func (b *HtmlBinder) Watch(elem jq.JQuery, ufn ModelUpdateFn) {}
func (b *HtmlBinder) BindInstance() DomBinder                 { return b }

type EventBinder struct{}

func (b *EventBinder) Bind(binding *Binding, elem jq.JQuery, value interface{}, args, outputs []string) {
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
func (b *EventBinder) Update(elem jq.JQuery, value interface{}, args, outputs []string) {}
func (b *EventBinder) Watch(elem jq.JQuery, ufn ModelUpdateFn)                          {}
func (b *EventBinder) BindInstance() DomBinder                                          { return b }

type EachBinder struct {
	marker    jq.JQuery
	prototype jq.JQuery
	indexFn   func(i int, v reflect.Value) (interface{}, reflect.Value)
	size      int
	binding   *Binding
}

func (b *EachBinder) BindInstance() DomBinder {
	return new(EachBinder)
}

func (b *EachBinder) Bind(binding *Binding, elem jq.JQuery, value interface{}, arg, outputs []string) {
	kind := reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.Slice:
		b.indexFn = func(i int, v reflect.Value) (interface{}, reflect.Value) {
			return i, v.Index(i)
		}
	case reflect.Map:
		b.indexFn = func(i int, v reflect.Value) (interface{}, reflect.Value) {
			key := v.MapKeys()[i]
			return key.String(), v.MapIndex(key)
		}
	default:
		panic(fmt.Sprintf("Wrong kind %v of target for the each binder, must be a slice or map.", kind.String()))
	}

	elem.RemoveAttr(BindPrefix + "each")
	b.marker = gJQ("<!-- wade each -->").InsertBefore(elem).First()
	b.prototype = elem.Clone()
	b.binding = binding
	println("Remove that shit!")
	elem.Remove()
}

func (b *EachBinder) Update(elem jq.JQuery, collection interface{}, args, outputs []string) {
	val := reflect.ValueOf(collection)

	for i := val.Len(); i < b.size; i++ {
		b.marker.Next().Remove()
	}

	for i := b.size; i < val.Len(); i++ {
		b.marker.After(b.prototype.Clone())
	}

	prev := b.marker
	m := make(map[string]interface{})
	noutput := len(outputs)
	if noutput > 2 {
		panic(fmt.Sprintf("Wrong output specification %v for the Each binder: only up to 2 outputs are allowed.", outputs))
	}
	if noutput != 0 {
		for i := val.Len() - 1; i >= 0; i-- {
			k, v := b.indexFn(i, val)
			nx := prev.Next()
			if noutput == 1 {
				m[outputs[0]] = v.Interface()
			} else {
				m[outputs[0]] = k
				m[outputs[1]] = v.Interface()
			}
			b.binding.Bind(nx, m, true)
			prev = nx
		}
	}
}
func (b *EachBinder) Watch(elem jq.JQuery, ufn ModelUpdateFn) {}
