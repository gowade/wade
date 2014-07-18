package bind

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	jq "github.com/gopherjs/jquery"
)

var (
	gJQ = jq.NewJQuery
)

const (
	BindPrefix         = "bind-"
	ReservedBindPrefix = "wade-rsvd"
)

func toString(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

type CustomElemManager interface {
	GetCustomTag(jq.JQuery) (CustomTag, bool)
}

type CustomTag interface {
	NewModel(jq.JQuery) interface{}
	TagContents(jq.JQuery)
}

type scopeSymbol interface {
	value() (reflect.Value, error)
	call([]reflect.Value) (reflect.Value, error)
}

type symbolTable interface {
	lookup(symbol string) (scopeSymbol, bool)
}

type scope struct {
	symTables []symbolTable
}

func (s *scope) lookup(symbol string) (sym scopeSymbol, err error) {
	for _, st := range s.symTables {
		var ok bool
		sym, ok = st.lookup(symbol)
		if ok {
			return
		}
	}

	err = fmt.Errorf("Unable to find symbol %v in the scope", symbol)
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

type funcSymbol struct {
	name string
	fn   reflect.Value
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

type modelSymbolTable struct {
	model reflect.Value
}

type modelFieldSymbol struct {
	name string
	eval *objEval
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
	return &scope{[]symbolTable{modelSymbolTable{reflect.ValueOf(model)}}}
}

type Binding struct {
	tm         CustomElemManager
	domBinders map[string]DomBinder
	helpers    mapSymbolTable

	scope     *scope
	pageModel interface{}
}

func NewBindEngine(tm CustomElemManager) *Binding {
	b := &Binding{
		tm:         tm,
		domBinders: defaultBinders(),
		helpers:    helpersSymbolTable(defaultHelpers()),
	}

	b.scope = &scope{[]symbolTable{b.helpers}}
	return b
}

// RegisterHelper registers fn as a helper with the given name.
//
// Helpers registered with this method are permanent, if you want to register
// a helper for just the current page, please use PageData.RegisterHelper.
func (b *Binding) RegisterHelper(name string, fn interface{}) {
	typ := reflect.TypeOf(fn)
	if typ.Kind() != reflect.Func {
		panic("Invalid helper, must be a function.")
	}

	if typ.NumOut() == 0 {
		panic("A helper must return something.")
	}

	if _, exist := b.helpers.lookup(name); !exist {
		b.helpers.registerFunc(name, fn)
		return
	}

	panic(fmt.Sprintf("Helper with name %v already exists.", name))
	return
}

func (b *Binding) newBindScope(model interface{}) *bindScope {
	s := b.scope
	if model != nil {
		s = &scope{[]symbolTable{modelSymbolTable{reflect.ValueOf(model)}}}
		s.merge(b.scope)
	}
	return &bindScope{s}
}

type objEval struct {
	fieldRefl reflect.Value
	modelRefl reflect.Value
	field     string
}

type bindable interface {
	bindObj() *objEval
}

type bindScope struct {
	scope *scope
}

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

func bindStringPanic(mess, bindstring string) {
	panic(fmt.Sprintf(mess+`, while processing bind string "%v".`, bindstring))
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

func (b *bindScope) evaluateBindString(bstr string) (root *expr, blist []bindable, value interface{}) {
	var err error
	root, blist, value, err = b.evaluate(bstr)
	if err != nil {
		bindStringPanic(err.Error(), bstr)
	}
	return
}

func (b *Binding) watchModel(binds []bindable, root *expr, bs *bindScope, callback func(interface{})) {
	for _, bi := range binds {
		//use watchjs to watch for changes to the model
		(func(bi bindable) {
			bo := bi.bindObj()
			obj := js.InternalObject(bo.modelRefl.Interface()).Get("$val")
			//workaround for gopherjs's protection disallowing js access to maps
			//setDummyHopFn(obj, "")
			js.Global.Call("watch",
				obj,
				bo.field,
				func(prop string, action string,
					_ js.Object,
					_2 js.Object) {
					newResult, _, _ := bs.evaluateRec(root)
					callback(newResult.Interface())
				})
		})(bi)
	}
}

func (b *Binding) processDomBind(astr, bstr string, elem jq.JQuery, bs *bindScope, once bool) {
	parts := strings.Split(astr, "-")
	if len(parts) <= 1 {
		panic(`Illegal "bind-".`)
	}

	if binder, ok := b.domBinders[parts[1]]; ok {
		binder = binder.BindInstance()
		args := make([]string, 0)
		if len(parts) >= 2 {
			for _, part := range parts[2:] {
				args = append(args, part)
			}
		}

		parts := strings.Split(bstr, "->")
		var bexpr string
		outputs := make([]string, 0)
		if len(parts) == 1 {
			bexpr = bstr
		} else {
			bexpr = strings.TrimSpace(parts[0])
			outputs = strings.Split(parts[1], ",")
			for i, ostr := range outputs {
				outputs[i] = strings.TrimSpace(ostr)
				for _, c := range outputs[i] {
					if !isValidExprChar(c) {
						bindStringPanic(fmt.Sprintf("invalid character %q", c), outputs[i])
					}
				}
			}
		}
		roote, binds, v := bs.evaluateBindString(bexpr)

		if len(binds) == 1 {
			fmodel := binds[0].bindObj().fieldRefl
			binder.Watch(elem, func(newVal string) {
				if !fmodel.CanSet() {
					panic("Cannot set field.")
				}
				fmodel.Set(reflect.ValueOf(newVal))
			})
		}

		metadata := fmt.Sprintf(`%v = "%v"`, astr, bstr)

		domBind := DomBind{
			Elem:     elem,
			Value:    v,
			Args:     args,
			outputs:  outputs,
			binding:  b,
			scope:    bs.scope,
			metadata: metadata,
		}
		(func(args, outputs []string) {
			binder.Bind(domBind)
			binder.Update(domBind)
			if !once {
				b.watchModel(binds, roote, bs, func(newResult interface{}) {
					domBind.Value = newResult
					binder.Update(domBind)
				})
			}
		})(args, outputs)
	} else {
		panic(fmt.Sprintf(`Dom binder "%v" does not exist.`, parts[1]))
	}
}

func (b *Binding) processAttrBind(astr, bstr string, elem jq.JQuery, bs *bindScope, once bool, tModel interface{}) {
	fbinds := strings.Split(bstr, ";")
	for i, fb := range fbinds {
		if i == len(fbinds)-1 && fb == "" {
			continue
		}
		fv := strings.Split(fb, ":")
		if len(fv) != 2 {
			bindStringPanic(`There should be one ":" in each attribute bind`, bstr)
		}
		field := strings.TrimSpace(fv[0])
		valuestr := strings.TrimSpace(fv[1])
		for _, c := range field {
			if !isValidExprChar(c) {
				bindStringPanic(fmt.Sprintf("invalid character %q", c), field)
			}
		}

		roote, binds, v := bs.evaluateBindString(valuestr)

		oe, ok := evaluateObjField(field, reflect.ValueOf(tModel))
		if !ok {
			bindStringPanic(fmt.Sprintf(`No such field "%v" to bind to`, field), bstr)
		}
		isCompat := func(src reflect.Type, dst reflect.Type) {
			if !src.AssignableTo(dst) {
				bindStringPanic(fmt.Sprintf(`Unassignable, incompatible types "%v" and "%v" of the model field and the value`,
					src.String(), dst.String()), bstr)
			}
		}
		isCompat(reflect.TypeOf(v), oe.fieldRefl.Type())
		oe.fieldRefl.Set(reflect.ValueOf(v))
		if !once {
			b.watchModel(binds, roote, bs, func(newResult interface{}) {
				nr := reflect.ValueOf(newResult)
				isCompat(nr.Type(), oe.fieldRefl.Type())
				oe.fieldRefl.Set(nr)
			})
		}
	}
}

func preventBinding(elem jq.JQuery, bindattr string) {
	elem.SetAttr(strings.Join([]string{ReservedBindPrefix, bindattr}, "-"), "t")
}

func preventTreeBinding(elem jq.JQuery, bindattr string) {
	elem.Find("*").Each(func(_ int, d jq.JQuery) {
		preventBinding(d, bindattr)
	})
}

func preventAllBinding(elem jq.JQuery) {
	elem.Find("*").Each(func(_ int, d jq.JQuery) {
		preventBinding(d, "all")
	})
}

func bindingPrevented(elem jq.JQuery, bindattr string) bool {
	return elem.Attr(ReservedBindPrefix+"-all") == "t" ||
		elem.Attr(strings.Join([]string{ReservedBindPrefix, bindattr}, "-")) == "t"
}

func wrapBindCall(elem jq.JQuery, bindattr, bindstr string, fn func(string, string)) func() {
	return func() {
		if !bindingPrevented(elem, bindattr) {
			fn(bindattr, bindstr)
			preventBinding(elem, bindattr)
		}
	}
}

// bind parses the bind string, make a list of binds (this doesn't actually bind the elements)
func (b *Binding) bindPrepare(relem jq.JQuery, bs *bindScope, once bool) (bindTasks []func()) {
	if relem.Length == 0 {
		panic("Incorrect element for bind.")
	}

	bindTasks = make([]func(), 0)

	relem.Children("*").Each(func(i int, elem jq.JQuery) {
		custag, isCustom := b.tm.GetCustomTag(elem)

		htmla := elem.Get(0).Get("attributes")
		attrs := make(map[string]string)
		for i := 0; i < htmla.Length(); i++ {
			attr := htmla.Index(i)
			attrs[attr.Get("name").Str()] = attr.Get("value").Str()
		}

		var customTagModel interface{} = nil
		if isCustom {
			customTagModel = custag.NewModel(elem)
		}

		for name, bstr := range attrs {
			if name == "bind" { //attribute binding
				if !isCustom {
					panic(fmt.Sprintf("Attribute binding syntax can only be used for custom elements."))
				}
				bindTasks = append(bindTasks,
					wrapBindCall(elem, name, bstr, func(astr, bstr string) {
						b.processAttrBind(astr, bstr, elem, bs, once, customTagModel)
					}))
			} else if strings.HasPrefix(name, BindPrefix) && //dom binding
				jqExists(elem) { //element still exists
				if isCustom {
					panic(`Dom binding is not allowed for custom element tags (they should not actually be rendered
			, so there's no point; but of course inside the custom element's contents it's allowed normally).
			If you want to bind the attributes of a custom element, use attribute binding instead.`)
				}
				bindTasks = append(bindTasks,
					wrapBindCall(elem, name, bstr, func(astr, bstr string) {
						b.processDomBind(astr, bstr, elem, bs, once)
					}))
			}
		}

		if isCustom {
			custag.TagContents(elem)
			bindTasks = append(bindTasks, b.bindPrepare(elem, b.newBindScope(customTagModel), once)...)
		} else {
			bindTasks = append(bindTasks, b.bindPrepare(elem, bs, once)...)
		}
	})

	return
}

// Bind binds a model to an element and all its children
func (b *Binding) Bind(relem jq.JQuery, model interface{}, once bool) {
	// we have to do 2 steps like this to avoid missing out binding when things are removed
	btasks := b.bindPrepare(relem, b.newBindScope(model), once)
	for _, fn := range btasks {
		fn()
	}
}

func (b *Binding) bindWithScope(relem jq.JQuery, model interface{}, once bool, s *scope) {
	// we have to do 2 steps like this to avoid missing out binding when things are removed
	btasks := b.bindPrepare(relem, &bindScope{s}, once)
	for _, fn := range btasks {
		fn()
	}
}
