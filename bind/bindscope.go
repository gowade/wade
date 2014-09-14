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

// evaluateRec recursively evaluates the parsed expressions and return the result value, it also
func (b *bindScope) evaluateRec(e *expr, watches []token) (v reflect.Value, err error) {
	err = nil

	wrapped := false

	switch e.name[0] {
	case '$':
		e.name, err = parseDollarExpr(e.name, watches)
		if err != nil {
			return
		}

		var sym scopeSymbol
		sym, err = b.scope.lookup(e.name)
		if err != nil {
			return
		}
		v, err = sym.value()
		return

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
		args[i], err = b.evaluateRec(e, watches)
		if err != nil {
			return
		}
	}

	sym, err := b.scope.lookup(e.name)
	if err != nil {
		return
	}

	switch e.typ {
	case ValueExpr:
		v, err = sym.value()
	case CallExpr:
		if wrapped {
			v = reflect.ValueOf(func() {
				_, er := sym.call(args)
				if er != nil {
					panic(er)
				}
			})

			return
		}
		v, err = sym.call(args)
	}

	if err != nil {
		return
	}

	return
}

// evaluate evaluates the bind string, returns the needed information for binding
func (b *bindScope) evaluate(bstr string) (calcRoot *expr, blist []bindable, watches []token, value interface{}, err error) {
	watches, calcRoot, err = parse(bstr)
	if err != nil {
		return
	}

	blist, value, err = b.evaluatePart(watches, calcRoot)
	return
}

func (b *bindScope) evaluatePart(watches []token, calcRoot *expr) (blist []bindable, value interface{}, err error) {
	blist = make([]bindable, len(watches))
	for i, watch := range watches {
		var sym scopeSymbol
		sym, err = b.scope.lookup(watch.v)
		if err != nil {
			return
		}
		var ok bool
		if blist[i], ok = sym.(bindable); !ok {
			err = fmt.Errorf("Cannot watch unaddressable value %v", watch.v)
			return
		}
	}

	var v reflect.Value
	v, err = b.evaluateRec(calcRoot, watches)
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
