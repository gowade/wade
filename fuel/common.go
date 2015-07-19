package main

import "fmt"

type varDeclArea struct {
	list      []string
	vars      map[string][]*codeNode
	prefixIdx map[string]int
	codeNode  *codeNode
	parent    *varDeclArea
}

func newVarDeclArea(parent *varDeclArea) *varDeclArea {
	return &varDeclArea{
		vars:      map[string][]*codeNode{},
		prefixIdx: map[string]int{},
		codeNode: &codeNode{
			typ:  VarDeclAreaCodeNode,
			code: "",
		},
		parent: parent,
	}
}

func (vda *varDeclArea) newVar(prefix string) string {
	vda.prefixIdx[prefix]++
	if vda.parent != nil {
		vda.prefixIdx[prefix] += vda.parent.prefixIdx[prefix]
	}

	return fmt.Sprintf("%v%v", prefix, vda.prefixIdx[prefix])
}

func (vda *varDeclArea) setVarDecl(varName string, nlist ...*codeNode) {
	if _, ok := vda.vars[varName]; !ok {
		vda.list = append(vda.list, varName)
	}
	vda.vars[varName] = nlist
}

func (vda *varDeclArea) saveToCN() {
	vda.codeNode.children = make([]*codeNode, 0)
	for _, varName := range vda.list {
		if cn, ok := vda.vars[varName]; ok {
			for _, d := range cn {
				vda.codeNode.children = append(vda.codeNode.children, d)
			}
		}
	}
}

func ncn(code string) *codeNode {
	return &codeNode{
		typ:  NakedCodeNode,
		code: code,
	}
}
