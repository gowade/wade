package core

import (
	"reflect"

	. "github.com/phaikawl/wade/scope"
)

type (
	bindScope struct {
		Scope
	}
)

func (b bindScope) evaluateRec(e *expr) (v interface{}, err error) {
	wrapped := false

	switch e.name[0] {
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

	var sym ScopeSymbol
	if e.preque != nil {
		var preVal interface{}
		preVal, err = b.evaluateRec(e.preque)
		if err != nil {
			return
		}

		sym, err = NewScope(preVal).Lookup(e.name[1:])
	} else {
		sym, err = b.Lookup(e.name)
	}

	if err != nil {
		return
	}

	var rv reflect.Value
	switch e.typ {
	case ValueExpr:
		rv, err = sym.Value()
		v = rv.Interface()

	case CallExpr:
		args := make([]reflect.Value, len(e.args))
		for i, e := range e.args {
			var av interface{}
			av, err = b.evaluateRec(e)
			args[i] = reflect.ValueOf(av)
			if err != nil {
				return
			}
		}

		if wrapped {
			v = func() {
				_, er := sym.Call(args, true)
				if er != nil {
					panic(er)
				}
			}

			return
		}

		rv, err = sym.Call(args, false)
		if rv.IsValid() && rv.CanInterface() {
			v = rv.Interface()
		}
	}

	if err != nil {
		return
	}

	return
}

// evaluate evaluates the bind string, returns the needed information for binding
func (b bindScope) evaluate(bstr string) (calcRoot *expr, value interface{}, err error) {
	calcRoot, err = parse(bstr)
	if err != nil {
		return
	}

	value, err = b.evaluatePart(calcRoot)
	return
}

func (b bindScope) evaluatePart(calcRoot *expr) (value interface{}, err error) {
	value, err = b.evaluateRec(calcRoot)
	if err != nil {
		return
	}

	return
}
