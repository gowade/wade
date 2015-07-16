package main

import (
	"fmt"
	"strings"

	"github.com/gowade/html"
	"github.com/gowade/wade/utils/htmlutils"
)

const (
	combinedAttrs = "combinedAttrs"
)

func NewHTMLCompiler(coms componentMap) *HTMLCompiler {
	return &HTMLCompiler{
		coms:    coms,
		comRefs: make(map[string][]comRef),
	}
}

type HTMLCompiler struct {
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

func (c *HTMLCompiler) elementCode(node *html.Node, key string, vda *varDeclArea, comRefs *comRefs) (*codeNode, error) {
	return c.elementCodeCC(node, key, vda, comRefs, false)
}

func (c *HTMLCompiler) elementCodeCC(node *html.Node, key string, vda *varDeclArea, comRefs *comRefs, comChild bool) (*codeNode, error) {
	children, err := c.genChildren(node, vda, comRefs)
	if err != nil {
		return nil, err
	}

	childrenCode := ncn("nil")
	if len(children) != 0 {
		childrenCode = &codeNode{
			typ:      ElemListCodeNode,
			code:     "",
			children: children,
		}
	}

	var attrsCode *codeNode
	if comChild {
		attrsCode = ncn(combinedAttrs)
	} else {
		attrsCode = elementAttrsCode(node.Attr)
	}

	cn := &codeNode{
		typ:  FuncCallCodeNode,
		code: CreateElementOpener,
		children: []*codeNode{
			&codeNode{typ: StringCodeNode, code: node.Data}, // element tag name
			ncn(key),
			attrsCode,
			childrenCode,
		},
	}

	return cn, nil
}

func (c *HTMLCompiler) genChildren(node *html.Node, vda *varDeclArea, comRefs *comRefs) (
	[]*codeNode, error) {
	children := make([]*codeNode, 0)
	i := 0
	for ch := node.FirstChild; ch != nil; ch = ch.NextSibling {
		// clean pesky linebreaks and tabs in the HTML code
		if ch.Type == html.TextNode && []rune(ch.Data)[0] == '\n' &&
			strings.TrimSpace(ch.Data) == "" &&
			strings.ToLower(node.Data) != "pre" {
			continue
		}

		l, err := c.generateRec(ch, vda, comRefs)
		if err != nil {
			return nil, err
		}

		chAppend(&children, l)

		i++
	}

	return children, nil
}

func (c *HTMLCompiler) getComponent(tagName string) (i componentInfo, ok bool) {
	if c.coms == nil {
		ok = false
		return
	}

	i, ok = c.coms[tagName]
	return
}

func (c *HTMLCompiler) generateRec(node *html.Node, vda *varDeclArea, comRefs *comRefs) (
	[]*codeNode, error) {
	return c.generateRecCC(node, vda, comRefs, false)
}

func (c *HTMLCompiler) generateRecCC(node *html.Node, vda *varDeclArea, comRefs *comRefs, comChild bool) (
	[]*codeNode, error) {
	if node.Type == html.TextNode {
		return textNodeCode(node.Data), nil
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
				var children []*codeNode
				children, err = c.genChildren(node, vda, nil)
				if err != nil {
					return nil, err
				}

				cn, err = c.componentInstCode(com, node, key, vda, &codeNode{
					typ:      CompositeCodeNode,
					code:     NodeListOpener,
					children: children,
				}, comChild)

			} else {
				cn, err = c.elementCodeCC(node, key, vda, comRefs, comChild)
			}

			if err != nil {
				return nil, err
			}

			for _, attr := range node.Attr {
				if attr.Key == "ref" {
					cn = ncn(comRefs.add(attr.Val, node.Data, cn))
				}
			}
		}

		if err != nil {
			return nil, err
		}

		return []*codeNode{cn}, nil
	}

	return nil, nil
}

func (c *HTMLCompiler) renderFuncOpener(tagName string, com *componentInfo) string {
	embedStr := ""
	if com != nil {
		tname := "*" + com.name
		embedStr = fmt.Sprintf(RenderEmbedString, tname)
	}

	return fmt.Sprintf(RenderFuncOpener, embedStr)
}

func (c *HTMLCompiler) Generate(node *html.Node, com *componentInfo) (*codeNode, error) {
	renderNode := node
	initCode := ncn("vdom.InternalRenderLock(); this.OnInvoke(); vdom.InternalRenderUnlock()")
	children := []*codeNode{initCode}
	if com != nil {
		htmlutils.RemoveGarbageTextChildren(node)

		if node.LastChild != node.FirstChild {
			return nil, fmt.Errorf(
				`Invalid HTML markup definition for %v, `+
					`it cannot have more than 1 direct child.`, node.Data)
		}

		if node.FirstChild == nil {
			return &codeNode{
				typ:  BlockCodeNode,
				code: c.renderFuncOpener(node.Data, com),
				children: []*codeNode{
					initCode,
					ncn("return nil"),
				},
			}, nil
		}

		renderNode = node.FirstChild

		if com.state != nil {
			children = append(children,
				ncn(componentSetStateCode()))
		}
	}

	vda := newVarDeclArea()

	var cnode *codeNode
	var l []*codeNode
	var err error

	if com != nil {
		refs := newComRefs(vda)
		l, err = c.generateRecCC(renderNode, vda, refs, true)
		if err != nil {
			return nil, err
		}
		cnode = l[0]

		refsVar, refsSet := componentRefsVarCode(com.name)
		if len(refs.refs) > 0 {
			c.comRefs[com.name] = refs.refs
			children = append(children, ncn(refsVar))
		}

		pac := &codeNode{
			typ:  FuncCallCodeNode,
			code: combinedAttrs + " := " + "wade.MergeMaps",
			children: []*codeNode{
				elementAttrsCode(renderNode.Attr),
				ncn("this.Com.Attrs"),
			},
		}

		vda.setVarDecl(combinedAttrs, pac)

		vda.saveToCN()
		children = append(children, vda.codeNode)

		for _, ref := range refs.refs {
			children = append(children, ncn(componentSetRefCode(ref.name, ref.varName, ref.elTag)))
		}

		if len(refs.refs) > 0 {
			children = append(children, ncn(refsSet))
		}
	} else {
		l, err = c.generateRec(renderNode, vda, nil)
		cnode = l[0]
		if err != nil {
			return nil, err
		}

		vda.saveToCN()
		children = append(children, vda.codeNode)
	}

	cnode.code = "ret := " + cnode.code
	ret := &codeNode{
		typ:      BlockCodeNode,
		code:     c.renderFuncOpener(node.Data, com),
		children: append(children, cnode, ncn("return ret")),
	}

	return ret, nil
}
