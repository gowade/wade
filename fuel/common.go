package main

import "fmt"

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

func extractSingleMustache(attrVal string) string {
	parts := parseTextMustache(attrVal)

	if len(parts) != 1 || !parts[0].isMustache {
		return ""
	}

	return parts[0].content
}
