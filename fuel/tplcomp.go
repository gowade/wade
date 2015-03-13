package main

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

var (
	nilCode = &codeNode{
		typ:  nakedCodeNode,
		code: "nil",
	}
)

func textNodeCode(text string) []*codeNode {
	parts := parseTextMustache(text)
	ret := make([]*codeNode, len(parts))

	for i, part := range parts {
		cnType := stringCodeNode
		if part.isMustache {
			cnType = nakedCodeNode
		}

		ret[i] = &codeNode{
			typ:  funcCallCodeNode,
			code: createTextNodeOpener,
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
			typ:  nakedCodeNode,
			code: mapFieldAssignmentCode(attr.Key, valueCode),
		}
	}

	return &codeNode{
		typ:      compositeCodeNode,
		code:     attributeMapOpener,
		children: assignments,
	}
}

func elementCode(node *html.Node) *codeNode {
	children := make([]*codeNode, 0)
	foreachChildren(node, func(_ int, c *html.Node) {
		children = append(children, generateRec(c)...)
	})

	childrenCode := nilCode
	if len(children) != 0 {
		childrenCode = &codeNode{
			typ:      compositeCodeNode,
			code:     elementListOpener,
			children: children,
		}
	}

	return &codeNode{
		typ:  funcCallCodeNode,
		code: createElementOpener,
		children: []*codeNode{
			&codeNode{typ: stringCodeNode, code: node.Data}, // element tag name
			elementAttrsCode(node.Attr),
			childrenCode,
		},
	}
}

func generateRec(node *html.Node) []*codeNode {
	if node.Type == html.TextNode {
		return textNodeCode(node.Data)
	}

	if node.Type == html.ElementNode {
		return []*codeNode{elementCode(node)}
	}

	return nil
}

func generate(node *html.Node) *codeNode {
	return generateRec(node)[0]
}
