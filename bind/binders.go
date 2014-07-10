package bind

import (
	"fmt"
	"reflect"
	"strings"

	jq "github.com/gopherjs/jquery"
)

const (
	WadePageAttr = "data-wade-page"
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

// BaseBinder provides the base so that binders will not have to provide empty
// implement for the methods
type BaseBinder struct{}

func (b *BaseBinder) Bind(binding *Binding, elem jq.JQuery, value interface{}, args, outputs []string) {
}
func (b *BaseBinder) Update(elem jq.JQuery, value interface{}, args, outputs []string) {}
func (b *BaseBinder) Watch(elem jq.JQuery, ufn ModelUpdateFn)                          {}

// ValueBinder is a 2-way binder that binds an element's value attribute.
// It takes no extra dash args.
// Meant to be used for <input>.
//
// Usage:
//	bind-value="Expression"
type ValueBinder struct{ *BaseBinder }

// Update sets the element's value attribute to a new value
func (b *ValueBinder) Update(elem jq.JQuery, value interface{}, args, outputs []string) {
	elem.SetVal(toString(value))
}

// Watch watches for javascript change event on the element
func (b *ValueBinder) Watch(elem jq.JQuery, ufn ModelUpdateFn) {
	tagname := strings.ToUpper(elem.Prop("tagName").(string))
	if tagname != "INPUT" {
		panic("Can only watch for changes on html input, textarea and select.")
	}

	elem.On(jq.CHANGE, func(evt jq.Event) {
		ufn(elem.Val())
	})
}
func (b *ValueBinder) BindInstance() DomBinder { return b }

// ValueBinder is a 1-way binder that binds an element's html content to
// the value of a model field.
// It takes no extra dash args.
//
// Usage:
//	bind-html="Expression"
type HtmlBinder struct{ BaseBinder }

// Update sets the element's html content to a new value
func (b *HtmlBinder) Update(elem jq.JQuery, value interface{}, args, outputs []string) {
	elem.SetHtml(toString(value))
}
func (b *HtmlBinder) BindInstance() DomBinder { return b }

// AttrBinder is a 1-way binder that binds a specified element's attribute
// to a model field value.
// It takes 1 extra dash arg that is the name of the html attribute to be bound.
//
// Usage:
//	bind-attr-thatAttribute="Expression"
type AttrBinder struct{ BaseBinder }

func (b *AttrBinder) Update(elem jq.JQuery, value interface{}, args, outputs []string) {
	if len(args) != 1 {
		panic(fmt.Sprintf(`Incorrect number of args %v for html attribute binder.
Usage: bind-attr-thatAttribute="Field".`, len(args)))
	}
	elem.SetAttr(args[0], toString(value))
}
func (b *AttrBinder) BindInstance() DomBinder { return b }

// EventBinder is a 1-way binder that binds a method of the model to an event
// that occurs on the element.
// It takes 1 extra dash arg that is the event name, for example "click",
// "change",...
//
// Usage:
//	bind-on-thatEventName="HandlerMethod"
type EventBinder struct{ BaseBinder }

func (b *EventBinder) Bind(binding *Binding, elem jq.JQuery, fni interface{}, args, outputs []string) {
	fn, ok := fni.(func())
	if !ok {
		panic(fmt.Sprintf("Wrong type %v for EventBinder's handler, must be of type func().",
			reflect.TypeOf(fni).String()))
	}
	if len(args) > 1 {
		panic("Too many dash arguments to event bind.")
	}
	elem.On(args[0], func(evt jq.Event) {
		evt.PreventDefault()
		fn()
	})
}
func (b *EventBinder) BindInstance() DomBinder { return b }

type indexFunc func(i int, v reflect.Value) (interface{}, reflect.Value)

// EachBinder is a 1-way binder that repeats an element according to a map
// or slice. It outputs a key and a value bound to each item.
// It takes no extra dash arg. The extra output after "->" are the names that
// receives the key and value, those names can be used inside the elment's
// content. Each key and value pair is bound separately to each element.
//
// Usage:
//	bind-each="Expression"
// Or
//	bind-each="Expression -> outputKey, outputValue"
// Example:
//	<div bind-each="Errors -> type, msg">
//		<p>Error type: <% type %></p>
//		<p>Message: <% msg %></p>
//	</div>
type EachBinder struct {
	*BaseBinder
	marker    jq.JQuery
	prototype jq.JQuery
	indexFn   indexFunc
	size      int
	binding   *Binding
}

func (b *EachBinder) BindInstance() DomBinder {
	return new(EachBinder)
}

func getIndexFunc(value interface{}) indexFunc {
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
	default:
		panic(fmt.Sprintf("Wrong kind %v of target for the each binder, must be a slice or map.", kind.String()))
	}
}

func (b *EachBinder) Bind(binding *Binding, elem jq.JQuery, value interface{}, arg, outputs []string) {
	elem.RemoveAttr(BindPrefix + "each")
	b.indexFn = getIndexFunc(value)
	b.marker = gJQ("<!-- wade each -->").InsertBefore(elem).First()
	b.prototype = elem.Clone()
	b.binding = binding

	elem.Remove()
	PreventBinding(elem)
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
	if noutput == 2 {
		for i := 0; i < b.size; i++ {
			k, v := b.indexFn(i, val)
			nx := b.prototype.Clone()
			prev.Next().ReplaceWith(nx)
			m[outputs[0]] = k
			m[outputs[1]] = v.Interface()
			b.binding.Bind(nx, m, true)
			prev = nx
		}
	} else if noutput != 0 {
		panic(fmt.Sprintf("Wrong output specification %v for the Each binder: there must be 2 outputs.", outputs))
	}
}

// PageBinder is used for <a> elements to set its href to the real page url
// and save necessary information for the proper page switching when the user
// clicks on the link. It should be used with the url() helper.
//
// Typical usage:
//	bind-page="url(`page-id`, arg1, arg2...)"
type PageBinder struct{ BaseBinder }

func (b *PageBinder) Update(elem jq.JQuery, value interface{}, args, outputs []string) {
	if strings.ToLower(elem.Prop("tagName").(string)) != "a" {
		panic("bind-page can only be used for links (<a> elements).")
	}
	uinf := value.(UrlInfo)
	elem.SetAttr("href", uinf.fullUrl)
	elem.SetAttr(WadePageAttr, uinf.path)
}
func (b *PageBinder) BindInstance() DomBinder { return b }

// IfBinder shows or remove an element according to a boolean field value.
//
// Usage:
//	bind-if="BooleanExpression"
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
		b.placeholder.ReplaceWith(elem)
		return
	}

	if !shown && jqExists(elem) {
		elem.ReplaceWith(b.placeholder)
	}
}
func (b *IfBinder) BindInstance() DomBinder { return new(IfBinder) }

// UnlessBinder is the reverse of IfBinder.
//
// Usage:
//	bind-ifn="BooleanExpression"
type UnlessBinder struct {
	*IfBinder
}

func (b *UnlessBinder) Update(elem jq.JQuery, value interface{}, args, outputs []string) {
	shown := !(value.(bool))
	b.IfBinder.Update(elem, shown, args, outputs)
}
func (b *UnlessBinder) BindInstance() DomBinder { return &UnlessBinder{new(IfBinder)} }
