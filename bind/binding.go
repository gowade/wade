package bind

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/phaikawl/wade/custom"
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/icommon"
)

const (
	BindPrefix         = "bind-"
	ReservedBindPrefix = "wade-rsvd"
)

type (
	CustomElem interface {
		Update() error
		Model() interface{}
		PrepareContents(func(dom.Selection)) error
		Element() dom.Selection
	}

	CustomTag interface {
		NewElem(dom.Selection) *custom.TagManager
	}

	Binding struct {
		tm         *custom.TagManager
		domBinders map[string]DomBinder
		helpers    helpersSymbolTable

		watcher   *Watcher
		scope     *scope
		pageModel interface{}
	}

	drmElem struct {
		dom.Selection
		rmList *[]dom.Selection
	}
)

func NewTestBindEngine() *Binding {
	return NewBindEngine(custom.NewTagManager(), NoopJsWatcher{})
}

func NewBindEngine(tm *custom.TagManager, jsWatcher JsWatcher) *Binding {
	b := &Binding{
		tm:         tm,
		watcher:    NewWatcher(jsWatcher),
		domBinders: defaultBinders(),
		helpers:    newHelpersSymbolTable(defaultHelpers()),
	}

	b.scope = &scope{[]symbolTable{b.helpers}}
	return b
}

func (b *Binding) Watcher() *Watcher {
	return b.watcher
}

func (b *Binding) RegisterDomBinder(name string, binder DomBinder) {
	if _, exists := b.domBinders[name]; exists {
		panic(fmt.Sprintf(`A binder with that name "%v" already exists.`, name))
	}

	b.domBinders[name] = binder
}

func (b *Binding) TagManager() *custom.TagManager {
	return b.tm
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
			return fmt.Errorf(`Cannot watch field "%v". Please make sure it's addressable. If you don't intend to watch for its changes, please use a pipe symbol ("|")`, bi.bindObj().field)
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

func reportError(err error, bstr string, elem dom.Selection) {
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

func (e drmElem) Remove() {
	*e.rmList = append(*e.rmList, e.Selection)
}

func (e drmElem) ReplaceWith(sel dom.Selection) {
	e.Before(sel)
	e.Remove()
}

func (b *Binding) processFieldBind(bstr string, elem dom.Selection, bs *bindScope, once bool, ce CustomElem) {
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

		oe, ok, err := evaluateObjField(field, reflect.ValueOf(ce.Model()))
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
			err = b.watchModel(binds, watches, roote, bs, func(newResult interface{}) {
				nr := reflect.ValueOf(newResult)
				checkCompat(nr.Type(), oe.fieldRefl.Type())
				oe.fieldRefl.Set(nr)

				err := ce.Update()
				if err != nil {
					panic(dom.ElementError(ce.Element(), err.Error()))
				}
			})

			if err != nil {
				bstrPanic(err.Error(), bstr, elem)
			}
		}
	}
}

func (b *Binding) bindCustomElemsRec(elem dom.Selection, bs *bindScope, once bool, scopeElem dom.Selection) {
	if !elem.IsElement() || !elem.Exists() {
		return
	}

	if bound, _ := elem.Attr(ReservedBindPrefix + "-bound"); bound == "true" {
		return
	}

	tag, isCustom := b.tm.GetTag(elem)
	if isCustom {
		if scopeElem != nil {
			if setag, _ := scopeElem.TagName(); tag.Name == setag {
				panic(dom.ElementError(elem,
					fmt.Sprintf(`Infinite loop detected. Usage of custom tag "%v" inside its own definition.`, tag.Name),
				))
			}
		}

		customElem := tag.NewElem(elem)

		if fieldBind, ok := elem.Attr("bind"); ok {
			b.processFieldBind(fieldBind, elem, bs, once, customElem)
		}

		err := customElem.PrepareContents(func(contentElems dom.Selection) {
			b.bindWithScope(contentElems, bs.scope, once, true, scopeElem)
		})

		if err != nil {
			dom.ElementError(elem, err.Error())
		}

		b.bindWithScope(elem, b.newModelScope(customElem.Model()), once, false, elem)
	}

	for _, e := range elem.Children().Elements() {
		b.bindCustomElemsRec(e, bs, once, scopeElem)
	}
}

func (b *Binding) processDomBind(astr, bstr string, elem dom.Selection, bs *bindScope, once bool) (removedElems []dom.Selection, err error) {
	parts := strings.Split(astr, "-")
	if len(parts) <= 1 {
		err = fmt.Errorf(`Something's wrong, illegal "bind-".`)
		return
	}

	removedElems = make([]dom.Selection, 0)

	if binder, ok := b.domBinders[parts[1]]; ok {
		binder = binder.BindInstance()
		args := make([]string, 0)
		if len(parts) >= 2 {
			for _, part := range parts[2:] {
				args = append(args, part)
			}
		}

		bexpr, outputs, err2 := parseDomBindstr(bstr)
		if err2 != nil {
			err = err2
			return
		}

		roote, binds, watches, v, err2 := bs.evaluate(bexpr)
		if err2 != nil {
			err = err2
			return
		}

		if len(binds) == 1 {
			fmodel := binds[0].bindObj().fieldRefl
			err = binder.Watch(elem, func(newVal string) {
				if !fmodel.CanSet() {
					bstrPanic("2-way data binding on unchangable field", bstr, elem)
				}
				fmodel.Set(reflect.ValueOf(newVal))
			})

			if err != nil {
				return
			}
		}

		domBind := DomBind{
			Elem:    drmElem{elem, &removedElems},
			Value:   v,
			Args:    args,
			outputs: outputs,
			binding: b,
			scope:   bs.scope,
		}

		(func(args, outputs []string, bstr string, elem dom.Selection) {
			err = binder.Bind(domBind)
			if err != nil {
				return
			}

			err = binder.Update(domBind)
			if err != nil {
				return
			}

			if !once {
				udb := domBind
				udb.Elem = elem

				err = b.watchModel(binds, watches, roote, bs, func(newResult interface{}) {
					udb.Value = newResult
					reportError(binder.Update(udb), bstr, elem)
					icommon.WrapperUnwrap(elem)
				})
			}
		})(args, outputs, bstr, elem)

	} else {
		err = fmt.Errorf(`Dom binder "%v" does not exist.`, parts[1])
	}

	return
}

func (b *Binding) bindDomRec(elem dom.Selection,
	bs *bindScope,
	once bool,
	additionalBinds []dom.Attr) (replaced dom.Selection) {

	if !elem.IsElement() || !elem.Exists() {
		return
	}

	//println(dom.DebugInfo(elem))
	//println(len(bs.scope.symTables))

	//println(dom.DebugInfo(elem))

	replaced = elem

	isWrapper := icommon.IsWrapperElem(elem)
	var abinds []dom.Attr
	if isWrapper {
		abinds = make([]dom.Attr, 0)
	}

	attrs := make([]dom.Attr, 0)
	if additionalBinds != nil {
		attrs = append(attrs, additionalBinds...)
	}

	attrs = append(attrs, elem.Attrs()...)

	removedElems := make([][]dom.Selection, 0)
	// perform binding
	for _, attr := range attrs {
		astr, bstr := attr.Name, attr.Value
		if strings.HasPrefix(astr, BindPrefix) && elem.Exists() {
			if isWrapper {
				abinds = append(abinds, dom.Attr{astr, bstr})
				continue
			}

			//println(astr, dom.DebugInfo(elem))

			rmdElems, err := b.processDomBind(astr, bstr, elem, bs, once)
			if err != nil {
				bstrPanic(err.Error(), bstr, elem)
			}

			removedElems = append(removedElems, rmdElems)

			//prevent duplicate binding
			elem.RemoveAttr(astr)
			elem.SetAttr("done-"+astr, bstr)
		}
	}

	if isWrapper {
		conts := elem.Contents()
		elem.ReplaceWith(conts)
		for _, child := range conts.Elements() {
			b.bindDomRec(child, bs, once, abinds)
		}

		replaced = conts
		return
	} else {
		if _, isCustom := b.tm.GetTag(elem); !isCustom {
			for _, child := range elem.Children().Elements() {
				b.bindDomRec(child, bs, once, nil)
			}
		}
	}

	for _, l := range removedElems {
		for _, e := range l {
			e.Remove()
		}
	}

	return
}

func (b *Binding) newModelScope(model interface{}) *scope {
	s := newModelScope(model)
	s.merge(b.scope)
	return s
}

func (b *Binding) BindModels(rootElem dom.Selection, models []interface{}, once bool) {
	if !rootElem.Children().First().Exists() {
		panic("Invalid root element for bind. It must be a node in a real html document, a <wroot> or a child of <wroot>.")
	}

	s := newScope()
	for _, model := range models {
		if model != nil {
			s.symTables = append(s.symTables, modelSymbolTable{reflect.ValueOf(model)})
		}
	}
	s.merge(b.scope)

	b.bindWithScope(rootElem, s, once, false, rootElem)
}

func (b *Binding) Bind(rootElem dom.Selection, model interface{}, once bool) {
	b.BindModels(rootElem, []interface{}{model}, once)
}

func (b *Binding) rootList(rootElems dom.Selection, bindRoot bool) []dom.Selection {
	if bindRoot {
		return rootElems.Elements()
	}

	return rootElems.Children().Elements()
}

func (b *Binding) bindWithScope(rootElems dom.Selection, s *scope, once bool, bindRoot bool, scopeElem dom.Selection) {
	bs := &bindScope{s}
	elems := b.rootList(rootElems, bindRoot)

	for _, e := range elems {
		if !e.IsElement() || !e.Exists() {
			continue
		}

		replacedElems := b.bindDomRec(e, bs, once, nil)
		for _, re := range replacedElems.Elements() {
			b.bindCustomElemsRec(re, bs, once, scopeElem)
		}
	}

	for _, re := range rootElems.Elements() {
		re.SetAttr(ReservedBindPrefix+"-bound", "true")
	}
}
