package main

import (
	"fmt"

	"golang.org/x/net/html"
)

type varDeclArea struct {
	vars      map[string][]*codeNode
	prefixIdx map[string]int
	codeNode  *codeNode
}

func newVarDeclArea() *varDeclArea {
	return &varDeclArea{
		vars:      map[string][]*codeNode{},
		prefixIdx: map[string]int{},
		codeNode: &codeNode{
			typ:  VarDeclAreaCodeNode,
			code: "",
		},
	}
}

func (vda *varDeclArea) newVar(prefix string) string {
	vda.prefixIdx[prefix]++
	varName := fmt.Sprintf("%v%v", prefix, vda.prefixIdx[prefix])
	if _, exists := vda.vars[varName]; exists {
		panic(fmt.Sprintf("var %v has already been declared.", varName))
	}

	return varName
}

func (vda *varDeclArea) setVarDecl(varName string, nlist ...*codeNode) {
	vda.vars[varName] = nlist
}

func (vda *varDeclArea) saveToCN() {
	vda.codeNode.children = make([]*codeNode, 0)
	for _, cn := range vda.vars {
		for _, d := range cn {
			vda.codeNode.children = append(vda.codeNode.children, d)
		}
	}
}

func ncn(code string) *codeNode {
	return &codeNode{
		typ:  NakedCodeNode,
		code: code,
	}
}

func lnode(code string) *codeNode {
	return &codeNode{
		typ:  SliceVarCodeNode,
		code: code,
	}
}

func extractSingleMustache(attrVal string) string {
	parts := parseTextMustache(attrVal)

	if len(parts) != 1 || !parts[0].isMustache {
		return ""
	}

	return parts[0].content
}

func forLoopCode(node *html.Node, vda *varDeclArea) (*codeNode, error) {
	keyName, valName := "_", "_"
	rrv := ""
	for _, attr := range node.Attr {
		switch attr.Key {
		case "k":
			keyName = attr.Val
		case "v":
			valName = attr.Val
		case "range":
			rrv = attr.Val

		default:
			return nil, fmt.Errorf(`Invalid attribute "%v" for for loop.`, attr.Key)
		}
	}

	rangeVar := extractSingleMustache(rrv)
	if rangeVar == "" {
		return nil, fmt.Errorf(
			`For loop's "range" attribute must be assigned to a `+
				`mustache representing a Go slice value. Got "%v" instead.`, rrv)

	}

	varName := vda.newVar("for")
	forVda := newVarDeclArea()
	apList := []*codeNode{{
		typ:  SliceVarCodeNode,
		code: varName,
	}}
	apList = append(apList, genChildren(node, forVda)...)

	forVda.saveToCN()

	vda.setVarDecl(
		varName,
		ncn(fmt.Sprintf(`%v := %v{}`, varName, ElementListOpener)),
		&codeNode{
			typ:  BlockCodeNode,
			code: fmt.Sprintf(`for __k, __v := range %v`, rangeVar),
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

func ifControlCode(node *html.Node, vda *varDeclArea) (*codeNode, error) {
	rcond := ""
	for _, attr := range node.Attr {
		switch attr.Key {
		case "cond":
			rcond = attr.Val

		default:
			return nil, fmt.Errorf(`Invalid attribute "%v" for if.`, attr.Key)
		}
	}

	cond := extractSingleMustache(rcond)
	if cond == "" {
		return nil, fmt.Errorf(
			`If's "cond" attribute must be assigned to a `+
				`mustache respresenting a Go boolean expression. Got "%v" instead.`, rcond)

	}

	varName := vda.newVar("if")
	ifVda := newVarDeclArea()
	children := genChildren(node, ifVda)

	ifVda.saveToCN()

	vda.setVarDecl(
		varName,
		ncn(fmt.Sprintf(`%v := %v{}`, varName, ElementListOpener)),
		&codeNode{
			typ:  BlockCodeNode,
			code: fmt.Sprintf(`if %v `, cond),
			children: []*codeNode{
				ifVda.codeNode,
				&codeNode{
					typ:      ElemListCodeNode,
					code:     fmt.Sprintf("%v = ", varName),
					children: children,
				},
			},
		})

	return lnode(fmt.Sprintf(varName)), nil
}
