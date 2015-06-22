package main

import (
	"fmt"

	"github.com/gowade/html"
)

func lnode(code string) *codeNode {
	return &codeNode{
		typ:  SliceVarCodeNode,
		code: code,
	}
}

func (c *HTMLCompiler) forLoopCode(node *html.Node, vda *varDeclArea) (*codeNode, error) {
	keyName, valName := "_", "_"
	var rangeAttr html.Attribute
	for _, attr := range node.Attr {
		switch attr.Key {
		case "k":
			keyName = attr.Val
		case "v":
			valName = attr.Val
		case "range":
			rangeAttr = attr

		default:
			return nil, fmt.Errorf(`Invalid attribute "%v" for for loop.`, attr.Key)
		}
	}

	if rangeAttr.Val == "" {
		return nil, fmt.Errorf(`for loop's "range" attribute cannot be empty.`)
	}

	if !rangeAttr.IsMustache {
		return nil, fmt.Errorf(
			`for loop's "range" attribute must be assigned to a `+
				`mustache representing a Go slice value. Got string value "%v" instead.`,
			rangeAttr.Val)

	}

	varName := vda.newVar("for")
	forVda := newVarDeclArea()
	apList := []*codeNode{{
		typ:  SliceVarCodeNode,
		code: varName,
	}}
	apList = append(apList, c.genChildren(node, forVda, nil)...)

	forVda.saveToCN()

	vda.setVarDecl(
		varName,
		ncn(fmt.Sprintf(`%v := %v{}`, varName, NodeListOpener)),
		&codeNode{
			typ:  BlockCodeNode,
			code: fmt.Sprintf(`for __k, __v := range %v`, rangeAttr.Val),
			children: []*codeNode{
				ncn(fmt.Sprintf(`%v, %v := __k, __v`, keyName, valName)),
				forVda.codeNode,
				&codeNode{
					typ:      AppendListCodeNode,
					code:     fmt.Sprintf("%v = ", varName),
					children: apList,
				},
			},
		})

	return lnode(fmt.Sprintf(varName)), nil
}

func (c *HTMLCompiler) ifControlCode(node *html.Node, vda *varDeclArea) (*codeNode, error) {
	var rcond html.Attribute
	for _, attr := range node.Attr {
		switch attr.Key {
		case "cond":
			rcond = attr

		default:
			return nil, fmt.Errorf(`Invalid attribute "%v" for if.`, attr.Key)
		}
	}

	cond := rcond.Val
	if cond == "" {
		return nil, fmt.Errorf(`if structure's "cond" attribute cannot be empty`)
	}

	if !rcond.IsMustache {
		return nil, fmt.Errorf(
			`if structure's "cond" attribute must be assigned to a `+
				`mustache respresenting a Go boolean expression. Got string value "%v" instead.`, cond)
	}

	varName := vda.newVar("if")
	ifVda := newVarDeclArea()

	child := c.generateRec(node.FirstChild, ifVda, nil)[0]
	ifVda.saveToCN()

	child.code = fmt.Sprintf("%v = ", varName) + child.code
	vda.setVarDecl(
		varName,
		ncn(fmt.Sprintf(`var %v vdom.Node`, varName)),
		&codeNode{
			typ:  BlockCodeNode,
			code: fmt.Sprintf(`if %v `, cond),
			children: []*codeNode{
				ifVda.codeNode,
				child,
			},
		})

	return ncn(varName), nil
}
