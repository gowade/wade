package main

import (
	"fmt"
	"strings"

	"github.com/gowade/html"
)

var (
	nilCode = &codeNode{
		typ:  NakedCodeNode,
		code: "nil",
	}
)

func (c HTMLCompiler) componentInstCode(com componentInfo, uNode *html.Node, key string, vda *varDeclArea, instChildren *codeNode) (*codeNode, error) {
	varName := vda.newVar("com")

	fields := make([]*codeNode, 0, len(com.argFields)+1)
	instChildren.code = fmt.Sprintf("Children: %v", instChildren.code)

	comCh := []*codeNode{
		ncn(fmt.Sprintf(`Name: "%v"`, com.name)),
		ncn(fmt.Sprintf(`VNode: %v`, varName)),
		ncn(fmt.Sprintf(`InternalRefsHolder: %v{}`, com.name+"Refs")),
		instChildren,
	}

	fields = append(fields, &codeNode{
		typ:      CompositeCodeNode,
		code:     "Com: " + ComponentDataOpener,
		children: comCh,
	})

	for _, attr := range uNode.Attr {
		if com.argFields[attr.Key] {
			vcode := attributeValueCode(attr)
			fields = append(fields, &codeNode{
				typ:  NakedCodeNode,
				code: fmt.Sprintf("%v: %v", attr.Key, vcode),
			})

			continue
		}

		return nil, fmt.Errorf(`Invalid field "%v" for component %v`, attr.Key, com.name)
	}

	typeIns := "&" + com.name
	if com.state.field != "" {
		if com.state.isPointer {
			fields = append(fields, ncn(
				fmt.Sprintf(`%v: &%v{}`, com.state.field, com.state.typ)))
		}
	}

	cn := &codeNode{
		typ:      CompositeCodeNode,
		code:     typeIns,
		children: fields,
	}

	cn.code = varName + fmt.Sprintf(` := %v("%v", %v, nil)`,
		CreateComElementOpener,
		com.name,
		key) +
		fmt.Sprintf("\n%v.Component = ", varName) + cn.code
	vda.setVarDecl(varName, cn)

	return ncn(varName), nil
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
		return nilCode
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
