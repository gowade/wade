package core

import (
	"fmt"
	. "github.com/phaikawl/wade/scope"
)

type ModelUpdateFn func(value string)

// Binder is the common interface for binders.
type Binder interface {
	ArgsFn() interface{}

	// Update is called whenever the model's field changes, to perform
	// dom updating, like setting the html content or setting
	// an html attribute for the elem
	Update(DomBind) error

	// Bind is similar to Update, but is called only once at the start, when
	// the bind is being processed
	Bind(DomBind) error

	// BindInstance is useful for binders that need to save some data for each
	// separate element. This method returns an instance of the binder to be used.
	BindInstance() Binder
}

type TwoWayBinder interface {
	// Listen is used in 2-way binders, here we perform listening the html element for changes
	// and updates the model field accordingly
	Listen(DomBind, ModelUpdateFn) error
}

type DomBind struct {
	Node     *VNode
	OldValue interface{}
	Value    interface{}
	Args     []string

	binding *Binding
	scope   *Scope
}

// Bind performs a bind.
func (d DomBind) Bind(node *VNode, m map[string]interface{}) {
	s := NewModelScope(m)
	s.Merge(d.scope)

	d.binding.bindWithScope(node, s)
}

// ProduceOutputs is a convenient method which performs call Bind on the element,
// producing values with name specified in names
// and values specified in outputs accordingly
func (d DomBind) ProduceOutputs(node *VNode, names []string, outputs ...interface{}) error {
	m := make(map[string]interface{})
	if len(outputs) == len(names) {
		for i, output := range names {
			m[output] = outputs[i]
		}

		d.Bind(node, m)
	} else {
		return fmt.Errorf("name list length is %v but %v outputs are specified.", len(names), len(outputs))
	}

	return nil
}

// BaseBinder provides a base to be embedded so that we will not have to write empty
// implementation for the unneeded methods
type BaseBinder struct{}

func (b BaseBinder) Bind(d DomBind) error {
	return nil
}

func (b BaseBinder) Update(d DomBind) error {
	return nil
}

func (b BaseBinder) Listen(d DomBind, ufn ModelUpdateFn) error {
	return nil
}

func (b BaseBinder) ArgsFn() interface{} {
	return func() {}
}
