package bind

import (
	"fmt"
	"reflect"
	"time"

	"github.com/gopherjs/gopherjs/js"
	"github.com/phaikawl/wade/dom"
	lb "github.com/phaikawl/wade/libs/binder"
)

var (
	ClientSide = js.Global != nil && !js.Global.Get("window").IsUndefined()
)

func defaultBinders() map[string]DomBinder {
	return map[string]DomBinder{
		"value": &ValueBinder{},
		"html":  &HtmlBinder{},
		"on":    &EventBinder{},
		"each":  new(EachBinder),
		"if":    new(IfBinder),
		"ifn":   &UnlessBinder{&IfBinder{}},
		"class": &ClassBinder{},
	}
}

// ValueBinder is a 2-way binder that binds an element's value attribute to a value.
// Meant to be used for <input> and <textarea>.
//
// It has an optional argument, catchKeyup which if true, the binder also performs
// value update on keyup. Defaults to false.
//
// Usage:
//	#value({optional}catchKeyup)="Expression"
type ValueBinder struct{ *BaseBinder }

// Update sets the element's value attribute to a new value
func (b *ValueBinder) Update(d DomBind) (err error) {
	d.Elem.SetVal(toString(d.Value))
	return
}

// Watch watches for javascript change event on the element.
//
func (b *ValueBinder) Watch(d DomBind, ufn ModelUpdateFn) error {
	elem := d.Elem
	tagname, _ := elem.TagName()
	if tagname != "input" && tagname != "textarea" {
		return fmt.Errorf("Can only watch for changes on html input, textarea and select")
	}

	events := "change"
	if len(d.Args) == 1 && d.Args[0] == "true" {
		events += " keyup"
	}

	elem.On(events, func(evt dom.Event) {
		ufn(elem.Val())
	})

	return nil
}
func (b *ValueBinder) BindInstance() DomBinder { return b }

// HtmlBinder is a 1-way binder that binds an element's html content to a value.
//
// Usage:
//	#html="Expression"
type HtmlBinder struct{ BaseBinder }

func (b *HtmlBinder) Update(d DomBind) error {
	d.Elem.SetHtml(toString(d.Value))
	return nil
}

func (b *HtmlBinder) BindInstance() DomBinder { return b }

// ClassBinder is a 1-way binder that adds/removes (toggle) a class based on
// a boolean value.
//
// It takes 1 required argument className which is the class name.
//
// Usage:
//	#class({required}className)="Expression"
type ClassBinder struct{ BaseBinder }

func (b *ClassBinder) Update(d DomBind) error {
	if len(d.Args) != 1 {
		return fmt.Errorf(`Incorrect number of args (%v). Need 1 argument.`, len(d.Args))
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

// EventBinder is a 1-way binder that binds an element's event to a function or method.
//
// It takes 1 argument that is the event name.
//
// Usage:
//	#on({required}eventName)="Expression"
//
//
// The Expression is evaluated like any other expressions, if you call a function,
// the value that gets bound is the return value, not the function call.
// You can use Wade.Go's '@' syntax to conveniently wrap a function call.
//
// Example: we have a function func Fn(super bool) int. You want to call Fn with super=true on click event.
// You would do it like this:
//  #on(click)="@Fn(true)"
// Note that
//  #on(click)="Fn(true)"
// is invalid because the bind value here is an int returned by Fn, an error will be raised.
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

// EachBinder is a 1-way binder that repeats an element according to a slice.
//
// It takes 2 arguments: the name to bind the index to, and the name to bind the value to.
// Usage:
//  #each=({required}indexBindName, [required]valueBindName)
//
// Example:
//  #each(i,product)="$Products"
//  #each(_,product)="$Products"
type EachBinder struct {
	*BaseBinder
	marker    dom.Selection
	prototype dom.Selection
	size      int
	lc        *listChanger
}

func (b *EachBinder) BindInstance() DomBinder {
	return new(EachBinder)
}

func (b *EachBinder) Bind(d DomBind) (err error) {
	b.marker = d.Elem.NewFragment("<!-- wade each -->")
	d.Elem.Before(b.marker)

	b.prototype = d.Elem.Clone()
	d.Banish(d.Elem)

	return
}

func (b *EachBinder) FullUpdate(d DomBind) (err error) {
	//populate the list
	val := reflect.ValueOf(d.Value)

	for i := val.Len(); i < b.size; i++ {
		b.marker.Next().Remove()
	}

	for i := b.size; i < val.Len(); i++ {
		b.marker.After(b.prototype.Clone())
	}

	b.size = val.Len()

	m, e := lb.GetLoopList(d.Value)
	if e != nil {
		err = e
		return
	}
	next := b.marker.Next()

	//js.Global.Get("console").Call("profile", "list")
	for i, item := range m {
		k, v := item.Key, item.Value
		nx := b.prototype.Clone()
		tnext := next.Next()
		next.ReplaceWith(nx)

		err = d.ProduceOutputs(nx, false, d.Args[:2], k.Interface(), v.Interface())
		if ClientSide {
			if i%10 == 0 {
				time.Sleep(0 * time.Millisecond)
			}
		}
		next = tnext
	}
	//js.Global.Get("console").Call("profileEnd")

	return
}

type listChanger struct {
	binder *EachBinder
	d      *DomBind
}

func (lc *listChanger) Add(i int, value reflect.Value) {
	children := lc.binder.marker.Parent().Contents().Elements()
	newe := lc.binder.prototype.Clone()
	midx := lc.binder.marker.Index()
	for mi := 0; midx+1+mi < len(children); mi++ {
		if children[midx+1+mi].IsElement() {
			children[midx+1+mi+i].Before(newe)
			lc.d.ProduceOutputs(newe, false, lc.d.Args[:2], i, value.Interface())
			return
		}
	}
	lc.binder.marker.After(newe)
	lc.d.ProduceOutputs(newe, false, lc.d.Args[:2], i, value.Interface())
}

func (lc *listChanger) Remove(i int) {
	children := lc.binder.marker.Parent().Contents().Elements()
	midx := lc.binder.marker.Index()
	for i := 0; ; i++ {
		if children[midx+1+i].IsElement() {
			children[midx+1+i].Remove()
			break
		}
	}
}

func (b *EachBinder) Update(d DomBind) (err error) {
	if reflect.TypeOf(d.Value).Kind() != reflect.Slice || d.OldValue == nil || len(d.Args) <= 2 {
		//then := time.Now()
		n := 1
		for i := 0; i < n; i++ {
			b.FullUpdate(d)
		}
		//println(time.Now().Sub(then).Seconds() / float64(n))
		return
	} else {
		if d.Args[2] == "mode_s" {
			performChange(&listChanger{b, &d}, reflect.ValueOf(d.OldValue), reflect.ValueOf(d.Value))
		} else {
			return fmt.Errorf("Invalid value for argument 3 to the each binder.")
		}
	}

	return
}

// IfBinder keeps or remove an element according to a boolean value.
//
// Usage:
//	#if="BooleanExpression"
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
//	#ifn="BooleanExpression"
type UnlessBinder struct {
	*IfBinder
}

func (b *UnlessBinder) Update(d DomBind) (err error) {
	d.Value = !(d.Value.(bool))
	b.IfBinder.Update(d)

	return
}
func (b *UnlessBinder) BindInstance() DomBinder { return &UnlessBinder{&IfBinder{}} }
