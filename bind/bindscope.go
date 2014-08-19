package bind

import "reflect"

type (
	objEval struct {
		fieldRefl reflect.Value
		modelRefl reflect.Value
		field     string
	}

	bindable interface {
		bindObj() *objEval
	}

	bindScope struct {
		scope *scope
	}
)

// evaluateRec recursively evaluates the parsed expressions and return the result value, it also
func (b *bindScope) evaluateRec(e *expr) (v reflect.Value, blist []bindable, err error) {
	err = nil
	blist = make([]bindable, 0)

	litVal, isLiteral, er := parseExpr(e.name)
	if er != nil {
		err = er
		return
	}
	if isLiteral {
		v = reflect.ValueOf(litVal)
		return
	}

	args := make([]reflect.Value, len(e.args))
	for i, e := range e.args {
		var cblist []bindable
		args[i], cblist, err = b.evaluateRec(e)
		if err != nil {
			return
		}

		blist = append(blist, cblist...)
	}

	sym, err := b.scope.lookup(e.name)
	if err != nil {
		return
	}

	switch e.typ {
	case ValueExpr:
		v, err = sym.value()
	case CallExpr:
		v, err = sym.call(args)
	}

	if err != nil {
		return
	}

	if mf, ok := sym.(bindable); ok {
		blist = append(blist, mf)
	}
	return
}

// evaluateBindstring evaluates the bind string, returns the needed information for binding
func (b *bindScope) evaluate(bstr string) (root *expr, blist []bindable, value interface{}, err error) {
	root, err = parse(bstr)
	if err != nil {
		return
	}

	var v reflect.Value
	v, blist, err = b.evaluateRec(root)
	if err != nil {
		return
	}
	if v.IsValid() && v.CanInterface() {
		value = v.Interface()
	}
	return
}

func (b *bindScope) clone() *bindScope {
	scope := newScope()
	scope.merge(b.scope)
	return &bindScope{scope}
}
