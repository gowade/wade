package bind

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/icommon"
)

const (
	BindPrefix         = "bind-"
	ReservedBindPrefix = "wade-rsvd"
)

type (
	CustomElemManager interface {
		GetCustomTag(dom.Selection) (CustomTag, bool)
	}

	CustomTag interface {
		NewModel(dom.Selection) interface{}
		PrepareTagContents(dom.Selection, interface{}, func(dom.Selection)) error
	}

	DomAttr struct {
		dom.Attr
		bs *bindScope
	}

	AdditionalBinds struct {
		binds map[string]string
		bs    *bindScope
	}

	jsWatcher interface {
		Watch(modelRefl reflect.Value, field string, callback func())
	}

	Binding struct {
		tm         CustomElemManager
		domBinders map[string]DomBinder
		helpers    mapSymbolTable

		watcher   jsWatcher
		scope     *scope
		pageModel interface{}
	}

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

func NewBindEngine(tm CustomElemManager, watcher jsWatcher) *Binding {
	b := &Binding{
		tm:         tm,
		watcher:    watcher,
		domBinders: defaultBinders(),
		helpers:    helpersSymbolTable(defaultHelpers()),
	}

	b.scope = &scope{[]symbolTable{b.helpers}}
	return b
}

// RegisterHelper registers a function as a global helper with the given name.
//
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

func (b *bindScope) clone() *bindScope {
	scope := newScope()
	scope.merge(b.scope)
	return &bindScope{scope}
}

func (b *Binding) watchModel(binds []bindable, root *expr, bs *bindScope, callback func(interface{})) {
	for _, bi := range binds {
		//use watchjs to watch for changes to the model
		(func(bi bindable) {
			bo := bi.bindObj()
			b.watcher.Watch(bo.modelRefl, bo.field, func() {
				newResult, _, _ := bs.evaluateRec(root)
				callback(newResult.Interface())
			})
		})(bi)
	}
}

func (b *Binding) processDomBind(astr, bstr string, elem dom.Selection, bs *bindScope, once bool) {
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
					icommon.WrapperUnwrap(elem)
				})
			}
		})(args, outputs)
	} else {
		panic(fmt.Sprintf(`Dom binder "%v" does not exist.`, parts[1]))
	}
}

func (b *Binding) processAttrBind(astr, bstr string, elem dom.Selection, bs *bindScope, once bool, tModel interface{}) {
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

func preventBinding(elem dom.Selection, bindattr string) {
	elem.SetAttr(strings.Join([]string{ReservedBindPrefix, bindattr}, "-"), "t")
}

func preventTreeBinding(elem dom.Selection, bindattr string) {
	preventBinding(elem, bindattr)
	for _, d := range elem.Find("*").Elements() {
		preventBinding(d, bindattr)
	}
}

func preventAllBinding(elem dom.Selection) {
	preventBinding(elem, "all")
	for _, d := range elem.Find("*").Elements() {
		preventBinding(d, "all")
	}
}

func bindingPrevented(elem dom.Selection, bindattr string) bool {
	allb, ok1 := elem.Attr(ReservedBindPrefix + "-all")
	bb, ok2 := elem.Attr(strings.Join([]string{ReservedBindPrefix, bindattr}, "-"))
	return (ok1 && allb == "t") || (ok2 && bb == "t")
}

func wrapBindCall(elem dom.Selection, bindattr, bindstr string, fn func(dom.Selection, string, string)) func() {
	return func() {
		if !bindingPrevented(elem, bindattr) {
			fn(elem, bindattr, bindstr)
			preventBinding(elem, bindattr)
		}
	}
}

// bind parses the bind string, make a list of binds (this doesn't actually bind the elements)
func (b *Binding) bindPrepare(relems dom.Selection, bs *bindScope, once bool, bindrelem bool, additionalbinds *AdditionalBinds) (bindTasks []func(), customElemTasks []func()) {
	if relems.Length() == 0 {
		panic("Incorrect element for bind.")
	}

	bindTasks = make([]func(), 0)
	customElemTasks = make([]func(), 0)

	for _, relem := range relems.Elements() {
		elems := make([]dom.Selection, 0)
		if bindrelem {
			elems = append(elems, relem)
		}

		elems = append(elems, relems.Contents().Elements()...)

		for idx, elem := range elems {
			if !elem.IsElement() {
				continue
			}

			custag, isCustom := b.tm.GetCustomTag(elem)
			isWrapper := icommon.IsWrapperElem(elem)
			var binds map[string]string
			if isWrapper || isCustom {
				binds = make(map[string]string)
			}

			bsclone := bs.clone()

			attrs := make([]DomAttr, 0)
			if additionalbinds != nil {
				for k, v := range additionalbinds.binds {
					attrs = append(attrs, DomAttr{dom.Attr{k, v}, additionalbinds.bs})
				}
			}

			for _, dattr := range elem.Attrs() {
				attrs = append(attrs, DomAttr{dattr, bsclone})
			}

			var customTagModel interface{} = nil
			if isCustom {
				customTagModel = custag.NewModel(elem)
			}

			for _, attr := range attrs {
				name, bstr, ebs := attr.Name, attr.Value, attr.bs
				tagname, _ := elem.TagName()
				if name == "bind" { //attribute binding
					if !isCustom {
						panic(fmt.Sprintf(`Processing bind string %v="%v": Element %v hasn't been registered as a custom element.`, name, bstr, tagname))
					}

					(func(customTagModel interface{}, bs *bindScope) {
						bindTasks = append(bindTasks,
							wrapBindCall(elem, name, bstr, func(elem dom.Selection, astr, bstr string) {
								b.processAttrBind(astr, bstr, elem, ebs, once, customTagModel)
							}))
					})(customTagModel, ebs)
				} else if strings.HasPrefix(name, BindPrefix) && //dom binding
					relem.Exists() == elem.Exists() { //element still exists

					if isWrapper || isCustom {
						binds[name] = bstr
						continue
					}

					(func(bs *bindScope) {
						bindTasks = append(bindTasks,
							wrapBindCall(elem, name, bstr, func(elem dom.Selection, astr, bstr string) {
								b.processDomBind(astr, bstr, elem, ebs, once)
							}))
					})(ebs)
				}
			}

			if !bindrelem || idx > 0 {
				(func(bs *bindScope) {
					if isCustom {
						(func(elem dom.Selection, customTagModel interface{}, binds map[string]string) {
							customElemTasks = append(customElemTasks, func() {
								err := custag.PrepareTagContents(elem, customTagModel,
									func(contentElem dom.Selection) {
										s := newModelScope(customTagModel)
										s.merge(bs.scope)
										b.bindWithScope(contentElem, once, true, s, nil)
									})

								if err != nil {
									dom.ElementError(elem, err.Error())
								}

								b.bindWithScope(elem, once, false, b.newModelScope(customTagModel), &AdditionalBinds{binds, bs})

								elem.ReplaceWith(elem.Contents())
							})
						})(elem, customTagModel, binds)
					} else {
						bt, cet := b.bindPrepare(elem, bs, once, false, &AdditionalBinds{binds, bs})
						bindTasks = append(bindTasks, bt...)
						customElemTasks = append(customElemTasks, cet...)
					}
				})(bs)
			}
		}
	}

	return
}

func (b *Binding) newModelScope(model interface{}) *scope {
	s := newModelScope(model)
	s.merge(b.scope)
	return s
}

// Bind binds a model to an element and its ascendants
func (b *Binding) Bind(relem dom.Selection, model interface{}, once bool, bindrelem bool) {
	b.bindWithScope(relem, once, bindrelem, b.newModelScope(model), nil)
}

// BindMergeScope merges the given scope to the basic scope and performs binding
func (b *Binding) BindModels(relem dom.Selection, models []interface{}, once bool, bindrelem bool) {
	s := newScope()
	for _, model := range models {
		if model != nil {
			s.symTables = append(s.symTables, modelSymbolTable{reflect.ValueOf(model)})
		}
	}
	s.merge(b.scope)

	b.bindWithScope(relem, once, bindrelem, s, nil)
}

func (b *Binding) bindWithScope(relem dom.Selection, once bool, bindrelem bool, s *scope, additionalbinds *AdditionalBinds) {
	// we have to do 2 steps like this to avoid missing out binding when things are removed
	btasks, customElemTasks := b.bindPrepare(relem, &bindScope{s}, once, bindrelem, additionalbinds)
	for _, fn := range btasks {
		fn()
	}

	for _, fn := range customElemTasks {
		fn()
	}
}
