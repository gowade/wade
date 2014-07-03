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
		"attr":  &AttrBinder{},
		"on":    &EventBinder{},
		"each":  new(EachBinder),
		"page":  &PageBinder{},
		"if":    new(IfBinder),
		"ifn":   &UnlessBinder{&IfBinder{}},
	}
}

type BaseBinder struct{}

func (b *BaseBinder) Bind(binding *Binding, elem jq.JQuery, value interface{}, args, outputs []string) {
}
func (b *BaseBinder) Update(elem jq.JQuery, value interface{}, args, outputs []string) {}
func (b *BaseBinder) Watch(elem jq.JQuery, ufn ModelUpdateFn)                          {}

type ValueBinder struct{ *BaseBinder }

func toString(value interface{}) string {
	return fmt.Sprintf("%v", value)
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

type HtmlBinder struct{ BaseBinder }

func (b *HtmlBinder) Update(elem jq.JQuery, value interface{}, args, outputs []string) {
	elem.SetHtml(toString(value))
}
func (b *HtmlBinder) BindInstance() DomBinder { return b }

type AttrBinder struct{ BaseBinder }

func (b *AttrBinder) Update(elem jq.JQuery, value interface{}, args, outputs []string) {
	if len(args) != 1 {
		panic(fmt.Sprintf(`Incorrect number of args %v for html attribute binder.
Usage: bind-attr-the_attr_name=the_attr_value.`, len(args)))
	}
	elem.SetAttr(args[0], toString(value))
}
func (b *AttrBinder) BindInstance() DomBinder { return b }

type EventBinder struct{ BaseBinder }

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
func (b *EventBinder) BindInstance() DomBinder { return b }

type EachBinder struct {
	*BaseBinder
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

	b.size = val.Len()

	prev := b.marker
	m := make(map[string]interface{})
	noutput := len(outputs)
	if noutput > 2 {
		panic(fmt.Sprintf("Wrong output specification %v for the Each binder: only up to 2 outputs are allowed.", outputs))
	}
	if noutput != 0 {
		for i := 0; i < b.size; i++ {
			k, v := b.indexFn(i, val)
			nx := b.prototype.Clone()
			prev.Next().ReplaceWith(nx)
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

type PageBinder struct{ BaseBinder }

func (b *PageBinder) Update(elem jq.JQuery, value interface{}, args, outputs []string) {
	uinf := value.(UrlInfo)
	elem.SetAttr("href", uinf.fullUrl)
	elem.SetAttr(WadePageAttr, uinf.path)
}
func (b *PageBinder) BindInstance() DomBinder { return b }

type IfBinder struct {
	*BaseBinder
	placeholder jq.JQuery
}

func (b *IfBinder) Bind(binding *Binding, elem jq.JQuery, value interface{}, args, outputs []string) {
	b.placeholder = gJQ("<!-- hidden elem -->")
}

func (b *IfBinder) Update(elem jq.JQuery, value interface{}, args, outputs []string) {
	shown := value.(bool)
	if shown && !jqExists(elem) {
		if jqExists(b.placeholder) {
			b.placeholder.ReplaceWith(elem)
			return
		}
		panic("Unexpected case, there might be a bug.")
	}

	if !shown && jqExists(elem) {
		elem.ReplaceWith(b.placeholder)
	}
}
func (b *IfBinder) BindInstance() DomBinder { return new(IfBinder) }

type UnlessBinder struct {
	*IfBinder
}

func (b *UnlessBinder) Update(elem jq.JQuery, value interface{}, args, outputs []string) {
	shown := !(value.(bool))
	b.IfBinder.Update(elem, shown, args, outputs)
}
func (b *UnlessBinder) BindInstance() DomBinder { return &UnlessBinder{new(IfBinder)} }
