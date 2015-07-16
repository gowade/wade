package main

import (
	"bytes"
	"fmt"
	"io"
	"text/template"
)

func newTpl(name string, code string) *template.Template {
	return template.Must(template.New(name).Parse(code))
}

type stateFieldTD struct {
	Name, Type string
}

type stateMethodsTD struct {
	Receiver, StateField, StateStruct, StateType string
	StateIsPointer                               bool
	Setters                                      []stateFieldTD
}

var stateMethodsTpl = newTpl("stateMethods", `
{{ $receiver := .Receiver }}
{{ $stateField := .StateField }}

func (this {{$receiver}}) InternalState() interface{} {
	return this.{{$stateField}}
}

func (this {{$receiver}}) InternalInitState(stateData interface{}) {
	if stateData != nil {
		this.{{$stateField}} = stateData.({{.StateType}})
	}{{if .StateIsPointer}} else {
		if this.{{$stateField}} == nil {
			this.{{$stateField}} = &{{.StateStruct}}{}
		}
	}{{end}}
}

{{range .Setters}}
func (this {{$receiver}}) set{{.Name}}(v {{.Type}}) {
	this.{{$stateField}}.{{.Name}} = v
	this.Rerender()
}
{{end}}
`)

type refFieldTD struct {
	Name, Type string
}

type refsTD struct {
	ComName  string
	TypeName string
	Fields   []refFieldTD
}

var refsTpl = newTpl("refs", `
type {{.TypeName}} struct {
{{range .Fields}}
	{{.Name}} {{.Type}}
{{end}}
}

func (this *{{.ComName}}) Refs() {{.TypeName}} {
	return this.Com.InternalRefsHolder.({{.TypeName}})
}
`)

func writeRerenderMethod(w io.Writer, comName string) {
	fmt.Fprintf(w, rerenderMethodTpl, comName)
}

var rerenderMethodTpl = `
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

type fieldAssTD struct {
	Name, Value string
}

type comInitFuncTD struct {
	ComType string
	Com     *comCreateTD
	Fields  []fieldAssTD
}

type comCreateTD struct {
	ComName, ComType string
	Children, Attrs  *bytes.Buffer
}

var comInitFuncTpl = newTpl("comInit", `func (i interface{}) {
com := i.(*{{.ComType}})
com.Com = {{template "comCreate" .Com}}
{{range .Fields}}
	com.{{.Name}} = {{.Value}}
{{end}}
}`)

func init() {
	template.Must(comInitFuncTpl.New("comCreate").Parse(`wade.Com{
	ComponentName: "{{.ComName}}",
	InternalRefsHolder: {{.ComType}}Refs{},
	Children: {{.Children}},
	Attrs: {{.Attrs}},
}`))
}
