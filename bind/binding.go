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

func (b *Binding) RegisterDomBinder(name string, binder DomBinder) {
	if _, exists := b.domBinders[name]; exists {
		panic(fmt.Sprintf(`A binder with that name "%v" already exists.`, name))
	}

	b.domBinders[name] = binder
}

// RegisterHelper registers a function as a global helper with the given name.
//
func (b *Binding) RegisterHelper(name string, fn interface{}) {
	typ := reflect.TypeOf(fn)
	if typ.Kind() != reflect.Func {
		panic(fmt.Sprintf("Invalid helper %v, must be a function.", name))
	}

	if typ.NumOut() == 0 {
		panic(fmt.Sprintf("Invalid helper %v, a helper must return something.", name))
	}

	if _, exist := b.helpers.lookup(name); !exist {
		b.helpers.registerFunc(name, fn)
		return
	}

	panic(fmt.Sprintf("Helper with name %v already exists.", name))
	return
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

func bstrPanic(mess, bindstring string, elem dom.Selection) {
	panic(dom.ElementError(elem, fmt.Sprintf(mess+`, while processing bind string "%v"`, bindstring)))
}

func reportBinderError(err error, bstr string, elem dom.Selection) {
	if err != nil {
		bstrPanic(err.Error(), bstr, elem)
	}
}

func parseDomBindstr(bstr string) (bexpr string, outputs []string, err error) {
	parts := strings.Split(bstr, "->")
	outputs = make([]string, 0)
	if len(parts) == 1 {
		bexpr = bstr
	} else {
		bexpr = strings.TrimSpace(parts[0])
		outputs = strings.Split(parts[1], ",")
		for i, ostr := range outputs {
			outputs[i] = strings.TrimSpace(ostr)
			for _, c := range outputs[i] {
				if !isValidExprChar(c) {
					err = fmt.Errorf("invalid character %q", c)
					return
				}
			}
		}
	}

	return
}

func (b *Binding) processDomBind(astr, bstr string, elem dom.Selection, bs *bindScope, once bool) {
	parts := strings.Split(astr, "-")
	if len(parts) <= 1 {
		bstrPanic(`Something's wrong, illegal "bind-".`, bstr, elem)
	}

	if binder, ok := b.domBinders[parts[1]]; ok {
		binder = binder.BindInstance()
		args := make([]string, 0)
		if len(parts) >= 2 {
			for _, part := range parts[2:] {
				args = append(args, part)
			}
		}

		bexpr, outputs, err := parseDomBindstr(bstr)
		if err != nil {
			bstrPanic(err.Error(), bstr, elem)
		}

		roote, binds, v, err := bs.evaluate(bexpr)
		if err != nil {
			bstrPanic(err.Error(), bstr, elem)
		}

		if len(binds) == 1 {
			fmodel := binds[0].bindObj().fieldRefl
			binder.Watch(elem, func(newVal string) {
				if !fmodel.CanSet() {
					bstrPanic("Cannot set field.", bstr, elem)
				}
				fmodel.Set(reflect.ValueOf(newVal))
			})
		}

		domBind := DomBind{
			Elem:    elem,
			Value:   v,
			Args:    args,
			outputs: outputs,
			binding: b,
			scope:   bs.scope,
		}

		(func(args, outputs []string, bstr string, elem dom.Selection) {
			reportBinderError(binder.Bind(domBind), bstr, elem)
			reportBinderError(binder.Update(domBind), bstr, elem)
			if !once {
				b.watchModel(binds, roote, bs, func(newResult interface{}) {
					domBind.Value = newResult
					reportBinderError(binder.Update(domBind), bstr, elem)
					icommon.WrapperUnwrap(elem)
				})
			}
		})(args, outputs, bstr, elem)
	} else {
		bstrPanic(fmt.Sprintf(`Dom binder "%v" does not exist.`, parts[1]), bstr, elem)
	}
}

func (b *Binding) processFieldBind(bstr string, elem dom.Selection, bs *bindScope, once bool, tModel interface{}) {
	fbinds := strings.Split(bstr, ";")
	for i, fb := range fbinds {
		if i == len(fbinds)-1 && fb == "" {
			continue
		}
		fv := strings.Split(fb, ":")
		if len(fv) != 2 {
			bstrPanic(fmt.Sprintf(`Invalid syntax. There should be a ":" in each binding of a field instead of %v`, len(fv)), bstr, elem)
		}
		field := strings.TrimSpace(fv[0])
		valuestr := strings.TrimSpace(fv[1])
		for _, c := range field {
			if !isValidExprChar(c) {
				bstrPanic(fmt.Sprintf("invalid character %q", c), field, elem)
			}
		}

		roote, binds, v, err := bs.evaluate(valuestr)
		if err != nil {
			bstrPanic(err.Error(), bstr, elem)
		}

		oe, ok := evaluateObjField(field, reflect.ValueOf(tModel))
		if !ok {
			bstrPanic(fmt.Sprintf(`No such field "%v" to bind to`, field), bstr, elem)
		}
		isCompat := func(src reflect.Type, dst reflect.Type) {
			if !src.AssignableTo(dst) {
				bstrPanic(fmt.Sprintf(`Unassignable, incompatible types "%v" and "%v" of the model field and the value`,
					src.String(), dst.String()), bstr, elem)
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
						bstrPanic(fmt.Sprintf(`Element %v hasn't been registered as a custom element.`, tagname),
							bstr, elem)
					}

					(func(customTagModel interface{}, bs *bindScope) {
						bindTasks = append(bindTasks,
							wrapBindCall(elem, name, bstr, func(elem dom.Selection, astr, bstr string) {
								b.processFieldBind(bstr, elem, ebs, once, customTagModel)
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
