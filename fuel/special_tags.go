package main

import (
	"bytes"
	"errors"
	"io"
	"strings"

	"github.com/gowade/html"
)

const (
	forSTag = "for"
)

func newDeclArea() *declArea {
	return &declArea{
		nameIdx: map[string]int{},
	}
}

type declCodeTD struct {
	VarName string
	Code    *bytes.Buffer
}

type declArea struct {
	nameIdx map[string]int
	decls   []declCodeTD
}

// adds a declaration to the declaration area, if varName is already taken,
// add a new number suffix to it, returns the new valid name
// and a new buffer for its code
func (z *declArea) declare(varName string) (string, *bytes.Buffer) {
	if z.nameIdx[varName] > 0 {
		z.nameIdx[varName]++
	}

	var buf bytes.Buffer
	newVarName := sfmt("%v%v", varName, z.nameIdx[varName])
	z.decls = append(z.decls, declCodeTD{
		VarName: newVarName,
		Code:    &buf,
	})

	return newVarName, &buf

}

func (z *declArea) code() *bytes.Buffer {
	buf, err := execTplBuf(varDeclTpl, varDeclTD{
		Vars: z.decls,
	})

	if err != nil {
		panic(err)
	}

	return buf
}

func nodeListCodeBuf() *bytes.Buffer {
	return bytes.NewBufferString("[]vdom.Node")
}

type (
	varDeclTD struct {
		Vars []declCodeTD
	}

	forTagVDOMTD struct {
		Items            string
		KeyName, ValName string
		VarName          string
		Decls            *bytes.Buffer
		Children         []*bytes.Buffer
	}
)

type specialTagFunc func(io.Writer, *html.Node, *declArea) error

func (z *htmlCompiler) specialTag(tagName string) specialTagFunc {
	switch tagName {
	case forSTag:
		return z.forTagGenerate
	}

	return nil
}

var (
	varDeclCode = `
		[[range .Vars]]
			[[.Code]]
		[[end]]
	`

	forTagVDOMCode = `
	var [[.VarName]] []vdom.Node
	for __k, __v := range [[.Items]] {
		[[if .KeyName]] [[.KeyName]] = __k [[else]] _ = __k [[end]]
		[[if .ValName]] [[.ValName]] = __v [[else]] _ = __v [[end]]

		[[.Decls]]
		[[.VarName]] = append([[.VarName]], [[template "children" .]]...)
	}`
)

var (
	varDeclTpl    = newTpl("varDecl", varDeclCode)
	forTagVDOMTpl = newTpl("forTag", forTagVDOMCode)
)

func exprApproxName(expr string) string {
	var buf bytes.Buffer
	for _, c := range expr {
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '.' {
			buf.WriteRune(c)
		}
	}

	split := strings.Split(buf.String(), ".")
	return strings.ToUpper(split[len(split)-1])
}

func fmtSTagError(specialTag, msg string) error {
	return errors.New(sfmt("special '%v' tag: %v", specialTag, msg))
}

func invalidAttribute(specialTag, attr string) error {
	return fmtSTagError(specialTag, sfmt("invalid attribute '%v'", attr))
}

func attrsRequireNotEmpty(specialTag string, attrs ...html.Attribute) error {
	for _, attr := range attrs {
		if attr.Val == "" || attr.IsEmpty {
			return fmtSTagError(specialTag,
				sfmt("attribute '%v' cannot be empty", attr.Key))
		}
	}

	return nil
}

func (z *htmlCompiler) forTagGenerate(w io.Writer, n *html.Node, da *declArea) error {
	var keyName, valName string
	var rangeAttr *html.Attribute
	for _, attr := range n.Attr {
		switch attr.Key {
		case "k":
			keyName = attr.Val
		case "v":
			valName = attr.Val
		case "range":
			rangeAttr = &attr
		default:
			return invalidAttribute(forSTag, attr.Key)
		}
	}

	if err := attrsRequireNotEmpty(forSTag, *rangeAttr); err != nil {
		return err
	}

	newDA := newDeclArea()
	children, err := z.childrenGenerate(n, newDA)
	if err != nil {
		return err
	}

	varName := sfmt("for%v", exprApproxName(rangeAttr.Val))
	varName, cbuf := da.declare(varName)

	return forTagVDOMTpl.Execute(cbuf, forTagVDOMTD{
		KeyName:  keyName,
		ValName:  valName,
		VarName:  varName,
		Items:    attributeValueCode(*rangeAttr),
		Children: children,
		Decls:    newDA.code(),
	})
}
