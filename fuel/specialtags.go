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

type specialTagFunc func(io.Writer, *html.Node, *declArea) error

func (z *htmlCompiler) specialTag(tagName string) specialTagFunc {
	switch tagName {
	case forSTag:
		return z.forTagGenerate
	}

	return nil
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

var (
	varDeclCode = `
		[[range .Vars]]
			[[.Code]]
		[[end]]
	`

	forTagVDOMCode = `
	var [[.VarName]] []vdom.Node
	for __k, __v := range [[.Items]] {
		[[if .KeyName]] [[.KeyName]] := __k [[else]] _ = __k [[end]]
		[[if .ValName]] [[.ValName]] := __v [[else]] _ = __v [[end]]

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
	// process the attributes
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

	// declare a variable to hold this loops's list of nodes inside a parent Declaration Area
	varName := sfmt("for%v", exprApproxName(rangeAttr.Val))
	varName, cbuf := da.declare(varName)

	// create a Declaration Area so that
	// control structures (e.g an if tag) nested inside this one
	// could declare variables
	newDA := newDeclArea(da)
	children, err := z.childrenGenerate(n, newDA)
	if err != nil {
		return err
	}

	return forTagVDOMTpl.Execute(cbuf, forTagVDOMTD{
		KeyName:  keyName,
		ValName:  valName,
		VarName:  varName,
		Items:    attributeValueCode(*rangeAttr),
		Children: children,
		Decls:    newDA.code(),
	})
}
