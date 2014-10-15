package bind

import (
	"fmt"
	"reflect"
)

type (
	Scope struct {
		symTables []symbolTable
	}

	ScopeSymbol interface {
		Value() (reflect.Value, error)
		call(args []reflect.Value, async bool) (reflect.Value, error)
	}

	symbolTable interface {
		lookup(symbol string) (ScopeSymbol, bool, error)
	}

	helpersSymbolTable struct {
		m map[string]ScopeSymbol
	}

	funcSymbol struct {
		name string
		fn   reflect.Value
	}

	modelSymbolTable struct {
		model reflect.Value
	}

	fieldSymbol struct {
		name string
		eval *ObjEval
	}
)

func newScope() *Scope {
	return &Scope{make([]symbolTable, 0)}
}

func (s *Scope) Lookup(symbol string) (sym ScopeSymbol, err error) {
	for _, st := range s.symTables {
		var ok bool
		sym, ok, err = st.lookup(symbol)
		if err != nil {
			return
		}

		if ok {
			return
		}
	}

	err = fmt.Errorf(`Unable to find symbol "%v" in the Scope`, symbol)
	return
}

func (s *Scope) merge(target *Scope) {
	for _, st := range target.symTables {
		s.symTables = append(s.symTables, st)
	}
}

func (st helpersSymbolTable) lookup(symbol string) (sym ScopeSymbol, ok bool, err error) {
	sym, ok = st.m[symbol]
	return
}

func (st helpersSymbolTable) registerFunc(name string, fn interface{}) {
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

func (fs funcSymbol) Value() (reflect.Value, error) {
	return fs.fn, nil
}

func (fs funcSymbol) call(args []reflect.Value, async bool) (v reflect.Value, err error) {
	if async {
		go func() {
			fs.fn.Call(args)
		}()
		return
	}

	defer func() {
		if r := recover(); r != nil {
			str, ok := r.(string)
			err = fmt.Errorf(str)
			if !ok {
				err = r.(error)
			}
		}
	}()
	v, err = callFunc(fs.fn, args)
	if err != nil {
		err = fmt.Errorf(`"%v": %v`, fs.name, err.Error())
	}
	return
}

func newHelpersSymbolTable(helpers map[string]interface{}) helpersSymbolTable {
	m := make(map[string]ScopeSymbol)
	for name, helper := range helpers {
		m[name] = newFuncSymbol(name, helper)
	}

	return helpersSymbolTable{m}
}

func (fs fieldSymbol) bindObj() *ObjEval {
	return fs.eval
}

func (fs fieldSymbol) Value() (v reflect.Value, err error) {
	return fs.eval.FieldRefl, nil
}

func (fs fieldSymbol) call(args []reflect.Value, async bool) (v reflect.Value, err error) {
	if fs.eval.FieldRefl.Kind() != reflect.Func {
		err = fmt.Errorf(`Cannot call "%v", it's not a method or a function.`, fs.name)
		return
	}

	if async {
		fs.eval.FieldRefl.Call(args)
		return
	}

	v, err = callFunc(fs.eval.FieldRefl, args)
	if err != nil {
		err = fmt.Errorf(`"%v": %v`, fs.name, err.Error())
	}

	return
}

func (st modelSymbolTable) lookup(symbol string) (sym ScopeSymbol, ok bool, err error) {
	if st.model.Kind() == reflect.Ptr && st.model.IsNil() {
		ok = false
		return
	}

	var eval *ObjEval
	eval, ok, err = evaluateObjField(symbol, st.model)
	if ok {
		sym = fieldSymbol{symbol, eval}
	}

	return
}

func newModelScope(model interface{}) *Scope {
	stl := []symbolTable{}
	if model != nil {
		stl = append(stl, modelSymbolTable{reflect.ValueOf(model)})
	}
	return &Scope{stl}
}
