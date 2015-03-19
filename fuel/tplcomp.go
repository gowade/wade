package main

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

var (
	nilCode = &codeNode{
		typ:  NakedCodeNode,
		code: "nil",
	}
)

func textNodeCode(text string) []*codeNode {
	parts := parseTextMustache(text)
	ret := make([]*codeNode, len(parts))

	for i, part := range parts {
		cnType := StringCodeNode
		if part.isMustache {
			cnType = NakedCodeNode
		}

		ret[i] = &codeNode{
			typ:  FuncCallCodeNode,
			code: CreateTextNodeOpener,
			children: []*codeNode{
				{
					typ:  cnType,
					code: part.content,
				},
			},
		}
	}

	return ret
}

func attributeValueCode(parts []textPart) string {
	fmtStr := ""
	mustaches := []string{}
	for _, part := range parts {
		if part.isMustache {
			fmtStr += "%v"
			mustaches = append(mustaches, part.content)
		} else {
			fmtStr += part.content
		}
	}

	mStr := strings.Join(mustaches, ", ")
	return fmt.Sprintf("fmt.Sprintf(`%v`, %v)", fmtStr, mStr)
}

func mapFieldAssignmentCode(field string, value string) string {
	return fmt.Sprintf(`"%v": %v`, field, value)
}

func elementAttrsCode(attrs []html.Attribute) *codeNode {
	if len(attrs) == 0 {
		return nilCode
	}

	assignments := make([]*codeNode, len(attrs))
	for i, attr := range attrs {
		valueCode := attributeValueCode(parseTextMustache(attr.Val))
		assignments[i] = &codeNode{
			typ:  NakedCodeNode,
			code: mapFieldAssignmentCode(attr.Key, valueCode),
		}
	}

	return &codeNode{
		typ:      CompositeCodeNode,
		code:     AttributeMapOpener,
		children: assignments,
	}
}

func chAppend(a *[]*codeNode, b []*codeNode) {
	for _, item := range b {
		if item != nil {
			*a = append(*a, item)
		}
	}
}

func elementCode(node *html.Node, vda *varDeclArea) *codeNode {
	switch node.Data {
	case "for":
		cn, err := forLoopCode(node, vda)
		if err != nil {
			fmt.Println(err.Error())
		}

		return cn
	case "if":
		cn, err := ifControlCode(node, vda)
		if err != nil {
			fmt.Println(err.Error())
		}

		return cn
	}

	children := make([]*codeNode, 0)
	foreachChildren(node, func(_ int, c *html.Node) {
		chAppend(&children, generateRec(c, vda))
	})

	childrenCode := nilCode
	if len(children) != 0 {
		childrenCode = &codeNode{
			typ:      ElemListCodeNode,
			code:     "",
			children: children,
		}
	}

	return &codeNode{
		typ:  FuncCallCodeNode,
		code: CreateElementOpener,
		children: []*codeNode{
			&codeNode{typ: StringCodeNode, code: node.Data}, // element tag name
			elementAttrsCode(node.Attr),
			childrenCode,
		},
	}
}

func generateRec(node *html.Node, vda *varDeclArea) []*codeNode {
	if node.Type == html.TextNode {
		return textNodeCode(node.Data)
	}

	if node.Type == html.ElementNode {
		return []*codeNode{elementCode(node, vda)}
	}

	return nil
}

func generate(node *html.Node) *codeNode {
	vda := newVarDeclArea()
	ret := &codeNode{
		typ:  BlockCodeNode,
		code: RenderFuncOpener,
		children: []*codeNode{
			vda.codeNode,
			ncn("return "),
			generateRec(node, vda)[0],
		},
	}

	vda.saveToCN()
	return ret
}
