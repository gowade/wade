package core

import (
	"fmt"
	"reflect"

	"github.com/phaikawl/wade/core"
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/utils"
)

var (
	Binders = map[string]core.Binder{
		"value": ValueBinder{},
		"on":    EventBinder{},
		"range": &RangeBinder{},
		"if":    &IfBinder{},
		"ifn":   IfnBinder{&IfBinder{}},
		"class": ClassBinder{},
	}
)

// ValueBinder is a 2-way binder that binds an element's value attribute to a value.
// Meant to be used for <input>, <textarea> and <select>.
//
// Parameters: A list of event names to listen to the dom, updating data from
// the real dom to the model.
//
// Usage:
//	#value(...events)="Expression"
type ValueBinder struct{ core.BaseBinder }

// Update sets the element's value attribute to a new value
func (b ValueBinder) Update(d core.DomBind) {
	d.Node.SetAttr("value", utils.ToString(d.Value))
	return
}

func (b ValueBinder) Listen(d core.DomBind, ufn core.ModelUpdateFn) {
	tagname := d.Node.TagName()
	if tagname != "input" && tagname != "textarea" && tagname != "select" {
		panic(fmt.Errorf("Can only watch for changes on html input, textarea and select"))
	}

	for _, event := range d.Args {
		d.Node.SetAttr("on"+event, func(evt dom.Event) {
			val, _ := evt.Target().Prop("value")
			ufn(val.(string))
		})
	}
}

// ClassBinder toggles a class based on a boolean value.
//
// Usage:
//	#class(className)="Expression"
type ClassBinder struct{ core.BaseBinder }

func (b ClassBinder) CheckArgsNo(n int) (bool, string) {
	return n == 1, "1"
}

func (b ClassBinder) Update(d core.DomBind) {
	class := d.Args[0]
	d.Node.SetClass(class, d.Value.(bool))
}

// EventBinder binds an element's event to a function.
//
// Usage:
//	#on(event)="Expression"
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
type EventBinder struct{ core.BaseBinder }

func (b EventBinder) CheckArgsNo(n int) (bool, string) {
	return n == 1, "1"
}

func (b EventBinder) Bind(d core.DomBind) {
	fni := d.Value
	if fni == nil {
		panic(fmt.Errorf(`Event must be bound to a handler function of type
		func(dom.Event), not a nil.
		Note that if you want to call a function,
		please wrap it or use the '@' syntax.`))
	}

	handler1, ok1 := fni.(func(dom.Event))

	if !ok1 {
		panic(fmt.Errorf(`Wrong type %v for EventBinder's bind target,
		must be a function of type func(dom.Event)`,
			reflect.TypeOf(fni).String()))
	}

	evtname := d.Args[0]
	d.Node.SetAttr("on"+evtname, func(evt dom.Event) {
		go func() {
			//gopherjs:blocking
			handler1(evt)
		}()
	})
}

// RangeBinder repeats an element according to a slice.
//
// It takes 2 arguments: the name to bind the index to, and the name to bind the value to.
// Usage:
//  #each=(indexBindName, valueBindName)
//
// Example:
//  #each(index,product)="Products"
//  #each(_,product)="Products"
type RangeBinder struct {
	core.BaseBinder
	prototype core.VNode
}

func (b *RangeBinder) Bind(d core.DomBind) {
	b.prototype = *d.Node
	d.RemoveBind(&b.prototype)

	*d.Node = core.V(core.GroupNode, "range", core.NoAttr(), core.NoBind(), []core.VNode{})

	return
}

func (b RangeBinder) Update(d core.DomBind) {
	val := reflect.ValueOf(d.Value)
	if val.Kind() != reflect.Slice {
		panic(fmt.Errorf(`Wrong type of expression for "range" binder, it must be a slice.`))
	}

	d.Node.Children = make([]core.VNode, val.Len())
	for i := 0; i < val.Len(); i++ {
		d.Node.Children[i] = b.prototype.Clone()

		d.ProduceOutputs(&d.Node.Children[i], d.Args[:2], i, val.Index(i).Interface())
	}

	return
}

// IfBinder keeps or remove an element according to a boolean value.
//
// Usage:
//	#if="BooleanExpression"
type IfBinder struct {
	core.BaseBinder
	nodeType core.NodeType
}

func (b *IfBinder) Bind(d core.DomBind) {
	b.nodeType = d.Node.Type
	return
}

func (b IfBinder) Update(d core.DomBind) {
	if !d.Value.(bool) {
		d.Node.Type = core.DeadNode
	} else {
		d.Node.Type = b.nodeType
	}

	return
}

// IfnBinder is the reverse of IfBinder.
//
// Usage:
//	#ifn="BooleanExpression"
type IfnBinder struct {
	*IfBinder
}

func (b IfnBinder) Update(d core.DomBind) {
	d.Value = !(d.Value.(bool))
	b.IfBinder.Update(d)

	return
}
