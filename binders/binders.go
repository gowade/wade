package binders

import (
	"fmt"
	"reflect"

	"github.com/phaikawl/wade/core"
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/utils"
)

var (
	binders = map[string]core.Binder{
		"value": ValueBinder{},
		"on":    &EventBinder{},
		"range": &RangeBinder{},
		"if":    &IfBinder{},
		"ifn":   IfnBinder{&IfBinder{}},
		"class": ClassBinder{},
	}
)

func Install(b *core.Binding) {
	for name, binder := range binders {
		b.RegisterBinder(name, binder)
	}
}

// ValueBinder is a 2-way binder that binds an element's value attribute to a value.
// Meant to be used for <input>, <textarea> and <select>.
//
// Parameters: A list of event names to listen to the dom, updating data from
// the real dom to the model.
//
// Usage:
//	#value(...events)="Expression"
type ValueBinder struct{ core.BaseBinder }

func (b ValueBinder) CheckArgsNo(n int) (bool, string) {
	return true, "any"
}

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
// The Expression is evaluated like any other expressions,
// if the expression is a function call,
// the value evaluated is the return value, not the function call.
type EventBinder struct {
	core.BaseBinder
	evt *dom.Event
}

func (b EventBinder) CheckArgsNo(n int) (bool, string) {
	return n == 1, "1"
}

func (b *EventBinder) BeforeBind(s core.ScopeAdder) {
	s.AddValues(map[string]interface{}{
		"$event": b.evt,
	})
}

func (b EventBinder) NewInstance() core.Binder {
	return &EventBinder{evt: new(dom.Event)}
}

func (b *EventBinder) Bind(d core.DomBind) {
	fni := d.Value
	if fni == nil {
		panic(fmt.Errorf(`Event must be bound to a handler function, not a nil.
		Please use the '@' syntax to wrap a function call.`))
	}

	handler, ok := fni.(func())

	if !ok {
		panic(fmt.Errorf(`Wrong type %v for EventBinder's bind target,
		must be a function of type func()`,
			reflect.TypeOf(fni).String()))
	}

	evtname := d.Args[0]
	d.Node.SetAttr("on"+evtname, func(evt dom.Event) {
		*b.evt = evt
		//gopherjs:blocking
		handler()
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

func (b RangeBinder) NewInstance() core.Binder {
	return &RangeBinder{}
}

func (b *RangeBinder) Bind(d core.DomBind) {
	b.prototype = *d.Node
	d.RemoveBind(&b.prototype)

	*d.Node = core.VPrep(core.VNode{
		Type: core.GroupNode,
		Data: "range",
	})

	return
}

func (b RangeBinder) CheckArgsNo(n int) (bool, string) {
	return n == 2, "2"
}

func (b RangeBinder) Update(d core.DomBind) {
	val := reflect.ValueOf(d.Value)
	if val.Kind() != reflect.Slice {
		panic(fmt.Errorf(`Wrong type of expression for "range" binder, it must be a slice.`))
	}

	d.Node.Children = make([]core.VNode, val.Len())
	for i := 0; i < val.Len(); i++ {
		d.Node.Children[i] = b.prototype.Clone()

		d.BindOutputs(&d.Node.Children[i], d.Args[:2], i, val.Index(i).Interface())
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

func (b IfBinder) NewInstance() core.Binder {
	return &IfBinder{}
}

func (b *IfBinder) Bind(d core.DomBind) {
	b.nodeType = d.Node.Type
	return
}

func (b *IfBinder) Update(d core.DomBind) {
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

func (b IfnBinder) NewInstance() core.Binder {
	return IfnBinder{&IfBinder{}}
}
