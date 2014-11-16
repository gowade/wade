package core

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	. "github.com/phaikawl/wade/scope"
)

const (
	AttrBindPrefix   = '@'
	BinderBindPrefix = '#'
)

type (
	Application interface {
		ErrChanPut(error)
	}

	Binding struct {
		app     Application
		tm      *ComManager
		binders map[string]Binder
		helpers HelpersSymbolTable

		scope     *Scope
		pageModel interface{}
	}

	BindingError struct {
		Err error
	}
)

func (be BindingError) Error() string {
	return be.Err.Error()
}

type DummyApp struct {
}

func (da DummyApp) ErrChanPut(err error) {
}

func NewTestBindEngine() *Binding {
	return NewBindEngine(DummyApp{}, NewComManager(nil))
}

func NewBindEngine(app Application, tm *ComManager) *Binding {
	b := &Binding{
		app:     app,
		tm:      tm,
		binders: map[string]Binder{},
		helpers: NewHelpersSymbolTable(defaultHelpers()),
	}

	b.scope = NewScope([]SymbolTable{b.helpers})
	return b
}

//RegisterBinder registers a binder
func (b *Binding) RegisterBinder(name string, binder Binder) {
	if _, exists := b.binders[name]; exists {
		panic(fmt.Sprintf(`A binder with that name "%v" already exists.`, name))
	}

	b.binders[name] = binder
}

func (b *Binding) ComponentManager() *ComManager {
	return b.tm
}

// RegisterHelper registers a function as a helper with the given name.
// Helpers are global.
func (b *Binding) RegisterHelper(name string, fn interface{}) {
	typ := reflect.TypeOf(fn)
	if typ.Kind() != reflect.Func {
		panic(fmt.Sprintf("Invalid helper %v, must be a function.", name))
	}

	if typ.NumOut() == 0 {
		panic(fmt.Sprintf("Invalid helper %v, a helper must return something.", name))
	}

	if _, exist, _ := b.helpers.Lookup(name); !exist {
		b.helpers.RegisterFunc(name, fn)
		return
	}

	panic(fmt.Sprintf("Helper with name %v already exists.", name))
	return
}

func bstrPanic(mess, bindstring string, node *VNode) {
	panic(fmt.Sprintf(mess+` While processing bind string "%v".`, bindstring))
}

func reportError(err error, bstr string, elem *VNode) {
	if err != nil {
		bstrPanic(err.Error(), bstr, elem)
	}
}

func (b *Binding) processAttrBind(attr, bstr string, node *VNode, scope *Scope) (err error) {
	bs := bindScope{scope}
	_, node.Attrs[attr], err = bs.evaluate(bstr)

	node.addCallback(func() (err error) {
		_, node.Attrs[attr], err = bs.evaluate(bstr)
		return
	})

	return
}

func (b *Binding) processMustache(node *VNode, scope *Scope) (err error) {
	bs := bindScope{scope}
	_, v, err := bs.evaluate(node.Binds[0].Expr)
	node.Data = toString(v)

	node.addCallback(func() (err error) {
		_, v, err := bs.evaluate(node.Binds[0].Expr)
		node.Data = toString(v)
		return
	})

	return
}

func (b *Binding) processFieldBind(field string, bstr string, node *VNode, scope *Scope, ci *componentInstance) (err error) {
	_, v, err := bindScope{scope}.evaluate(bstr)
	if err != nil {
		return
	}

	oe, ok, err := EvaluateObjField(field, reflect.ValueOf(ci.model))
	if err != nil {
		return
	}

	if !ok {
		return fmt.Errorf(`No such field "%v" to bind to.`, field)
	}

	src, dst := reflect.TypeOf(v), oe.FieldRefl.Type()
	if !src.AssignableTo(dst) {
		return fmt.Errorf(`Unassignable, incompatible types "%v" and "%v" of the model field and the value.`,
			dst.String(), src.String())
	}

	oe.FieldRefl.Set(reflect.ValueOf(v))
	node.addCallback(func() (err error) {
		_, v, err := bindScope{scope}.evaluate(bstr)
		if err != nil {
			return
		}

		oe.FieldRefl.Set(reflect.ValueOf(v))

		return
	})

	return
}

func (b *Binding) bindComponent(node *VNode, scope *Scope, cv *componentView) (err error) {
	ci, err := cv.NewInstance(node)
	if err != nil {
		return fmt.Errorf(`Failed initialization of the component instance, error in its Init(). Error: %v.`, err.Error())
	}

	for _, bind := range node.Binds {
		if bind.Type == AttrBind {
			//bindinfo += fmt.Sprintf("{%v: [%v]} ", hattr.Name, hattr.Value)

			field := strings.Split(bind.Name, ".")[0]
			if ok, fieldName := cv.HasAttr(field); ok {
				b.processFieldBind(fieldName, bind.Expr, node, scope, ci)
			} else {
				b.processAttrBind(bind.Name, bind.Expr, node, scope)
			}
		}
	}

	ci.prepareInner(scope)

	if err != nil {
		return
	}

	for i, _ := range node.Children {
		b.bindWithScope(&node.Children[i], b.newModelScope(ci.model))
	}
	return
}

var (
	NameRegexp = regexp.MustCompile(`\w+`)
)

func checkName(strs []string) error {
	for _, str := range strs {
		if !NameRegexp.MatchString(str) {
			return fmt.Errorf("Invalid name %v", str)
		}
	}

	return nil
}

func parseBinderLHS(astr string) (binder string, args []string, err error) {
	lp := strings.IndexRune(astr, '(')
	if lp != -1 {
		if astr[len(astr)-1] != ')' {
			err = fmt.Errorf("Invalid syntax for left hand side of binding")
			return
		}

		binder = astr[:lp]
		args = strings.Split(astr[lp+1:len(astr)-1], ",")
	} else {
		binder = astr
		args = []string{}
	}

	err = checkName(append(args, binder))

	return
}

func (b *Binding) processBinderBind(astr, bstr string, node *VNode, scope *Scope) (err error) {
	binderName, args, err := parseBinderLHS(astr)
	if err != nil {
		return
	}

	binder, ok := b.binders[binderName]
	if !ok {
		err = fmt.Errorf(`Binder "%v" does not exist`, binderName)
		return
	}

	binder = binder.BindInstance()

	bs := bindScope{scope}
	roote, v, err2 := bs.evaluate(bstr)
	if err2 != nil {
		err = err2
		return
	}

	domBind := DomBind{
		Node:    node,
		Value:   v,
		Args:    args,
		binding: b,
		scope:   bs.scope,
	}

	if tw, ok := binder.(TwoWayBinder); ok {
		if roote.typ != ValueExpr {
			err = fmt.Errorf("No function call is allowed in an expression for 2-way binding.")
			return
		}

		var sym ScopeSymbol
		sym, err = scope.Lookup(roote.name)
		if err != nil {
			return
		}

		err = tw.Listen(domBind, func(newVal string) {
			bdb, ok := sym.(Bindable)
			if !ok {
				bstrPanic("2-way data binding on unchangable field", bstr, node)
			}

			bdb.BindObj().FieldRefl.Set(reflect.ValueOf(newVal))
		})

		if err != nil {
			return
		}
	}

	//gopherjs:blocking
	err = binder.Bind(domBind)
	if err != nil {
		return
	}

	node.addCallback(func() (err error) {
		//gopherjs:blocking
		_, domBind.Value, _ = bs.evaluate(bstr)
		return binder.Update(domBind)
	})

	return
}

//var (
//	MustacheRegex = regexp.MustCompile("{{([^{}]+)}}")
//)

//func (b *Binding) processMustaches(elem dom.Selection, once bool, bs *bindScope) error {
//	text := elem.Text()
//	if strings.Index(text, "{{") == -1 {
//		return nil
//	}

//	matches := MustacheRegex.FindAllStringSubmatch(text, -1)
//	if matches != nil {
//		splitted := MustacheRegex.Split(text, -1)

//		textNodes := elem.NewEmptySelection()
//		for i, m := range matches {
//			cr, blist, v, err := bs.evaluate(m[1])
//			if err != nil {
//				return err
//			}

//			node := elem.NewTextNode(toString(v))

//			if !once {
//				err = b.watchModel(v, blist, cr, bs, func(val interface{}) {
//					node.SetText(toString(val))
//				})

//				if err != nil {
//					return err
//				}
//			}

//			if splitted[i] != "" {
//				bf := elem.NewTextNode(splitted[i])
//				textNodes = textNodes.Add(bf)
//			}

//			textNodes = textNodes.Add(node)
//		}

//		if splitted[len(splitted)-1] != "" {
//			bf := elem.NewTextNode(splitted[len(splitted)-1])
//			textNodes = textNodes.Add(bf)
//		}

//		elem.ReplaceWith(textNodes)
//	}

//	return nil
//}

func (b *Binding) bindRec(node *VNode,
	scope *Scope,
	additionalBinds []Bindage) {
	if node.Type == DeadNode {
		return
	}

	if node.scope != nil {
		scope = node.scope
	}

	tagName := node.TagName()

	var cv *componentView
	isComponent := false

	isElement := tagName != ""

	if isElement {
		cv, isComponent = b.tm.GetComponent(tagName)
	}

	if node.Type == GhostNode {
		for i, _ := range node.Children {
			b.bindRec(&node.Children[i], scope, node.Binds)
		}
		return
	}

	if node.Type == MustacheNode {
		err := b.processMustache(node, scope)
		if err != nil {
			bstrPanic(err.Error(), node.Data, node)
		}

		return
	}

	// perform binding
	var err error
	for _, bind := range append(additionalBinds, node.Binds...) {
		astr, bstr := bind.Name, bind.Expr

		switch bind.Type {
		case AttrBind:
			err = b.processAttrBind(astr, bstr, node, scope)

		case BinderBind:
			err = b.processBinderBind(astr, bstr, node, scope)
		default:
			panic("Invalid bind type.")
		}

		if err != nil {
			bstrPanic(err.Error(), bstr, node)
		}
	}

	if node.Type == DeadNode {
		return
	}

	if isComponent {
		err := b.bindComponent(node, scope, cv)
		if err != nil {
			panic(err)
		}

		return
	}

	for i, _ := range node.Children {
		b.bindRec(&node.Children[i], scope, []Bindage{})
	}

	return
}

func (b *Binding) newModelScope(model interface{}) *Scope {
	s := NewModelScope(model)
	s.Merge(b.scope)
	return s
}

func ScopeFromModels(models []interface{}) (s *Scope) {
	s = NewScope([]SymbolTable{})
	for _, model := range models {
		if model != nil {
			s.AddSymTables(NewModelSymbolTable(model))
		}
	}

	return
}

func (b *Binding) BindModels(rootElem *VNode, models []interface{}) {
	s := ScopeFromModels(models)
	s.Merge(b.scope)

	b.bindWithScope(rootElem, s)
}

func (b *Binding) Bind(rootNode *VNode, model interface{}) {
	b.BindModels(rootNode, []interface{}{model})
}

func (b *Binding) bindWithScope(rootNode *VNode, s *Scope) {
	b.bindRec(rootNode, s, []Bindage{})
}
