package main

import (
	"bytes"
	"fmt"
	"io"
	"text/template"
)

type (
	importTD struct {
		Path string
		Name string
	}

	preludeTD struct {
		Pkg     string
		Imports []importTD
	}

	stateFieldTD struct {
		Name, Type string
	}

	stateMethodsTD struct {
		Receiver, StateField, StateType string
		Setters                         []stateFieldTD
	}

	refFieldTD struct {
		Name, Type string
	}

	refsTD struct {
		ComName  string
		TypeName string
		Fields   []refFieldTD
	}

	fieldAssTD struct {
		Name, Value string
	}

	comInitFuncTD struct {
		ComType string
		Com     *comCreateTD
		Fields  []fieldAssTD
	}

	comCreateTD struct {
		ComName, ComType string
		Children, Attrs  *bytes.Buffer
	}

	comDefTD struct {
		ComName string
	}

	renderFuncTD struct {
		ComName string
		Return  *bytes.Buffer
		Decls   *bytes.Buffer
	}

	elementVDOMTD struct {
		Tag      string
		Key      string
		Attrs    map[string]string
		Children []*bytes.Buffer
	}

	textNodeVDOMTD struct {
		Text string
	}
)

const (
	childrenVDOMCode = `[[if .Children]]vdom.NewNodeList(
	[[$last := lastIdx .Children]]
	[[range $i, $c := .Children]]
	[[$c]][[if lt $i $last]],[[end]][[end]])[[else]]nil[[end]]`

	textNodeVDOMCode = `vdom.NewTextNode([[.Text]])`

	elementVDOMCode = `[[define "attrs"]]` +
		`[[if .Attrs]]` +
		`vdom.Attributes{` +
		`[[range $key, $value := .Attrs]]
			"[[$key]]": [[$value]],
		[[end]]` +
		`}[[else]]nil[[end]]` +
		`[[end]]` +
		`vdom.NewElement("[[.Tag]]", [[.Key]], [[template "attrs" .]],` +
		`[[template "children" .]])`

	renderFuncCode = `
func [[if .ComName]](this *[[.ComName]])[[end]] Render() *vdom.Element {
	[[.Decls]]
	return [[.Return]]
}
`

	preludeCode = `package [[.Pkg]]

// THIS FILE IS AUTOGENERATED BY WADE.GO FUEL
// CHANGES WILL BE OVERWRITTEN
import (
	"fmt"

	"github.com/gowade/wade/vdom"
	//"github.com/gowade/wade"
[[range .Imports]]
	[[.Path]][[.Name]]
[[end]]
)

func init() {
	_, _ = fmt.Printf, vdom.NewElement
}
`

	stateMethodsCode = `
[[ $receiver := .Receiver ]]
[[ $stateField := .StateField ]]

func (this [[$receiver]]) InternalState() interface{} {
	return this.[[$stateField]]
}

func (this [[$receiver]]) InternalInitState(stateData interface{}) {
	if stateData != nil {
		this.[[$stateField]] = stateData.(*[[.StateType]])
	} else {
		if this.[[$stateField]] == nil {
			var t [[.StateType]]
			this.[[$stateField]] = &t
		}
	}
}

[[range .Setters]]
func (this [[$receiver]]) set[[.Name]](v [[.Type]]) {
	this.[[$stateField]].[[.Name]] = v
	this.Rerender()
}
[[end]]
`

	refsCode = `
type [[.TypeName]] struct {
[[range .Fields]]
	[[.Name]] [[.Type]]
[[end]]
}

func (this *[[.ComName]]) Refs() [[.TypeName]] {
	return this.Com.InternalRefsHolder.([[.TypeName]])
}`

	rerenderMethodCode = `
func (this *%v) Rerender() {
	if vdom.InternalRenderLocked() {
		return
	}

	r := this.Render(nil)
	vdom.PerformDiff(r, this.VNode.Render().(*vdom.Element), this.VNode.DOMNode())
	this.VNode.ComRend = r
	this.VNode = r
}
`

	comInitCode = `
func (i interface{}) {
	com := i.(*[[.ComType]])
	com.Com = [[template "comCreate" .Com]]
[[range .Fields]]
	com.[[.Name]] = [[.Value]]
[[end]]
}`

	comCreateCode = `
wade.Com{
	ComponentName: "[[.ComName]]",
	InternalRefsHolder: [[.ComType]]Refs{},
	Children: [[.Children]],
	Attrs: [[.Attrs]],
}`

	comDefCode = `
type [[.ComName]] struct { wade.Com }
`
)

func newTpl(name string, code string) *template.Template {
	return template.Must(gTpl.New(name).Parse(code))
}

var (
	funcMap = template.FuncMap{
		"lastIdx": func(l []*bytes.Buffer) int {
			return len(l) - 1
		},
	}
	gTpl            = template.New("root").Delims("[[", "]]").Funcs(funcMap)
	childrenVDOMTpl = newTpl("children", childrenVDOMCode)
	textNodeVDOMTpl = newTpl("txvdom", textNodeVDOMCode)
	elementVDOMTpl  = newTpl("elvdom", elementVDOMCode)
	renderFuncTpl   = newTpl("renderFunc", renderFuncCode)
	preludeTpl      = newTpl("prelude", preludeCode)
	stateMethodsTpl = newTpl("stateMethods", stateMethodsCode)
	refsTpl         = newTpl("refs", refsCode)
	comInitFuncTpl  = newTpl("comInit", comInitCode)
	comCreateTpl    = newChildTpl(comInitFuncTpl, "comCreate", comCreateCode)
	comDefTpl       = newTpl("comDef", comDefCode)
)

func writeRerenderMethod(w io.Writer, comName string) {
	fmt.Fprintf(w, rerenderMethodCode, comName)
}

func newChildTpl(parent *template.Template, name, code string) *template.Template {
	return template.Must(parent.New(name).Parse(code))
}

func addChildTpl(parent *template.Template, name string, child *template.Template) *template.Template {
	return template.Must(parent.AddParseTree(name, child.Tree))
}
