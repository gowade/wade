package main

import (
	"fmt"

	"github.com/gowade/html"
	"github.com/gowade/wade/utils/htmlutils"
)

func NewHTMLCompiler(coms componentMap) *HTMLCompiler {
	return &HTMLCompiler{
		coms: coms,
	}
}

type HTMLCompiler struct {
	errors []error
	coms   componentMap
}

func (c *HTMLCompiler) Errors() []error {
	return c.errors
}

func (c *HTMLCompiler) Error() error {
	if c.errors == nil || len(c.errors) == 0 {
		return nil
	}

	return c.errors[0]
}

func (c *HTMLCompiler) elementCode(node *html.Node, vda *varDeclArea) *codeNode {
	children := c.genChildren(node, vda)
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

func (c *HTMLCompiler) genChildren(node *html.Node, vda *varDeclArea) []*codeNode {
	children := make([]*codeNode, 0)
	i := 0
	for ch := node.FirstChild; ch != nil; ch = ch.NextSibling {
		chAppend(&children, c.generateRec(ch, vda))

		i++
	}

	return filterTextStrings(children)
}

func (c *HTMLCompiler) addError(err error) {
	c.errors = append(c.errors, err)
}

func (c *HTMLCompiler) getComponent(tagName string) (i componentInfo, ok bool) {
	if c.coms == nil {
		ok = false
		return
	}

	i, ok = c.coms[tagName]
	return
}

func (c *HTMLCompiler) generateRec(node *html.Node, vda *varDeclArea) []*codeNode {
	if node.Type == html.TextNode {
		return textNodeCode(node.Data)
	}

	if node.Type == html.ElementNode {
		var cn *codeNode
		var err error
		switch node.Data {
		case "for":
			htmlutils.RemoveGarbageTextChildren(node)
			cn, err = c.forLoopCode(node, vda)
		case "if":
			htmlutils.RemoveGarbageTextChildren(node)
			cn, err = c.ifControlCode(node, vda)

		default:
			if com, ok := c.getComponent(node.Data); ok {
				cn, err = componentInstCode(com, node, &codeNode{
					typ:      CompositeCodeNode,
					code:     NodeListOpener,
					children: c.genChildren(node, vda),
				})
			} else {
				cn = c.elementCode(node, vda)
			}
		}

		if err != nil {
			c.addError(err)
		}

		return []*codeNode{cn}
	}

	return nil
}

func (c *HTMLCompiler) renderFuncOpener(tagName string, com *componentInfo) string {
	embedStr := ""
	if com != nil {
		tname := com.name
		if com.stateField != "" {
			tname = "*" + tname
		}
		embedStr = fmt.Sprintf(RenderEmbedString, tname)
	}

	return fmt.Sprintf(RenderFuncOpener, embedStr)
}

func (c *HTMLCompiler) Generate(node *html.Node, com *componentInfo) *codeNode {
	renderNode := node
	children := make([]*codeNode, 0)
	if com != nil {
		htmlutils.RemoveGarbageTextChildren(node)
		if node.FirstChild == nil || node.LastChild != node.FirstChild {
			c.addError(fmt.Errorf(
				`Invalid HTML markup definition for %v, please make sure it contains `+
					`exactly 1 child.`, node.Data))
			return nil
		}

		renderNode = node.FirstChild

		if com.stateField != "" {
			children = append(children, ncn(
				fmt.Sprintf(ComponentSetStateCode, com.stateField, com.stateType)))
		}
	}

	c.errors = make([]error, 0)
	vda := newVarDeclArea()

	cnode := c.generateRec(renderNode, vda)[0]
	cnode.code = "return " + cnode.code
	ret := &codeNode{
		typ:  BlockCodeNode,
		code: c.renderFuncOpener(node.Data, com),
		children: append(children,
			vda.codeNode,
			cnode),
	}

	vda.saveToCN()
	return ret
}
