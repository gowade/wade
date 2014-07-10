package bind

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/gopherjs/gopherjs/js"
	jq "github.com/gopherjs/jquery"
)

var (
	gJQ = jq.NewJQuery
)

const (
	BindPrefix         = "bind-"
	ReservedBindPrefix = "wade-rsvd-bound"
)

func toString(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

// DomBinder is the common interface for Dom binders.
type DomBinder interface {
	// Update is called whenever the model's field changes, to perform
	// dom updating, like setting the html content or setting
	// an html attribute for the elem
	Update(elem jq.JQuery, value interface{}, arg, outputs []string)

	// Bind is similar to Update, but is called only once at the start, when
	// the bind is being processed
	Bind(b *Binding, elem jq.JQuery, value interface{}, arg, outputs []string)

	// Watch is used in 2-way binders, it watches the html element for changes
	// and updates the model field accordingly
	Watch(elem jq.JQuery, updateFn ModelUpdateFn)

	// BindInstance is useful for binders that need to save some data for each
	// separate element. This method returns an instance of the binder to be used.
	BindInstance() DomBinder
}

type ModelUpdateFn func(value string)

type CustomElemManager interface {
	IsCustomElem(jq.JQuery) bool
	ModelForElem(jq.JQuery) interface{}
}

type Binding struct {
	tm         CustomElemManager
	domBinders map[string]DomBinder
	helpers    map[string]interface{}
}

func NewBindEngine(tm CustomElemManager) *Binding {
	return &Binding{
		tm:         tm,
		domBinders: defaultBinders(),
		helpers:    defaultHelpers(),
	}
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

	if _, exist := b.helpers[name]; !exist {
		b.helpers[name] = fn
		return
	}
	panic(fmt.Sprintf("Helper with name %v already exists.", name))
	return
}

// Delete a helper
func (b *Binding) DeleteHelper(name string) {
	if _, ok := b.helpers[name]; ok {
		delete(b.helpers, name)
	}
	panic(fmt.Sprintf("No such helper %v", name))
}

func jqExists(elem jq.JQuery) bool {
	return elem.Parents("html").Length > 0
}

// getReflectField returns the field value of an object, be it a struct instance
// or a map
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
		return rv, fmt.Errorf(`Unhandled type %v for accessing "%v"`, o.Type().Name(), field)
	}

	if !rv.IsValid() {
		return rv, fmt.Errorf(`No such field "%v" in %+v`, field, o.Interface())
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
	return c == '`' || c == '.' || c == '_' || unicode.IsLetter(c) || unicode.IsDigit(c)
}

// tokenize simply splits the bind target string syntax into expressions (SomeObject.SomeField) and punctuations (().,), making
// it a little bit easier to parse
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
	strlitMode := false //string literal mode
	for _, c := range spec {
		if !strlitMode {
			switch c {
			case ' ':
				if token != "" {
					err = errors.New("Invalid space")
					return
				}
			case '(', ')', ',':
				flush()
				tokens = append(tokens, Token{PuncToken, string(c)})
			case '`':
				strlitMode = true
				token += string(c)
			default:
				if isValidExprChar(c) {
					token += string(c)
				} else {
					err = fmt.Errorf("Character '%q' is not allowed", c)
					return
				}
			}
		} else {
			if c == '`' {
				strlitMode = false
			} else if !unicode.IsDigit(c) && !unicode.IsLetter(c) && !strings.ContainsRune(",(-_.)", c) {
				err = fmt.Errorf("Use of characters other than numbers, " +
					"letters, parentheses ('(', ')'), dash ('-'), comma (','), " +
					"underscore ('_'), and dot ('.') is forbidden " +
					"inside string literals of bind string, " +
					"heavy processing and logic should not be in html template. Consider " +
					"moving your data to the model instead of putting it into the bind string.")
				return
			}
			token += string(c)
		}
	}
	flush()

	return
}

// parse parses the bind target string, populate information into a tree of Expr pointers.
// Each helper call has a list arguments, each argument may be another helper call or an object expression.
func parse(spec string) (root *Expr, err error) {
	tokens, err := tokenize(spec)
	if err != nil {
		return
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

type Value struct {
	oe    *ObjEval
	value interface{}
}

func (val *Value) Set(v reflect.Value) {
	if val.oe != nil {
		val.oe.fieldRefl.Set(v)
		return
	}
	panic("Unexpected, cannot set value for literal expression.")
}

func (val *Value) Val() reflect.Value {
	if val.oe != nil {
		return val.oe.fieldRefl
	}
	return reflect.ValueOf(val.value)
}

// evaluateRec recursively evaluates the parsed expressions and return the result value, it also
// populates the tree of Expr with the value evaluated with evaluateObj if not available
func (b *Binding) evaluateRec(expr *Expr, model interface{}) (v reflect.Value, err error) {
	err = nil
	if len(expr.args) == 0 {
		var val *Value
		val, err = evaluateExpr(expr.name, model)
		if err != nil {
			return
		}
		expr.eval = val.oe
		v = val.Val()
		return
	}

	if helper, ok := b.helpers[expr.name]; ok {
		args := make([]reflect.Value, len(expr.args))
		for i, e := range expr.args {
			args[i], err = b.evaluateRec(e, model)
			if err != nil {
				return
			}
		}
		ftype := reflect.TypeOf(helper)
		nin := ftype.NumIn()
		var ok bool
		if ftype.IsVariadic() {
			ok = len(args) >= nin-1
		} else {
			ok = nin == len(args)
		}
		if !ok {
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
	panic(fmt.Sprintf(mess+`, while processing bind string "%v".`, bindstring))
}

// evaluateBindstring evaluates the bind string, returns the needed information for binding
func (b *Binding) evaluate(spec string, model interface{}) (root *Expr, blist []*Expr, value interface{}, err error) {
	root, err = parse(spec)
	if err != nil {
		return
	}
	v, err := b.evaluateRec(root, model)
	if err != nil {
		return
	}
	value = v.Interface()
	blist = make([]*Expr, 0)
	getBindList(root, &blist)
	return
}

func (b *Binding) evaluateBindString(bstr string, model interface{}) (root *Expr, blist []*Expr, value interface{}) {
	var err error
	root, blist, value, err = b.evaluate(bstr, model)
	if err != nil {
		bindStringPanic(err.Error(), bstr)
	}
	return
}

// getBindList fetches the list of objects that need to be bound from the *Expr tree into a list
func getBindList(expr *Expr, list *([]*Expr)) {
	if expr == nil {
		return
	}

	if len(expr.args) == 0 && expr.eval != nil {
		*list = append(*list, expr)
		return
	}

	for _, e := range expr.args {
		getBindList(e, list)
	}
}

// evaluateObj uses reflection to access the field hierarchy in an object string
// and return the necessary values
func evaluateObj(obj string, model interface{}) (*ObjEval, error) {
	if obj != "" && model == nil {
		return nil, fmt.Errorf(`This page doens't have a model so we cannot bind to "%v"`, obj)
	}
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

func evaluateExpr(expr string, model interface{}) (v *Value, err error) {
	err = nil
	expr = strings.TrimSpace(expr)
	re := []rune(expr)
	numberMode := false
	floatMode := false
	for i, c := range expr {
		switch {
		case c == '`':
			if i == 0 { //string literal
				if re[len(expr)-1] == '`' {
					v = &Value{nil, string(re[1 : len(re)-1])}
					return
				}
				err = fmt.Errorf("No matching quote.")
				return
			} else {
				err = fmt.Errorf("Invalid quote")
				return
			}
		case unicode.IsDigit(c):
			if i == 0 {
				numberMode = true
			}
		case unicode.IsLetter(c) || c == '_':
			if numberMode {
				err = fmt.Errorf("Invalid: dynamic expression cannot start with a number")
				return
			}
		case c == '.':
			if floatMode {
				err = fmt.Errorf("Multiple dot '.' for a number, invalid")
				return
			}
			if numberMode {
				floatMode = true
			}
		default:
			err = fmt.Errorf("Invalid character '%q'", c)
			return
		}
	}

	switch {
	case floatMode:
		var f float64
		f, err = strconv.ParseFloat(expr, 32)
		v = &Value{nil, float32(f)}
		return
	case numberMode:
		var i int
		i, err = strconv.Atoi(expr)
		v = &Value{nil, i}
		return
	default:
		var oe *ObjEval
		oe, err = evaluateObj(expr, model)
		v = &Value{oe, nil}
	}

	return
}

func jsGetType(obj js.Object) string {
	return js.Global.Get("Object").Get("prototype").Get("toString").Call("call", obj).Str()
}

func (b *Binding) watchModel(binds []*Expr, root *Expr, model interface{}, callback func(interface{})) {
	for _, expr := range binds {
		//use watchjs to watch for changes to the model
		//println(js.InternalObject(expr.eval.modelRefl.Interface()))
		(func(expr *Expr) {
			obj := js.InternalObject(expr.eval.modelRefl.Interface()).Get("$val")
			//workaround for gopherjs's protection disallowing js access to maps
			//setDummyHopFn(obj, "")
			js.Global.Call("watch",
				obj,
				expr.eval.field,
				func(prop string, action string,
					newVal interface{},
					oldVal js.Object) {
					//v = expr.eval.fieldRefl.Interface()
					newResult, _ := b.evaluateRec(root, model)
					callback(newResult.Interface())
				})
		})(expr)
	}
}

func (b *Binding) processDomBind(astr, bstr string, elem jq.JQuery, model interface{}, once bool) {
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
		roote, binds, v := b.evaluateBindString(bexpr, model)

		if len(binds) == 1 {
			fmodel := binds[0].eval.fieldRefl
			binder.Watch(elem, func(newVal string) {
				if !fmodel.CanSet() {
					panic("Cannot set field.")
				}
				fmodel.Set(reflect.ValueOf(newVal))
			})
		}

		(func(args, outputs []string) {
			binder.Bind(b, elem, v, args, outputs)
			binder.Update(elem, v, args, outputs)
			if !once {
				b.watchModel(binds, roote, model, func(newResult interface{}) {
					binder.Update(elem,
						newResult,
						args, outputs)
				})
			}
		})(args, outputs)
	} else {
		panic(fmt.Sprintf(`Dom binder "%v" does not exist.`, parts[1]))
	}
}

func (b *Binding) processAttrBind(astr, bstr string, elem jq.JQuery, model interface{}, once bool) {
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

		roote, binds, v := b.evaluateBindString(valuestr, model)

		tModel := b.tm.ModelForElem(elem)
		oe, err := evaluateObj(field, tModel)
		if err != nil {
			bindStringPanic("custom tag attribute check: "+err.Error(), bstr)
		}
		isCompat := func(src reflect.Type, dst reflect.Type) {
			if !src.AssignableTo(dst) {
				bindStringPanic(fmt.Sprintf("Unassignable, incompatible types %v and %v of the model field and the value",
					src.String(), dst.String()), bstr)
			}
		}
		isCompat(reflect.TypeOf(v), oe.fieldRefl.Type())
		oe.fieldRefl.Set(reflect.ValueOf(v))
		if !once {
			b.watchModel(binds, roote, model, func(newResult interface{}) {
				nr := reflect.ValueOf(newResult)
				isCompat(nr.Type(), oe.fieldRefl.Type())
				oe.fieldRefl.Set(nr)
			})
		}
	}
}

func preventBinding(elem jq.JQuery) {
	elem.SetAttr(ReservedBindPrefix, "t")
}

func PreventBinding(elem jq.JQuery) {
	elem.Find("*").Each(func(_ int, d jq.JQuery) {
		preventBinding(d)
	})
}

func bindingPrevented(elem jq.JQuery) bool {
	return elem.Attr(ReservedBindPrefix) == "t"
}

func wrapBindFunc(astr, bstr string, elem jq.JQuery, model interface{}, once bool,
	fn func(astr, bstr string, elem jq.JQuery, model interface{}, once bool)) func() {
	return func() {
		if !bindingPrevented(elem) {
			fn(astr, bstr, elem, model, once)
			preventBinding(elem)
		}
	}
}

// bind parses the bind string, make a list of binds (this doesn't actually bind the elements)
func (b *Binding) bindPrepare(relem jq.JQuery, model interface{}, once bool) (bindTasks []func()) {
	if relem.Length == 0 {
		panic("Incorrect element for bind.")
	}

	bindTasks = make([]func(), 0)

	relem.Children("*").Each(func(i int, elem jq.JQuery) {
		isCustag := b.tm.IsCustomElem(elem)

		htmla := elem.Get(0).Get("attributes")
		attrs := make(map[string]string)
		for i := 0; i < htmla.Length(); i++ {
			attr := htmla.Index(i)
			attrs[attr.Get("name").Str()] = attr.Get("value").Str()
		}

		for name, bstr := range attrs {
			if name == "bind" { //attribute binding
				if !isCustag {
					panic(fmt.Sprintf("Attribute binding syntax can only be used for registered custom elements."))
				}
				bindTasks = append(bindTasks,
					wrapBindFunc(name, bstr, elem, model, once, b.processAttrBind))

				continue
			} else if strings.HasPrefix(name, BindPrefix) && //dom binding
				jqExists(elem) { //element still exists
				if isCustag {
					panic(`Dom binding is not allowed for custom element tags (they should not actually be rendered
			, so there's no point; but of course inside the custom element's contents it's allowed normally).
			If you want to bind the attributes of a custom element, use the field binding syntax instead.`)
				}
				bindTasks = append(bindTasks,
					wrapBindFunc(name, bstr, elem, model, once, b.processDomBind))
			}
		}

		bindTasks = append(bindTasks, b.bindPrepare(elem, model, once)...)
	})

	return
}

// Bind binds a model to an element and all its children
func (b *Binding) Bind(relem jq.JQuery, model interface{}, once bool) {
	// we have to do 2 steps like this to avoid missing out binding when things are removed
	btasks := b.bindPrepare(relem, model, once)
	for _, fn := range btasks {
		fn()
	}
}
