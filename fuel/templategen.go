package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/gowade/html"
	"unicode"
)

func isComponentArgName(attrName string) bool {
	if attrName == "" {
		return false
	}

	return unicode.IsUpper([]rune(attrName)[0])
}

func cnToBuffer(cn *codeNode) *bytes.Buffer {
	var buf bytes.Buffer
	emitDomCode(&buf, cn)
	return &buf
}

func (c HTMLCompiler) componentInstCode(com componentInfo, uNode *html.Node, key string, vda *varDeclArea, instChildren *codeNode, comChild bool) (*codeNode, error) {
	fields := make([]fieldAssTD, 0)

	attrs := make([]html.Attribute, 0, len(uNode.Attr))
	for _, attr := range uNode.Attr {
		if !isComponentArgName(attr.Key) {
			attrs = append(attrs, attr)
		}
	}

	var ac *codeNode
	if comChild {
		ac = ncn(combinedAttrs)
	} else {
		ac = elementAttrsCode(attrs)
	}

	for _, attr := range uNode.Attr {
		if isComponentArgName(attr.Key) {
			vcode := attributeValueCode(attr)
			fields = append(fields, fieldAssTD{
				Name:  attr.Key,
				Value: vcode,
			})
		}
	}

	comType := com.fullName()
	var buf bytes.Buffer
	err := comInitFuncTpl.ExecuteTemplate(&buf, "comInit", &comInitFuncTD{
		ComType: comType,
		Com: &comCreateTD{
			ComName:  com.name,
			ComType:  comType,
			Children: cnToBuffer(instChildren),
			Attrs:    cnToBuffer(ac),
		},
		Fields: fields,
	})
	if err != nil {
		panic(err)
	}

	return &codeNode{
		typ:  FuncCallCodeNode,
		code: CreateComElementOpener,
		children: []*codeNode{
			&codeNode{typ: StringCodeNode, code: com.name},
			ncn(key),
			ncn("&" + comType + "{}"),
			ncn(buf.String()),
		},
	}, nil
}

func textNodeCode(text string) []*codeNode {
	parts := parseTextMustache(text)
	ret := make([]*codeNode, len(parts))

	for i, part := range parts {
		var cn *codeNode
		if part.isMustache {
			cn = ncn(valueToStringCode(part.content))
		} else {
			cn = &codeNode{
				typ:  StringCodeNode,
				code: part.content,
			}
		}

		ret[i] = &codeNode{
			typ:      FuncCallCodeNode,
			code:     CreateTextNodeOpener,
			children: []*codeNode{cn},
		}
	}

	return ret
}

func strAttributeValueCode(parts []textPart) string {
	if len(parts) == 1 && !parts[0].isMustache {
		return "`" + parts[0].content + "`"
	}

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

func attributeValueCode(attr html.Attribute) string {
	if attr.IsMustache {
		return attr.Val
	}
	parts := parseTextMustache(attr.Val)
	return strAttributeValueCode(parts)
}

func elementAttrsCode(attrs []html.Attribute) *codeNode {
	if len(attrs) == 0 {
		return ncn("nil")
	}

	assignments := make([]*codeNode, 0, len(attrs))
	for _, attr := range attrs {
		if attr.Key == "ref" || attr.Key == "key" {
			continue
		}

		assignments = append(assignments, &codeNode{
			typ:  NakedCodeNode,
			code: mapFieldAssignmentCode(attr.Key, attributeValueCode(attr)),
		})
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
