package bind

import (
	"fmt"

	"github.com/phaikawl/wade/dom"
)

type ModelUpdateFn func(value string)

// DomBinder is the common interface for Dom binders.
type DomBinder interface {
	// Update is called whenever the model's field changes, to perform
	// dom updating, like setting the html content or setting
	// an html attribute for the elem
	Update(DomBind) error

	// Bind is similar to Update, but is called only once at the start, when
	// the bind is being processed
	Bind(DomBind) error

	// Watch is used in 2-way binders, here we perform listening the html element for changes
	// and updates the model field accordingly
	Watch(DomBind, ModelUpdateFn) error

	// BindInstance is useful for binders that need to save some data for each
	// separate element. This method returns an instance of the binder to be used.
	BindInstance() DomBinder
}

type DomBind struct {
	Elem     dom.Selection
	OldValue interface{}
	Value    interface{}
	Args     []string

	binding *Binding
	scope   *Scope
}

// Bind performs a bind.
// Arguments: elem is the element to perform on; m is the map of values; once is whether
// bind only once (no watching) or watch for changes; bindRoot is whether we bind the root element
// or not
func (d DomBind) Bind(elem dom.Selection, m map[string]interface{}, once bool, bindRoot bool) {
	s := newModelScope(m)
	s.merge(d.scope)

	d.binding.bindWithScope(elem, s, once, bindRoot, d.Elem)
}

// Banish performs real removal of the element, no binding for it and its descendants will be performed
func (d DomBind) Banish(elem dom.Selection) {
	if e, ok := elem.(drmElem); ok {
		e.Selection.Remove()
	} else {
		e.Remove()
	}
}

// ProduceOutputs is a convenient method which performs call Bind on the element,
// producing values with name specified in names
// and values specified in outputs accordingly
func (d DomBind) ProduceOutputs(elem dom.Selection, once bool, names []string, outputs ...interface{}) error {
	m := make(map[string]interface{})
	if len(outputs) == len(names) {
		for i, output := range names {
			m[output] = outputs[i]
		}

		d.Bind(elem, m, once, true)
	} else {
		return fmt.Errorf("name list length is %v but %v outputs are specified.", len(names), len(outputs))
	}

	return nil
}

// BaseBinder provides a base to be embedded so that we will not have to write empty
// implementation for the unneeded methods
type BaseBinder struct{}

func (b *BaseBinder) Bind(d DomBind) error {
	return nil
}
func (b *BaseBinder) Update(d DomBind) error {
	return nil
}
func (b *BaseBinder) Watch(d DomBind, ufn ModelUpdateFn) error {
	return nil
}
