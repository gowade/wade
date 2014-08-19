package bind

import (
	"fmt"
	"reflect"
)

type (
	scope struct {
		symTables []symbolTable
	}

	scopeSymbol interface {
		value() (reflect.Value, error)
		call([]reflect.Value) (reflect.Value, error)
	}

	symbolTable interface {
		lookup(symbol string) (scopeSymbol, bool)
	}

	funcSymbol struct {
		name string
		fn   reflect.Value
	}

	modelSymbolTable struct {
		model reflect.Value
	}

	modelFieldSymbol struct {
		name string
		eval *objEval
	}
)

func newScope() *scope {
	return &scope{make([]symbolTable, 0)}
}

func (s *scope) lookup(symbol string) (sym scopeSymbol, err error) {
	for _, st := range s.symTables {
		var ok bool
		sym, ok = st.lookup(symbol)
		if ok {
			return
		}
	}

	err = fmt.Errorf(`Unable to find symbol "%v" in the scope`, symbol)
	return
}

func (s *scope) merge(target *scope) {
	for _, st := range target.symTables {
		s.symTables = append(s.symTables, st)
	}
}

type mapSymbolTable struct {
	m map[string]scopeSymbol
}

func (st mapSymbolTable) lookup(symbol string) (sym scopeSymbol, ok bool) {
	sym, ok = st.m[symbol]
	return
}

func (st mapSymbolTable) registerFunc(name string, fn interface{}) {
	st.m[name] = newFuncSymbol(name, fn)
}

func newFuncSymbol(name string, fn interface{}) funcSymbol {
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		panic(fmt.Sprintf(`Can't create funcSymbol "%v" from a non-function.`, name))
	}

	if fnType.NumOut() > 1 {
		panic(fmt.Sprintf(`"%v": funcSymbol cannot have more than 1 return value.`, name))
	}

	return funcSymbol{name, reflect.ValueOf(fn)}
}

func (fs funcSymbol) value() (reflect.Value, error) {
	return fs.fn, nil
}

func (fs funcSymbol) call(args []reflect.Value) (v reflect.Value, err error) {
	v, err = callFunc(fs.fn, args)
	if err != nil {
		err = fmt.Errorf(`"%v": %v`, fs.name, err.Error())
	}
	return
}

func helpersSymbolTable(helpers map[string]interface{}) mapSymbolTable {
	m := make(map[string]scopeSymbol)
	for name, helper := range helpers {
		m[name] = newFuncSymbol(name, helper)
	}

	return mapSymbolTable{m}
}

func (mf modelFieldSymbol) bindObj() *objEval {
	return mf.eval
}

func (mf modelFieldSymbol) value() (v reflect.Value, err error) {
	return mf.eval.fieldRefl, nil
}

func (mf modelFieldSymbol) call(args []reflect.Value) (v reflect.Value, err error) {
	if mf.eval.fieldRefl.Kind() != reflect.Func {
		err = fmt.Errorf(`Cannot call "%v", it's not a method.`, mf.name)
		return
	}

	v, err = callFunc(mf.eval.fieldRefl, args)
	if err != nil {
		err = fmt.Errorf(`"%v": %v`, mf.name, err.Error())
	}
	return
}

func (st modelSymbolTable) lookup(symbol string) (sym scopeSymbol, ok bool) {
	if st.model.Kind() == reflect.Ptr && st.model.IsNil() {
		ok = false
		return
	}

	var eval *objEval
	eval, ok = evaluateObjField(symbol, st.model)
	if ok {
		sym = modelFieldSymbol{symbol, eval}
	}

	return
}

func newModelScope(model interface{}) *scope {
	stl := []symbolTable{}
	if model != nil {
		stl = append(stl, modelSymbolTable{reflect.ValueOf(model)})
	}
	return &scope{stl}
}
