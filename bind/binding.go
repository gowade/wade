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
		Watch(fieldRefl reflect.Value, modelRefl reflect.Value, field string, callback func())
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

	if _, exist, _ := b.helpers.lookup(name); !exist {
		b.helpers.registerFunc(name, fn)
		return
	}

	panic(fmt.Sprintf("Helper with name %v already exists.", name))
	return
}

func (b *Binding) watchModel(binds []bindable, watches []token, root *expr, bs *bindScope, callback func(interface{})) error {
	for _, bi := range binds {
		if !bi.bindObj().fieldRefl.CanAddr() {
			return fmt.Errorf("Cannot watch this field. Please make sure it's addressable (struct fields are addressable; immutable values returned by a function are not).")
		}
		//use watchjs to watch for changes to the model
		(func(bi bindable) {
			bo := bi.bindObj()
			b.watcher.Watch(bo.fieldRefl, bo.modelRefl, bo.field, func() {
				newResult, _ := bs.evaluateRec(root, watches)
				callback(newResult.Interface())
			})
		})(bi)
	}

	return nil
}

func bstrPanic(mess, bindstring string, elem dom.Selection) {
	panic(dom.ElementError(elem, fmt.Sprintf(mess+`. While processing bind string "%v"`, bindstring)))
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

		roote, binds, watches, v, err := bs.evaluate(bexpr)
		if err != nil {
			bstrPanic(err.Error(), bstr, elem)
		}

		if len(binds) == 1 {
			fmodel := binds[0].bindObj().fieldRefl
			reportBinderError(binder.Watch(elem, func(newVal string) {
				if !fmodel.CanSet() {
					bstrPanic("Cannot set field.", bstr, elem)
				}
				fmodel.Set(reflect.ValueOf(newVal))
			}), bstr, elem)
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
				b.watchModel(binds, watches, roote, bs, func(newResult interface{}) {
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
	tokens, err := tokenize(bstr)
	if err != nil {
		bstrPanic(err.Error(), bstr, elem)
	}
	fbinds, err := parseFieldBind(tokens)
	if err != nil {
		bstrPanic(err.Error(), bstr, elem)
	}

	for field, btoks := range fbinds {
		watches, roote, er := parseBind(btoks)
		if er != nil {
			bstrPanic(er.Error(), bstr, elem)
		}

		binds, v, er := bs.evaluatePart(watches, roote)
		if er != nil {
			bstrPanic(er.Error(), bstr, elem)
		}

		oe, ok, err := evaluateObjField(field, reflect.ValueOf(tModel))
		if err != nil {
			bstrPanic(err.Error(), bstr, elem)
		}

		if !ok {
			bstrPanic(fmt.Sprintf(`No such field "%v" to bind to`, field), bstr, elem)
		}

		checkCompat := func(src reflect.Type, dst reflect.Type) {
			if !src.AssignableTo(dst) {
				bstrPanic(fmt.Sprintf(`Unassignable, incompatible types "%v" and "%v" of the model field and the value`,
					src.String(), dst.String()), bstr, elem)
			}
		}

		checkCompat(reflect.TypeOf(v), oe.fieldRefl.Type())
		oe.fieldRefl.Set(reflect.ValueOf(v))
		if !once {
			b.watchModel(binds, watches, roote, bs, func(newResult interface{}) {
				nr := reflect.ValueOf(newResult)
				checkCompat(nr.Type(), oe.fieldRefl.Type())
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
func (b *Binding) bindPrepare(relems dom.Selection, bs *bindScope, once bool, bindrelem bool, additionalbinds *AdditionalBinds, custagProcessing string) (bindTasks []func(), customElemTasks []func()) {
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

			tagname, _ := elem.TagName()
			custag, isCustom := b.tm.GetCustomTag(elem)
			if isCustom && tagname == custagProcessing {
				panic(dom.ElementError(elem,
					fmt.Sprintf(`Usage of custom tag "%v" inside its own definition. It would lead to an infinite loop!`, tagname, tagname)))
			}

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

			//Make a list of binds to perform (add to bindTasks)
			for _, attr := range attrs {
				name, bstr, ebs := attr.Name, attr.Value, attr.bs
				tagname, _ := elem.TagName()
				if name == "bind" { //field binding
					if !isCustom {
						bstrPanic(fmt.Sprintf(`Element %v hasn't been registered as a custom element`, tagname),
							bstr, elem)
					}

					(func(customTagModel interface{}, bs *bindScope) {
						bindTasks = append(bindTasks,
							wrapBindCall(elem, name, bstr, func(elem dom.Selection, astr, bstr string) {
								b.processFieldBind(bstr, elem, ebs, once, customTagModel)
							}))
					})(customTagModel, ebs)
				} else if strings.HasPrefix(name, BindPrefix) && //dom binding
					relem.Exists() == elem.Exists() {

					if isWrapper || isCustom {
						binds[name] = bstr
						continue
					}

					(func(bs *bindScope) {
						bindTasks = append(bindTasks,
							wrapBindCall(elem, name, bstr, func(elem dom.Selection, astr, bstr string) {
								b.processDomBind(astr, bstr, elem, bs, once)
							}))
					})(ebs)
				}
			}

			// Perform custom element rendering if it's a custom element, otherwise
			// get bind tasks and custom elem tasks from descendants.
			// Custom element's descendants are used as contents, so we don't recur to them
			if !bindrelem || idx > 0 {
				(func(bs *bindScope) {
					if isCustom {
						(func(elem dom.Selection, customTagModel interface{}, binds map[string]string) {
							customElemTasks = append(customElemTasks, func() {
								if bindingPrevented(elem, "-all-") {
									return
								}
								err := custag.PrepareTagContents(elem, customTagModel,
									func(contentElems dom.Selection) {
										b.bindWithScope(contentElems, once, true, bs.scope, nil)
									})

								if err != nil {
									dom.ElementError(elem, err.Error())
								}

								b.bindCustomElem(elem, once, false, b.newModelScope(customTagModel), &AdditionalBinds{binds, bs})

								elem.ReplaceWith(elem.Contents())
							})
						})(elem, customTagModel, binds)
					} else {
						bt, cet := b.bindPrepare(elem, bs, once, false, &AdditionalBinds{binds, bs}, custagProcessing)
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

func (b *Binding) Bind(relem dom.Selection, model interface{}, once bool, bindrelem bool) {
	b.BindModels(relem, []interface{}{model}, once, bindrelem)
}

func (b *Binding) bindWithScope(relem dom.Selection, once bool, bindrelem bool, s *scope, additionalbinds *AdditionalBinds) {
	// we have to do 2 steps like this to avoid missing out binding when things are removed
	btasks, customElemTasks := b.bindPrepare(relem, &bindScope{s}, once, bindrelem, additionalbinds, "")
	for _, fn := range btasks {
		fn()
	}

	for _, fn := range customElemTasks {
		fn()
	}
}

func (b *Binding) bindCustomElem(relem dom.Selection, once bool, bindrelem bool, s *scope, additionalbinds *AdditionalBinds) {
	tn, _ := relem.TagName()
	btasks, customElemTasks := b.bindPrepare(relem, &bindScope{s}, once, bindrelem, additionalbinds, tn)
	for _, fn := range btasks {
		fn()
	}

	for _, fn := range customElemTasks {
		fn()
	}
}
