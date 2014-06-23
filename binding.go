package wade

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/gopherjs/gopherjs/js"
	jq "github.com/gopherjs/jquery"
)

type UpdateDomFn func(elem jq.JQuery, value interface{}, arg []string)
type BindDomFn func(elem jq.JQuery, value interface{}, arg, outputs []string)
type ModelUpdateFn func(value string)
type WatchDomFn func(elem jq.JQuery, updateFn ModelUpdateFn)

type DomBinder struct {
	update UpdateDomFn
	watch  WatchDomFn
	bind   BindDomFn
}

type binding struct {
	wade       *Wade
	domBinders map[string]DomBinder
	helpers    map[string]interface{}
}

func newBindEngine(wade *Wade) *binding {
	return &binding{
		wade:       wade,
		domBinders: defaultBinders(),
		helpers:    defaultHelpers(),
	}
}

//getReflectField returns the field value of an object, be it a struct instance
//or a map
func getReflectField(o reflect.Value, field string) (reflect.Value, error) {
	if o.Kind() == reflect.Ptr {
		o = o.Elem()
	}

	var rv reflect.Value
	switch o.Kind() {
	case reflect.Struct:
		rv = o.FieldByName(field)
		if !rv.IsValid() {
			rv = o.Addr().MethodByName(field)
		}
	case reflect.Map:
		rv = o.MapIndex(reflect.ValueOf(field))
	default:
		return rv, fmt.Errorf("Unhandled type for accessing %v", field)
	}

	if !rv.IsValid() {
		return rv, fmt.Errorf("No such field %v", field)
	}

	//if !rv.CanSet() {
	//	panic("Unaddressable")
	//}

	return rv, nil
}

type TokenType int

const (
	ExprToken TokenType = 1
	PuncToken           = 2
)

type Token struct {
	kind TokenType
	v    string
}

type Expr struct {
	name string
	args []*Expr
	eval *ObjEval
}

func isValidExprChar(c rune) bool {
	return c == '.' || unicode.IsLetter(c) || unicode.IsDigit(c)
}

//tokenize simply splits the bind target string syntax into expressions (SomeObject.SomeField) and punctuations (().,), making
//it a little bit easier to parse
func tokenize(spec string) (tokens []Token, err error) {
	tokens = make([]Token, 0)
	err = nil
	var token string
	flush := func() {
		if token != "" {
			if strings.HasPrefix(token, ".") || strings.HasSuffix(token, ".") {
				err = errors.New("Invalid '.'")
				return
			}
			tokens = append(tokens, Token{ExprToken, token})
		}
		token = ""
	}
	for _, c := range spec {
		switch c {
		case ' ':
			if token != "" {
				err = errors.New("Invalid space")
				return
			}
		case '(', ')', ',':
			flush()
			tokens = append(tokens, Token{PuncToken, string(c)})
		default:
			if isValidExprChar(c) {
				token += string(c)
			} else {
				err = fmt.Errorf("Character '%v' is not allowed", c)
				return
			}
		}
	}
	flush()

	return
}

//parse parses the bind target string, populate information into a tree of Expr pointers.
//Each helper call has a list arguments, each argument may be another helper call or an object expression.
func parse(spec string) (root *Expr, err error) {
	tokens, err := tokenize(spec)
	if err != nil {
		bindStringPanic(err.Error(), spec)
	}
	invalid := func() {
		err = errors.New("Invalid syntax")
	}
	if len(tokens) == 0 {
		err = errors.New("Empty bind string")
	}
	if tokens[0].kind != ExprToken {
		invalid()
		return
	}
	stack := make([]*Expr, 0)
	exprOf := make([]*Expr, len(tokens))
	root = &Expr{tokens[0].v, make([]*Expr, 0), nil}
	exprOf[0] = root
	var parent *Expr = nil
	for ii, token := range tokens[1:] {
		i := ii + 1 //i starts from 1 instead of 1, more intuitive
		switch token.v {
		case "(":
			if tokens[i-1].kind != ExprToken {
				invalid()
				return
			}
			parent = exprOf[i-1]
			stack = append(stack, parent)
		case ")":
			if parent == nil {
				invalid()
				return
			}
			stack = stack[:len(stack)-1]

		case ",":
			if !(tokens[i-1].kind == ExprToken || tokens[i-1].v == ")") {
				invalid()
				return
			}
		//expression
		default:
			expr := &Expr{tokens[i].v, make([]*Expr, 0), nil}
			exprOf[i] = expr
			if len(stack) == 0 {
				invalid()
				return
			}
			stack[len(stack)-1].args = append(stack[len(stack)-1].args, expr)
		}
	}

	return
}

type ObjEval struct {
	fieldRefl reflect.Value
	modelRefl reflect.Value
	field     string
}

//evaluateRec recursively evaluates the parsed expressions and return the result value, it also
//populates the tree of Expr with the value evaluated with evaluateObj if not available
func (b *binding) evaluateRec(expr *Expr, model interface{}) (v reflect.Value, err error) {
	err = nil
	if len(expr.args) == 0 {
		if expr.eval == nil {
			expr.eval, err = evaluateObj(expr.name, model)
			if err != nil {
				return
			}
		}
		v = expr.eval.fieldRefl
		return
	}

	if helper, ok := b.helpers[expr.name]; ok {
		args := make([]reflect.Value, len(expr.args))
		for i, e := range expr.args {
			args[i], err = b.evaluateRec(e, model)
		}
		if reflect.TypeOf(helper).NumIn() != len(args) {
			println()
			err = fmt.Errorf(`Invalid number of arguments to helper "%v"`, expr.name)
			return
		}
		v = reflect.ValueOf(helper).Call(args)[0]
		return
	}

	err = fmt.Errorf(`Invalid helper "%v".`, expr.name)
	return
}

func bindStringPanic(mess, bindstring string) {
	panic(fmt.Sprintf(mess+", while processing bind string `%v`.", bindstring))
}

//evaluateBindstring evaluates the bind string, returns the needed information for binding
func (b *binding) evaluateBindString(spec string, model interface{}) (root *Expr, blist []*Expr, value interface{}) {
	var err error
	root, err = parse(spec)
	if err != nil {
		bindStringPanic(err.Error(), spec)
	}
	v, err := b.evaluateRec(root, model)
	if err != nil {
		bindStringPanic(err.Error(), spec)
	}
	value = v.Interface()
	blist = make([]*Expr, 0)
	getBindList(root, &blist)
	return
}

//getBindList fetches the list of objects that need to be bound from the *Expr tree into a list
func getBindList(expr *Expr, list *([]*Expr)) {
	if len(expr.args) == 0 {
		*list = append(*list, expr)
		return
	}

	for _, e := range expr.args {
		getBindList(e, list)
	}
}

//evaluateObj uses reflection to access the field hierarchy in an object string
//and return the necessary values
func evaluateObj(obj string, model interface{}) (*ObjEval, error) {
	flist := strings.Split(obj, ".")
	vals := make([]reflect.Value, len(flist)+1)
	o := reflect.ValueOf(model)
	if o.Kind() == reflect.Ptr {
		o = o.Elem()
	}
	vals[0] = o
	var err error
	for i, field := range flist {
		o, err = getReflectField(o, field)
		if err != nil {
			return nil, err
		}
		vals[i+1] = o
	}

	return &ObjEval{
		fieldRefl: vals[len(vals)-1],
		modelRefl: vals[len(vals)-2],
		field:     flist[len(flist)-1],
	}, nil
}

func (b *binding) watch(binds []*Expr, root *Expr, model interface{}, callback func(interface{})) {
	for _, expr := range binds {
		//use watchjs to watch for changes to the model
		(func(expr *Expr) {
			js.Global.Call("watch",
				expr.eval.modelRefl.Interface(),
				expr.eval.field,
				func(prop string, action string,
					newVal interface{},
					oldVal js.Object) {

					expr.eval.fieldRefl.Set(reflect.ValueOf(newVal))
					newResult, _ := b.evaluateRec(root, model)
					callback(newResult.Interface())
				})
		})(expr)
	}
}

//bind parses the bind string, binds the element with a model
func (b *binding) Bind(relem jq.JQuery, model interface{}) {
	if relem.Length == 0 {
		panic("Incorrect element for bind.")
	}

	relem.Find("*").Each(func(i int, elem jq.JQuery) {
		_, isCustag := b.wade.custags[elem.Prop("tagName").(string)]

		attrs := elem.Get(0).Get("attributes")
		for i := 0; i < attrs.Length(); i++ {
			attr := attrs.Index(i)
			name := attr.Get("name").Str()
			bstr := attr.Get("value").Str()
			if name == "bind" { //attribute binding
				if !isCustag {
					panic(fmt.Sprintf("Attribute binding syntax can only be used for custom elements."))
				}
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
					value := strings.TrimSpace(fv[1])
					for _, c := range field {
						if !isValidExprChar(c) {
							bindStringPanic(fmt.Sprintf("invalid character %v", c), field)
						}
					}
					roote, binds, v := b.evaluateBindString(value, model)

					//if tModel == nil {
					//	tModelm[field] = v
					//	(func(field string) {
					//		b.watch(binds, roote, model, func(newResult interface{}) {
					//			tModelm[field] = newResult
					//		})
					//	})(field)
					//} else {
					tModel := b.wade.modelForCustomElem(elem)
					oe, err := evaluateObj(field, tModel)
					if err != nil {
						bindStringPanic("custom tag attribute check: "+err.Error(), bstr)
					}
					oe.fieldRefl.Set(reflect.ValueOf(v))
					b.watch(binds, roote, model, func(newResult interface{}) {
						oe.fieldRefl.Set(reflect.ValueOf(newResult))
					})
					//}
				}
			} else if strings.HasPrefix(name, BindPrefix) { //dom binding
				if isCustag {
					panic(`Dom binding is not allowed for custom element tags (they should not actually be rendered
, so there's no point; but of course inside the custom element's contents it's allowed normally).
If you want to bind the attributes of a custom element, use the field binding syntax instead.`)
				}

				parts := strings.Split(name, "-")
				if len(parts) <= 1 {
					panic(`Illegal "bind-".`)
				}
				if binder, ok := b.domBinders[parts[1]]; ok {
					args := make([]string, 0)
					if len(parts) >= 2 {
						for _, part := range parts[2:] {
							args = append(args, part)
						}
					}

					roote, binds, v := b.evaluateBindString(bstr, model)

					if binder.watch != nil {
						if len(binds) > 1 {
							panic(`Cannot use two-way data binding with multiple bound objects, it's illogical.
Consider using a one-way binder instead or use only 1 argument object in the bind string.`)
						}
						fmodel := binds[0].eval.fieldRefl
						binder.watch(elem, func(newVal string) {
							if !fmodel.CanSet() {
								panic("Cannot set field.")
							}
							fmodel.Set(reflect.ValueOf(newVal))
						})
					}

					(func(args []string) {
						if binder.bind != nil {
							binder.bind(elem, v, args, make([]string, 0))
						}
						if binder.update != nil {
							binder.update(elem, v, args)
							b.watch(binds, roote, model, func(newResult interface{}) {
								//println(newResult)
								//println(bstr)
								binder.update(elem,
									newResult,
									args)
							})
						}
					})(args)
				} else {
					panic(fmt.Sprintf(`Dom binder "%v" does not exist.`, binder))
				}
			}
		}
	})
}
