package bind

import (
	"fmt"
	"reflect"
)

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

func (b *bindScope) evaluateRec(e *expr, old uintptr, repl reflect.Value) (v reflect.Value, blist []bindable, err error) {
	err = nil
	blist = make([]bindable, 0)

	wrapped := false
	watch := false

	switch e.name[0] {
	case '$':
		e.name = e.name[1:]
		watch = true

	case '@':
		wrapped = true
		e.name = e.name[1:]

	default:
		if litVal, isLiteral, er := parseLiteralExpr(e.name); isLiteral {
			if er != nil {
				err = er
				return
			}

			v = reflect.ValueOf(litVal)
			return
		}
	}

	args := make([]reflect.Value, len(e.args))
	for i, e := range e.args {
		var blc []bindable
		args[i], blc, err = b.evaluateRec(e, old, repl)
		if err != nil {
			return
		}

		blist = append(blist, blc...)
	}

	sym, err := b.scope.lookup(e.name)
	if err != nil {
		return
	}

	switch e.typ {
	case ValueExpr:
		v, err = sym.value()
		if old != 0 && old == v.UnsafeAddr() {
			v = repl
		}

		if watch {
			blist = append(blist, sym.(bindable))
		}

	case CallExpr:
		if wrapped {
			v = reflect.ValueOf(func() {
				_, er := sym.call(args, true)
				if er != nil {
					panic(er)
				}
			})

			return
		}

		v, err = sym.call(args, false)

		if watch {
			err = fmt.Errorf("Watching a function call is not supported")
			return
		}
	}

	if err != nil {
		return
	}

	return
}

// evaluate evaluates the bind string, returns the needed information for binding
func (b *bindScope) evaluate(bstr string) (calcRoot *expr, blist []bindable, value interface{}, err error) {
	calcRoot, err = parse(bstr)
	if err != nil {
		return
	}

	blist, value, err = b.evaluatePart(calcRoot)
	return
}

func (b *bindScope) evaluatePart(calcRoot *expr) (blist []bindable, value interface{}, err error) {
	var v reflect.Value
	v, blist, err = b.evaluateRec(calcRoot, 0, reflect.ValueOf(nil))
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
