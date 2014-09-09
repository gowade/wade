package bind

import (
	"fmt"
	"reflect"

	"github.com/phaikawl/wade/dom"
	lb "github.com/phaikawl/wade/libs/binder"
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
		"class": &ClassBinder{},
	}
}

// ValueBinder is a 2-way binder that binds an element's value attribute.
// It takes no extra dash args.
// Meant to be used for <input>.
//
// Usage:
//	bind-value="Expression"
type ValueBinder struct{ *BaseBinder }

// Update sets the element's value attribute to a new value
func (b *ValueBinder) Update(d DomBind) (err error) {
	d.Elem.SetVal(toString(d.Value))
	return
}

// Watch watches for javascript change event on the element
func (b *ValueBinder) Watch(elem dom.Selection, ufn ModelUpdateFn) error {
	tagname, _ := elem.TagName()
	if tagname != "input" {
		return fmt.Errorf("Can only watch for changes on html input, textarea and select")
	}

	elem.On("change", func(evt dom.Event) {
		ufn(elem.Val())
	})

	return nil
}
func (b *ValueBinder) BindInstance() DomBinder { return b }

// HtmlBinder is a 1-way binder that binds an element's html content to
// the value of a model field.
// It takes no extra dash args.
//
// Usage:
//	bind-html="Expression"
type HtmlBinder struct{ BaseBinder }

// Update sets the element's html content to a new value
func (b *HtmlBinder) Update(d DomBind) error {
	d.Elem.SetHtml(toString(d.Value))
	return nil
}
func (b *HtmlBinder) BindInstance() DomBinder { return b }

// AttrBinder is a 1-way binder that binds a specified element's attribute
// to a model field value.
// It takes 1 extra dash arg that is the name of the html attribute to be bound.
//
// Usage:
//	bind-attr-<attribute>="Expression"
type AttrBinder struct{ BaseBinder }

func (b *AttrBinder) Update(d DomBind) error {
	if len(d.Args) != 1 {
		return fmt.Errorf(`Incorrect number of args (%v)
Correct Usage: bind-attr-<attribute>="Field".`, len(d.Args))
	}
	d.Elem.SetAttr(d.Args[0], toString(d.Value))

	return nil
}
func (b *AttrBinder) BindInstance() DomBinder { return b }

// ClassBinder is a 1-way binder that adds/removes (toggle) a class based on
// a boolean value.
// It takes 1 extra dash arg that is the name of the class to be bound.
//
// Usage:
//	bind-class-<class>="Expression"
type ClassBinder struct{ BaseBinder }

func (b *ClassBinder) Update(d DomBind) error {
	if len(d.Args) != 1 {
		return fmt.Errorf(`Incorrect number of args (%v)
Correct Usage: bind-class-<class>="Field".`, len(d.Args))
	}

	class := d.Args[0]
	enable := d.Value.(bool)
	hasClass := d.Elem.HasClass(class)
	if enable && !hasClass {
		d.Elem.AddClass(class)
	} else if !enable && hasClass {
		d.Elem.RemoveClass(class)
	}

	return nil
}

func (b *ClassBinder) BindInstance() DomBinder { return b }

// EventBinder is a 1-way binder that binds a method of the model to an event
// that occurs on the element.
// It takes 1 extra dash arg that is the event name, for example "click",
// "change",...
//
// Usage:
//	bind-on-<eventName>="HandlerMethod"
type EventBinder struct{ BaseBinder }

func heuristicPreventDefault(evtname string, elem dom.Selection) bool {
	if evtname == "click" {
		if elem.Is("button") {
			return true
		}

		if elem.Is("a") {
			href, hashref := elem.Attr("href")
			return !hashref || href == "" || href == "#"
		}
	}

	return false
}

func (b *EventBinder) Bind(d DomBind) error {
	fni := d.Value
	if fni == nil {
		return fmt.Errorf("Event must be bound to a handler function of type func() or func(dom.Event), not a nil. Note that generally the return value of a call is used for binding, not the call itself. So you may need to use a function that returns a handler function for this.")
	}
	handler0, ok0 := fni.(func())
	handler1, ok1 := fni.(func(dom.Event))

	if !ok0 && !ok1 {
		return fmt.Errorf("Wrong type %v for EventBinder's bind target, must be a function of type of type func() or func(dom.Event)",
			reflect.TypeOf(fni).String())
	}

	if len(d.Args) > 1 {
		return fmt.Errorf("Too many dash arguments to event bind")
	}

	evtname := d.Args[0]
	d.Elem.On(evtname, func(evt dom.Event) {
		if heuristicPreventDefault(evtname, d.Elem) {
			evt.PreventDefault()
		}
		go func() {
			if ok0 {
				//gopherjs:blocking
				handler0()
			} else if ok1 {
				//gopherjs:blocking
				handler1(evt)
			}
		}()
	})

	return nil
}
func (b *EventBinder) BindInstance() DomBinder { return b }

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
	marker    dom.Selection
	prototype dom.Selection
	size      int
}

func (b *EachBinder) BindInstance() DomBinder {
	return new(EachBinder)
}

func (b *EachBinder) Bind(d DomBind) (err error) {
	d.Elem.RemoveAttr("bind-each")
	b.marker = d.Elem.NewFragment("<!-- wade each -->")
	d.Elem.Before(b.marker)

	b.prototype = d.Elem.Clone()
	d.Banish(d.Elem)

	return
}

func (b *EachBinder) Update(d DomBind) (err error) {
	val := reflect.ValueOf(d.Value)

	for i := val.Len(); i < b.size; i++ {
		b.marker.Next().Remove()
	}

	for i := b.size; i < val.Len(); i++ {
		b.marker.After(b.prototype.Clone())
	}

	b.size = val.Len()

	prev := b.marker

	m, e := lb.GetLoopList(d.Value)
	if e != nil {
		err = e
		return
	}

	for _, item := range m {
		k, v := item.Key, item.Value
		nx := b.prototype.Clone()
		prev.Next().ReplaceWith(nx)
		err = d.ProduceOutputs(nx, true, true, k.Interface(), v.Interface())
		prev = nx
	}

	return
}

// PageBinder is used for <a> elements to set its href to the real page url
// and save necessary information for the proper page switching when the user
// clicks on the link. It should be used with the url() helper.
//
// Typical usage:
//	bind-page="url(`page-id`, arg1, arg2...)"
type PageBinder struct{ BaseBinder }

func (b *PageBinder) Update(d DomBind) error {
	tagname, _ := d.Elem.TagName()
	if tagname != "a" {
		return fmt.Errorf("bind-page can only be used for links (<a> elements)")
	}
	uinf := d.Value.(UrlInfo)
	d.Elem.SetAttr("href", uinf.fullUrl)
	d.Elem.SetAttr(WadePageAttr, uinf.path)

	return nil
}
func (b *PageBinder) BindInstance() DomBinder { return b }

// IfBinder keeps or remove an element according to a boolean field value.
//
// Usage:
//	bind-if="BooleanExpression"
type IfBinder struct {
	*BaseBinder
	placeholder dom.Selection
}

func (b *IfBinder) Bind(d DomBind) (err error) {
	b.placeholder = d.Elem.NewFragment("<!-- hidden elem -->")
	return
}

func (b *IfBinder) Update(d DomBind) (err error) {
	shown := d.Value.(bool)
	if shown && b.placeholder.Exists() {
		b.placeholder.ReplaceWith(d.Elem)
		return
	}

	if !shown && d.Elem.Exists() {
		d.Elem.ReplaceWith(b.placeholder)
	}

	return
}
func (b *IfBinder) BindInstance() DomBinder { return new(IfBinder) }

// UnlessBinder is the reverse of IfBinder.
//
// Usage:
//	bind-ifn="BooleanExpression"
type UnlessBinder struct {
	*IfBinder
}

func (b *UnlessBinder) Update(d DomBind) (err error) {
	d.Value = !(d.Value.(bool))
	b.IfBinder.Update(d)

	return
}
func (b *UnlessBinder) BindInstance() DomBinder { return &UnlessBinder{&IfBinder{}} }
