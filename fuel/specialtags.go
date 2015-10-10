package main

import (
	"bytes"
	"errors"
	"io"
	"strings"

	"github.com/gowade/whtml"
)

const (
	forSTag     = "for"
	ifSTag      = "if"
	switchSTag  = "switch"
	caseSTag    = "case"
	defaultSTag = "default"
)

type specialTagFunc func(io.Writer, *whtml.Node, *declArea, refsMap) error

func (z *htmlCompiler) specialTag(tagName string) specialTagFunc {
	switch tagName {
	case forSTag:
		return z.forTagGenerate
	case ifSTag:
		return z.ifTagGenerate
	case switchSTag:
		return z.switchTagGenerate
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
		Children         []childCode
	}

	ifTagVDOMTD struct {
		Cond     string
		VarName  string
		Decls    *bytes.Buffer
		Children []childCode
	}

	caseTagVDOMTD struct {
		Expr     string
		Children []childCode
		Decls    *bytes.Buffer
	}

	switchTagVDOMTD struct {
		VarName string
		Expr    string
		Cases   []*caseTagVDOMTD
		Default *caseTagVDOMTD
	}
)

var (
	varDeclCode = `
		[[range .Vars]]
			[[.Code]]
		[[end]]
	`

	forTagVDOMCode = `
	[[.VarName]] := []vdom.VNode{}
	for __k, __v := range [[.Items]] {
		[[if .KeyName]] [[.KeyName]] := __k [[else]] _ = __k [[end]]
		[[if .ValName]] [[.ValName]] := __v [[else]] _ = __v [[end]]

		[[.Decls]]
		[[.VarName]] = append([[.VarName]], [[template "children" .Children]]...)
	}`

	ifTagVDOMCode = `
	[[.VarName]] := []vdom.VNode{}
	if [[.Cond]] {
		[[.Decls]]
		[[.VarName]] = [[template "children" .Children]]
	}
	`

	switchTagVDOMCode = `
	[[.VarName]] := []vdom.VNode{}
	[[$varName := .VarName]]
	switch [[.Expr]] {
	[[range .Cases]]
	case [[.Expr]]:
		[[.Decls]]
		[[$varName]] = [[template "children" .Children]]	
	[[end]]
	[[if .Default]]
		[[$varName]] = [[template "children" .Default.Children]]
	[[end]]
	}
	`
)

var (
	varDeclTpl       = newTpl("varDecl", varDeclCode)
	forTagVDOMTpl    = newTpl("forTag", forTagVDOMCode)
	ifTagVDOMTpl     = newTpl("ifTag", ifTagVDOMCode)
	switchTagVDOMTpl = newTpl("switchTag", switchTagVDOMCode)
)

// exprApproxName tries to return a meaningful name for a control structure variable
func exprApproxName(expr string) string {
	var buf bytes.Buffer
	for _, c := range expr {
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '.' || c == ' ' {
			buf.WriteRune(c)
		}
	}

	s := buf.String()
	sf := strings.Fields(s)
	if len(sf) > 0 {
		s = sf[len(sf)-1]
	}

	split := strings.Split(s, ".")
	return strings.ToUpper(split[len(split)-1])
}

func fmtSTagError(specialTag, msg string) error {
	return errors.New(sfmt("special '%v' tag: %v", specialTag, msg))
}

func invalidAttribute(specialTag, attr string) error {
	return fmtSTagError(specialTag, sfmt("invalid attribute '%v'", attr))
}

func attrRequireNotEmpty(specialTag string, attr whtml.Attribute) error {
	if attr.Val == "" {
		return fmtSTagError(specialTag,
			sfmt("attribute '%v' cannot be empty", attr.Key))
	}

	return nil
}

func (z *htmlCompiler) forTagGenerate(
	w io.Writer, n *whtml.Node,
	da *declArea, refs refsMap,
) error {

	// process the attributes
	var keyName, valName string
	var rangeAttr whtml.Attribute
	for _, attr := range n.Attrs {
		switch attr.Key {
		case "k":
			keyName = attr.Val
		case "v":
			valName = attr.Val
		case "range":
			rangeAttr = attr
		default:
			return invalidAttribute(forSTag, attr.Key)
		}
	}

	if err := attrRequireNotEmpty(forSTag, rangeAttr); err != nil {
		return err
	}

	// declare a variable to hold this loops's list of nodes inside a parent Declaration Area
	varName := sfmt("for%v", exprApproxName(rangeAttr.Val))
	varName, cbuf := da.declare(varName)

	w.Write([]byte(varName))

	// create a Declaration Area so that
	// control structures (e.g an if tag) nested inside this one
	// could declare variables
	newDA := newDeclArea(da)
	children, err := z.childrenGenerate(n, newDA, refs)
	if err != nil {
		return err
	}

	return forTagVDOMTpl.Execute(cbuf, forTagVDOMTD{
		KeyName:  keyName,
		ValName:  valName,
		VarName:  varName,
		Items:    attributeValueCode(rangeAttr),
		Children: children,
		Decls:    newDA.code(),
	})
}

func (z *htmlCompiler) ifTagGenerate(
	w io.Writer, n *whtml.Node,
	da *declArea, refs refsMap,
) error {

	var condAttr whtml.Attribute
	for _, attr := range n.Attrs {
		switch attr.Key {
		case "cond":
			condAttr = attr
		default:
			return invalidAttribute(ifSTag, attr.Key)
		}
	}

	if err := attrRequireNotEmpty(ifSTag, condAttr); err != nil {
		return err
	}

	varName := sfmt("if%v", exprApproxName(condAttr.Val))
	varName, cbuf := da.declare(varName)
	w.Write([]byte(varName))

	newDA := newDeclArea(da)
	children, err := z.childrenGenerate(n, newDA, refs)
	if err != nil {
		return err
	}

	return ifTagVDOMTpl.Execute(cbuf, ifTagVDOMTD{
		VarName:  varName,
		Cond:     attributeValueCode(condAttr),
		Children: children,
		Decls:    newDA.code(),
	})
}

func invalidChildTag(parentTag, childTag string) error {
	return fmtSTagError(parentTag, sfmt("invalid child tag '%v'", childTag))
}

func (z *htmlCompiler) newCaseTagTD(n *whtml.Node, parentDA *declArea, refs refsMap, expr string) (
	*caseTagVDOMTD, error) {

	newDA := newDeclArea(parentDA)
	children, err := z.childrenGenerate(n, newDA, refs)
	if err != nil {
		return nil, err
	}

	return &caseTagVDOMTD{
		Children: children,
		Expr:     expr,
		Decls:    newDA.code(),
	}, nil
}

func (z *htmlCompiler) caseTagGenerate(n *whtml.Node, da *declArea, refs refsMap) (
	*caseTagVDOMTD, error) {

	var exprAttr whtml.Attribute
	for _, attr := range n.Attrs {
		switch attr.Key {
		case "expr":
			exprAttr = attr
		default:
			return nil, invalidAttribute(caseSTag, attr.Key)
		}
	}

	if err := attrRequireNotEmpty(caseSTag, exprAttr); err != nil {
		return nil, err
	}

	return z.newCaseTagTD(n, da, refs, attributeValueCode(exprAttr))
}

func (z *htmlCompiler) switchGetCases(n *whtml.Node, da *declArea, refs refsMap) (
	cases []*caseTagVDOMTD, deflt *caseTagVDOMTD, err error) {

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == whtml.TextNode {
			continue
		}

		switch c.Data {
		case caseSTag:
			var cs *caseTagVDOMTD
			cs, err = z.caseTagGenerate(c, da, refs)
			if err != nil {
				return
			}

			cases = append(cases, cs)
		case defaultSTag:
			if deflt != nil {
				err = fmtSTagError(switchSTag,
					sfmt("multiple '%v' child tags are not allowed.", defaultSTag))
				return
			}

			deflt, err = z.newCaseTagTD(c, da, refs, "")
			if err != nil {
				return
			}
		default:
			err = invalidChildTag(switchSTag, caseSTag)
			return
		}
	}

	return
}

func (z *htmlCompiler) switchTagGenerate(
	w io.Writer, n *whtml.Node,
	da *declArea, refs refsMap,
) error {

	var exprAttr whtml.Attribute
	for _, attr := range n.Attrs {
		switch attr.Key {
		case "expr":
			exprAttr = attr
		default:
			return invalidAttribute(switchSTag, attr.Key)
		}
	}

	var exprCode string
	if exprAttr.Val != "" {
		exprCode = attributeValueCode(exprAttr)
	}

	varName := sfmt("switch%v", exprApproxName(exprAttr.Val))
	varName, cbuf := da.declare(varName)
	w.Write([]byte(varName))

	cases, deflt, err := z.switchGetCases(n, da, refs)
	if err != nil {
		return err
	}

	return switchTagVDOMTpl.Execute(cbuf, switchTagVDOMTD{
		VarName: varName,
		Expr:    exprCode,
		Cases:   cases,
		Default: deflt,
	})
}
