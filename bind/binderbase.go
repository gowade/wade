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
	Update(DomBind)

	// Bind is similar to Update, but is called only once at the start, when
	// the bind is being processed
	Bind(DomBind)

	// Watch is used in 2-way binders, it watches the html element for changes
	// and updates the model field accordingly
	Watch(elem dom.Selection, updateFn ModelUpdateFn)

	// BindInstance is useful for binders that need to save some data for each
	// separate element. This method returns an instance of the binder to be used.
	BindInstance() DomBinder
}

type DomBind struct {
	Elem    dom.Selection
	Value   interface{}
	Args    []string
	outputs []string

	binding  *Binding
	scope    *scope
	metadata string
}

func (d DomBind) bind(elem dom.Selection, model interface{}, once bool, bindrelem bool) {
	s := newModelScope(model)
	s.merge(d.scope)
	d.binding.bindWithScope(elem, once, bindrelem, s, nil)
}

func (d DomBind) RemoveBinding(elem dom.Selection) {
	preventAllBinding(elem)
}

func (d DomBind) ProduceOutputs(elem dom.Selection, optional bool, once bool, outputs ...interface{}) {
	m := make(map[string]interface{})
	if len(outputs) == len(d.outputs) {
		for i, output := range d.outputs {
			m[output] = outputs[i]
		}

		d.bind(elem, m, once, true)
	} else {
		if !optional || len(outputs) != 0 {
			panic(fmt.Errorf("Wrong output specification for `%v`: there must be %v outputs instead of %v.",
				d.metadata, len(d.outputs), len(outputs)))
		}
	}
}

func (d DomBind) Panic(msg string) {
	panic(d.metadata + ": " + msg)
}

// BaseBinder provides the base so that binders will not have to provide empty
// implement for the methods
type BaseBinder struct{}

func (b *BaseBinder) Bind(d DomBind) {
}
func (b *BaseBinder) Update(d DomBind)                            {}
func (b *BaseBinder) Watch(elem dom.Selection, ufn ModelUpdateFn) {}
