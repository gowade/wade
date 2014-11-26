package scope

import (
	"fmt"
	"reflect"
)

type (
	Scope struct {
		symTables []SymbolTable
	}

	ScopeSymbol interface {
		Value() (reflect.Value, error)
		Call(args []reflect.Value, async bool) (reflect.Value, error)
	}

	SymbolTable interface {
		Lookup(symbol string) (ScopeSymbol, bool, error)
	}

	ModelSymbolTable struct {
		model reflect.Value
	}

	FieldSymbol struct {
		name string
		eval *ObjEval
	}
)

func (s Scope) Lookup(symbol string) (sym ScopeSymbol, err error) {
	for _, st := range s.symTables {
		var ok bool
		sym, ok, err = st.Lookup(symbol)
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

func (s Scope) LookupValue(symbol string) (value interface{}, err error) {
	sym, err := s.Lookup(symbol)
	if err != nil {
		return
	}

	v, err := sym.Value()
	if err != nil {
		return
	}

	value = v.Interface()
	return
}

func (s Scope) Merge(target Scope) Scope {
	return Scope{append(s.symTables, target.symTables...)}
}

func (fs FieldSymbol) BindObj() *ObjEval {
	return fs.eval
}

func (fs FieldSymbol) Value() (v reflect.Value, err error) {
	return fs.eval.FieldRefl, nil
}

func (fs FieldSymbol) Call(args []reflect.Value, async bool) (v reflect.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in sym.Call()", r)
			err = fmt.Errorf("%v", r)
		}
	}()

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

func (st ModelSymbolTable) Lookup(symbol string) (sym ScopeSymbol, ok bool, err error) {
	if st.model.Kind() == reflect.Ptr && st.model.IsNil() {
		ok = false
		return
	}

	var eval *ObjEval
	eval, ok, err = EvaluateObjField(symbol, st.model)
	if ok {
		sym = FieldSymbol{symbol, eval}
	}

	return
}

func NewScope(models ...interface{}) Scope {
	stl := []SymbolTable{}
	for _, model := range models {
		if model != nil {
			stl = append(stl, ModelSymbolTable{reflect.ValueOf(model)})
		}
	}
	return Scope{stl}
}

func (s Scope) Len() int {
	return len(s.symTables)
}

func NewModelSymbolTable(model interface{}) ModelSymbolTable {
	return ModelSymbolTable{reflect.ValueOf(model)}
}
