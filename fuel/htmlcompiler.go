package main

import (
	"fmt"
	"strings"

	"github.com/gowade/html"
	"github.com/gowade/wade/utils/htmlutils"
)

func NewHTMLCompiler(coms componentMap) *HTMLCompiler {
	return &HTMLCompiler{
		coms:    coms,
		comRefs: make(map[string][]comRef),
	}
}

type HTMLCompiler struct {
	errors  []error
	coms    componentMap
	comRefs map[string][]comRef
}

type comRef struct {
	name    string
	varName string
	elTag   string
}

type comRefs struct {
	refs []comRef
	vda  *varDeclArea
}

func newComRefs(vda *varDeclArea) *comRefs {
	return &comRefs{make([]comRef, 0), vda}
}

func (r *comRefs) add(refName string, elTag string, code *codeNode) string {
	vname := r.vda.newVar("ref")
	code.code = vname + " := " + code.code
	r.vda.setVarDecl(vname, code)
	r.refs = append(r.refs, comRef{refName, vname, elTag})
	return vname
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

func (c *HTMLCompiler) elementCode(node *html.Node, key string, vda *varDeclArea, comRefs *comRefs) *codeNode {
	children := c.genChildren(node, vda, comRefs)
	childrenCode := nilCode
	if len(children) != 0 {
		childrenCode = &codeNode{
			typ:      ElemListCodeNode,
			code:     "",
			children: children,
		}
	}

	cn := &codeNode{
		typ:  FuncCallCodeNode,
		code: CreateElementOpener,
		children: []*codeNode{
			&codeNode{typ: StringCodeNode, code: node.Data}, // element tag name
			ncn(key),
			elementAttrsCode(node.Attr),
			childrenCode,
		},
	}

	return cn
}

func (c *HTMLCompiler) genChildren(node *html.Node, vda *varDeclArea, comRefs *comRefs) []*codeNode {
	children := make([]*codeNode, 0)
	i := 0
	for ch := node.FirstChild; ch != nil; ch = ch.NextSibling {
		chAppend(&children, c.generateRec(ch, vda, comRefs))

		i++
	}

	return children
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

func (c *HTMLCompiler) generateRec(node *html.Node, vda *varDeclArea, comRefs *comRefs) []*codeNode {
	if node.Type == html.TextNode {
		return textNodeCode(node.Data)
	}

	if node.Type == html.ElementNode {
		var cn *codeNode
		var err error
		switch node.Data {
		case "render":
			cn, err = c.renderTagCode(node, vda)

		case "for":
			htmlutils.RemoveGarbageTextChildren(node)
			cn, err = c.forLoopCode(node, vda)
		case "if":
			htmlutils.RemoveGarbageTextChildren(node)
			cn, err = c.ifControlCode(node, vda)
		case "switch":
			htmlutils.RemoveGarbageTextChildren(node)
			cn, err = c.switchControlCode(node, vda)

		default:
			key := `""`
			for _, attr := range node.Attr {
				if attr.Key == "key" {
					key = valueToStringCode(attributeValueCode(attr))
				}
			}

			parts := strings.Split(node.Data, ":")
			comName := strings.Join(parts, ".")

			if com, ok := c.getComponent(comName); ok {
				children := c.genChildren(node, vda, nil)
				cn, err = c.componentInstCode(com, node, key, vda, &codeNode{
					typ:      CompositeCodeNode,
					code:     NodeListOpener,
					children: children,
				})
			} else {
				cn = c.elementCode(node, key, vda, comRefs)
			}

			for _, attr := range node.Attr {
				if attr.Key == "ref" {
					cn = ncn(comRefs.add(attr.Val, node.Data, cn))
				}
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
		if com.state.field != "" {
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

		if com.state.field != "" {
			children = append(children,
				ncn(componentSetStateCode(com.state.field, com.state.typ, com.state.isPointer)))
		}
	}

	c.errors = make([]error, 0)
	vda := newVarDeclArea()

	var cnode *codeNode
	if com != nil {
		refs := newComRefs(vda)
		cnode = c.generateRec(renderNode, vda, refs)[0]
		refsVar, refsSet := componentRefsVarCode(com.name)
		if len(refs.refs) > 0 {
			c.comRefs[com.name] = refs.refs
			children = append(children, ncn(refsVar))
		}

		vda.saveToCN()
		children = append(children, vda.codeNode)

		for _, ref := range refs.refs {
			children = append(children, ncn(componentSetRefCode(ref.name, ref.varName, ref.elTag)))
		}

		if len(refs.refs) > 0 {
			children = append(children, ncn(refsSet))
		}
	} else {
		cnode = c.generateRec(renderNode, vda, nil)[0]
		vda.saveToCN()
		children = append(children, vda.codeNode)
	}

	cnode.code = "return " + cnode.code
	ret := &codeNode{
		typ:      BlockCodeNode,
		code:     c.renderFuncOpener(node.Data, com),
		children: append(children, cnode),
	}

	return ret
}
