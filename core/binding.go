package core

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/phaikawl/wade/scope"
	"github.com/phaikawl/wade/utils"
)

const (
	AttrBindPrefix    = '@'
	BinderBindPrefix  = '#'
	SpecialAttrPrefix = "!"
	BelongAttrName    = SpecialAttrPrefix + "belong"
	GroupAttrName     = SpecialAttrPrefix + "group"
)

type (
	Binding struct {
		tm                *ComManager
		binders           map[string]Binder
		UniversalSymtable map[string]interface{}
	}

	BindingError struct {
		Err error
	}
)

func (be BindingError) Error() string {
	return be.Err.Error()
}

func NewBindEngine(tempConv templateConverter, universalSymtable map[string]interface{}) *Binding {
	b := &Binding{
		tm:                NewComManager(tempConv),
		binders:           map[string]Binder{},
		UniversalSymtable: universalSymtable,
	}

	return b
}

func (b *Binding) NewScope(models ...interface{}) scope.Scope {
	return scope.NewScope(append(models, map[string]interface{}{
		"$": b.UniversalSymtable,
	})...)
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

func bstrPanic(mess, bindstring string, node *VNode) {
	panic(fmt.Sprintf(mess+` - while processing bind string "%v".`, bindstring))
}

func reportError(err error, bstr string, elem *VNode) {
	if err != nil {
		bstrPanic(err.Error(), bstr, elem)
	}
}

func (b *Binding) processAttrBind(attr, bstr string, node *VNode, bs bindScope) (err error) {
	_, node.Attrs[attr], err = bs.evaluate(bstr)

	node.addCallback(func() (err error) {
		_, node.Attrs[attr], err = bs.evaluate(bstr)
		return
	})

	return
}

func (b *Binding) processMustache(node *VNode, bs bindScope) (err error) {
	_, v, err := bs.evaluate(node.Binds[0].Expr)
	node.Data = utils.ToString(v)

	node.addCallback(func() (err error) {
		_, v, err := bs.evaluate(node.Binds[0].Expr)
		node.Data = utils.ToString(v)
		return
	})

	return
}

func (b *Binding) processFieldBind(field string, bstr string, node *VNode, bs bindScope, ci *componentInstance) (err error) {
	_, v, err := bs.evaluate(bstr)
	if err != nil {
		return
	}

	oe, ok, err := scope.EvaluateObjField(field, reflect.ValueOf(ci.model))
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
		_, v, err := bs.evaluate(bstr)
		if err != nil {
			return
		}

		oe.FieldRefl.Set(reflect.ValueOf(v))

		return
	})

	return
}

func (b *Binding) bindComponent(node *VNode, bs bindScope, cv *componentView) (err error) {
	ci, err := cv.NewInstance(node)
	if err != nil {
		return fmt.Errorf(`Failed initialization of the component instance, error in its Init(). Error: %v.`, err.Error())
	}

	for _, bind := range node.Binds {
		if bind.Type == AttrBind {
			//bindinfo += fmt.Sprintf("{%v: [%v]} ", hattr.Name, hattr.Value)

			field := strings.Split(bind.Name, ".")[0]
			b.processAttrBind(bind.Name, bind.Expr, node, bs)
			if ok, fieldName := cv.HasAttr(field); ok {
				b.processFieldBind(fieldName, bind.Expr, node, bs, ci)
			}
		}
	}

	ci.origNode.scope = &bs.Scope
	b.bindWithScope(&ci.origNode, bs.Scope)

	for i, _ := range node.Children {
		b.Bind(&node.Children[i], ci.model)
	}

	node.addCallback(func() (err error) {
		ci.processUpdate()
		return
	})

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

func (b *Binding) processBinderBind(astr, bstr string, node *VNode, bs bindScope) (err error) {
	binderName, args, err := parseBinderLHS(astr)
	if err != nil {
		return
	}

	binder, ok := b.binders[binderName]
	if !ok {
		err = fmt.Errorf(`Binder "%v" does not exist`, binderName)
		return
	}

	ok, required := binder.CheckArgsNo(len(args))
	if !ok {
		err = fmt.Errorf(`Invalid number of arguments for the "%v" binder. Given %v, required %v.`, binderName, len(args), required)
	}

	if mb, ok := binder.(MutableBinder); ok {
		binder = mb.NewInstance()
	}

	binder.BeforeBind(&bs)

	roote, v, err2 := bs.evaluate(bstr)
	if err2 != nil {
		err = err2
		return
	}

	domBind := DomBind{
		Node:     node,
		Value:    v,
		Args:     args,
		BindName: astr,
		binding:  b,
		scope:    bs.Scope,
	}

	if tw, ok := binder.(TwoWayBinder); ok {
		if roote.typ != ValueExpr {
			err = fmt.Errorf("No function call is allowed in an expression for 2-way binding.")
			return
		}

		var sym scope.ScopeSymbol
		sym, err = bs.Lookup(roote.name)
		if err != nil {
			return
		}

		tw.Listen(domBind, func(newVal string) {
			bdb, ok := sym.(scope.Bindable)
			if !ok {
				bstrPanic("2-way data binding on unchangable field", bstr, node)
			}

			bdb.BindObj().FieldRefl.Set(reflect.ValueOf(newVal))
		})
	}

	//gopherjs:blocking
	binder.Bind(domBind)

	node.addCallback(func() (err error) {
		//gopherjs:blocking
		_, domBind.Value, _ = bs.evaluate(bstr)
		binder.Update(domBind)
		return
	})

	return
}

func (b *Binding) bindRec(node *VNode,
	scope scope.Scope) {

	if node.Type == DeadNode {
		return
	}

	if node.scope != nil {
		scope = *node.scope
	}

	tagName := node.TagName()

	var cv *componentView
	isComponent := false

	isElement := tagName != ""

	if isElement {
		cv, isComponent = b.tm.GetComponent(tagName)
	}

	bs := bindScope{scope}

	if node.Type == MustacheNode {
		err := b.processMustache(node, bs)
		if err != nil {
			bstrPanic(err.Error(), node.Data, node)
		}

		return
	}

	// perform binding
	var err error
	for _, bind := range node.Binds {
		astr, bstr := bind.Name, bind.Expr

		switch bind.Type {
		case AttrBind:
			if isComponent {
				continue
			}

			err = b.processAttrBind(astr, bstr, node, bs)

		case BinderBind:
			err = b.processBinderBind(astr, bstr, node, bs)
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
		err := b.bindComponent(node, bindScope{scope}, cv)
		if err != nil {
			panic(err)
		}

		return
	}

	for i, _ := range node.Children {
		b.bindRec(&node.Children[i], scope)
	}

	return
}

func (b *Binding) Bind(rootElem *VNode, models ...interface{}) {
	b.bindWithScope(rootElem, b.NewScope(models...))
}

func (b *Binding) bindWithScope(rootNode *VNode, s scope.Scope) {
	b.bindRec(rootNode, s)
}
