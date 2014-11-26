package core

import (
	"fmt"

	"github.com/phaikawl/wade/scope"
)

type ModelUpdateFn func(value string)

// Binder is the common interface for binders.
type Binder interface {
	// Check the number of arguments and return whether it's legal or not
	CheckArgsNo(argsNo int) (bool, string)

	// Update is called whenever the DOM is rendered/rerendered
	Update(DomBind)

	// Bind is called once when a bind is executed
	Bind(DomBind)
}

type TwoWayBinder interface {
	// Listen is used in 2-way binders, here we perform listening the html element for changes
	// and updates the model field accordingly
	Listen(DomBind, ModelUpdateFn)
}

type DomBind struct {
	Node  *VNode
	Value interface{}
	Args  []string

	BindName string
	binding  *Binding
	scope    scope.Scope
}

// Bind performs a bind.
func (d DomBind) Bind(node *VNode, m interface{}) {
	d.binding.bindWithScope(node, scope.NewScope(m).Merge(d.scope))
}

func (d DomBind) RemoveBind(node *VNode) {
	for i := range node.Binds {
		if node.Binds[i].Name == d.BindName {
			node.Binds = append(node.Binds[:i], node.Binds[i+1:]...)
			return
		}
	}
}

// ProduceOutputs is a convenient method which performs call Bind on the element,
// producing values with name specified in names
// and values specified in outputs accordingly
func (d DomBind) ProduceOutputs(node *VNode, names []string, outputs ...interface{}) {
	m := make(map[string]interface{})
	if len(outputs) == len(names) {
		for i, output := range names {
			m[output] = outputs[i]
		}

		d.Bind(node, m)
	} else {
		panic(fmt.Errorf("name list length is %v but %v outputs are given.",
			len(names), len(outputs)))
	}

	return
}

// BaseBinder provides a base to be embedded so that we will not have to write empty
// implementation for the unneeded methods
type BaseBinder struct{}

func (b BaseBinder) Bind(d DomBind) {
}

func (b BaseBinder) Update(d DomBind) {
}

func (b BaseBinder) Listen(d DomBind, ufn ModelUpdateFn) {
}

func (b BaseBinder) CheckArgsNo(n int) (bool, string) {
	return n == 0, "0"
}
