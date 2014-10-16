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

	HelpersSymbolTable struct {
		m map[string]ScopeSymbol
	}

	FuncSymbol struct {
		name string
		fn   reflect.Value
	}

	ModelSymbolTable struct {
		model reflect.Value
	}

	FieldSymbol struct {
		name string
		eval *ObjEval
	}
)

func NewScope(symtables []SymbolTable) *Scope {
	return &Scope{symtables}
}

func (s *Scope) Lookup(symbol string) (sym ScopeSymbol, err error) {
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

func (s *Scope) LookupValue(symbol string) (value interface{}, err error) {
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

func (s *Scope) Merge(target *Scope) {
	s.symTables = append(s.symTables, target.symTables...)
}

func (s *Scope) AddSymTables(tables ...SymbolTable) {
	s.symTables = append(s.symTables, tables...)
}

func (st HelpersSymbolTable) Lookup(symbol string) (sym ScopeSymbol, ok bool, err error) {
	sym, ok = st.m[symbol]
	return
}

func (st HelpersSymbolTable) RegisterFunc(name string, fn interface{}) {
	st.m[name] = newFuncSymbol(name, fn)
}

func newFuncSymbol(name string, fn interface{}) FuncSymbol {
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		panic(fmt.Sprintf(`Can't create FuncSymbol "%v" from a non-function.`, name))
	}

	if fnType.NumOut() > 1 {
		panic(fmt.Sprintf(`"%v": FuncSymbol cannot have more than 1 return value.`, name))
	}

	return FuncSymbol{name, reflect.ValueOf(fn)}
}

func (fs FuncSymbol) Value() (reflect.Value, error) {
	return fs.fn, nil
}

func (fs FuncSymbol) Call(args []reflect.Value, async bool) (v reflect.Value, err error) {
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

func NewHelpersSymbolTable(helpers map[string]interface{}) HelpersSymbolTable {
	m := make(map[string]ScopeSymbol)
	for name, helper := range helpers {
		m[name] = newFuncSymbol(name, helper)
	}

	return HelpersSymbolTable{m}
}

func (fs FieldSymbol) BindObj() *ObjEval {
	return fs.eval
}

func (fs FieldSymbol) Value() (v reflect.Value, err error) {
	return fs.eval.FieldRefl, nil
}

func (fs FieldSymbol) Call(args []reflect.Value, async bool) (v reflect.Value, err error) {
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

func NewModelScope(model interface{}) *Scope {
	stl := []SymbolTable{}
	if model != nil {
		stl = append(stl, ModelSymbolTable{reflect.ValueOf(model)})
	}
	return &Scope{stl}
}

func NewModelSymbolTable(model interface{}) ModelSymbolTable {
	return ModelSymbolTable{reflect.ValueOf(model)}
}
