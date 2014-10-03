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

	barray struct {
		slice []bindable
		size  int
	}
)

func (a *barray) add(value bindable) {
	a.slice[a.size] = value
	a.size++
}

func (b *bindScope) evaluateRec(e *expr, blist *barray, old uintptr, repl interface{}) (v interface{}, err error) {
	err = nil

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

			v = litVal
			return
		}
	}

	var sym scopeSymbol
	if e.preque != nil {
		var preVal interface{}
		preVal, err = b.evaluateRec(e.preque, blist, old, repl)
		if err != nil {
			return
		}

		sym, err = newModelScope(preVal).lookup(e.name[1:])
	} else {
		sym, err = b.scope.lookup(e.name)
	}

	if err != nil {
		return
	}

	var rv reflect.Value
	switch e.typ {
	case ValueExpr:
		rv, err = sym.value()
		if old != 0 && old == rv.UnsafeAddr() {
			v = repl
		} else {
			v = rv.Interface()
		}

		if watch {
			blist.add(sym.(bindable))
		}

	case CallExpr:
		args := make([]reflect.Value, len(e.args))
		for i, e := range e.args {
			var av interface{}
			av, err = b.evaluateRec(e, blist, old, repl)
			args[i] = reflect.ValueOf(av)
			if err != nil {
				return
			}
		}

		if wrapped {
			v = func() {
				_, er := sym.call(args, true)
				if er != nil {
					panic(er)
				}
			}

			return
		}

		rv, err = sym.call(args, false)
		if rv.IsValid() && rv.CanInterface() {
			v = rv.Interface()
		}

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
func (b *bindScope) evaluate(bstr string) (calcRoot *expr, blist *barray, value interface{}, err error) {
	var nwatches int
	calcRoot, nwatches, err = parse(bstr)
	if err != nil {
		return
	}

	blist, value, err = b.evaluatePart(calcRoot, nwatches)
	return
}

func (b *bindScope) evaluatePart(calcRoot *expr, nwatches int) (blist *barray, value interface{}, err error) {
	blist = &barray{make([]bindable, nwatches), 0}
	value, err = b.evaluateRec(calcRoot, blist, 0, nil)
	if err != nil {
		return
	}

	return
}

func (b *bindScope) clone() *bindScope {
	scope := newScope()
	scope.merge(b.scope)
	return &bindScope{scope}
}
