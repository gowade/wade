package main

import (
	"bytes"
	"io"
	"strings"
	//"fmt"

	"github.com/gowade/html"
)

func newHTMLCompiler(htmlFile string, w io.Writer, root *html.Node, pkg *fuelPkg) *htmlCompiler {
	return &htmlCompiler{
		fileName: htmlFile,
		w:        w,
		root:     root,
		pkg:      pkg,
	}
}

func compileHTMLFile(fileName string, w io.Writer, root *html.Node) error {
	compiler := newHTMLCompiler(fileName, w, root, nil)
	return compiler.Generate()
}

type htmlCompiler struct {
	fileName string
	w        io.Writer
	root     *html.Node
	pkg      *fuelPkg
}

const (
	keyAttrName = "key"
)

func extractKeyFromAttrs(attrs []html.Attribute) (key string, retAttrs []html.Attribute) {
	retAttrs = make([]html.Attribute, 0, len(attrs))
	for _, attr := range attrs {
		if attr.Key == keyAttrName {
			key = attr.Val
		} else {
			retAttrs = append(retAttrs, attr)
		}
	}

	return key, attrs
}

func toTplAttrs(attrs []html.Attribute) map[string]string {
	m := make(map[string]string)
	for _, attr := range attrs {
		m[attr.Key] = attributeValueCode(attr)
	}

	return m
}

func (z *htmlCompiler) ComponentGenerate() error {
	return nil
}

func (z *htmlCompiler) childrenGenerate(parent *html.Node, da *declArea) ([]*bytes.Buffer, error) {
	var children []*bytes.Buffer
	for c := parent.FirstChild; c != nil; c = c.NextSibling {
		// clean pesky linebreaks and tabs in the HTML code
		if c.Type == html.TextNode && []rune(c.Data)[0] == '\n' &&
			justPeskySpaces(c.Data) &&
			strings.ToLower(parent.Data) != "pre" {
			continue
		}

		// generate
		var buf bytes.Buffer
		err := z.nodeGenerate(&buf, c, da)
		if err != nil {
			return nil, err
		}

		children = append(children, &buf)
	}

	return children, nil
}

func (z *htmlCompiler) elementGenerate(w io.Writer, el *html.Node, da *declArea) error {
	key, htmlAttrs := extractKeyFromAttrs(el.Attr)
	children, err := z.childrenGenerate(el, da)
	if err != nil {
		return err
	}

	return elementVDOMTpl.Execute(w, elementVDOMTD{
		Tag:      el.Data,
		Key:      strAttributeValueCode(parseTextMustache(key)),
		Attrs:    toTplAttrs(htmlAttrs),
		Children: children,
	})
}

func escapeNewlines(str string) string {
	var buf bytes.Buffer
	for _, c := range []rune(str) {
		if c == '\n' {
			buf.WriteString(`\n`)
		} else {
			buf.WriteRune(c)
		}
	}

	return buf.String()
}

func (z *htmlCompiler) textNodeGenerate(w io.Writer, node *html.Node) error {
	parts := parseTextMustache(node.Data)

	return textNodeVDOMTpl.Execute(w, textNodeVDOMTD{
		Text: strAttributeValueCode(parts),
	})
}

func (z *htmlCompiler) nodeGenerate(w io.Writer, node *html.Node, da *declArea) error {
	switch node.Type {
	case html.ElementNode:
		if fn := z.specialTag(node.Data); fn != nil {
			return fn(w, node, da)
		}

		return z.elementGenerate(w, node, da)
	case html.TextNode:
		return z.textNodeGenerate(w, node)
	}

	return nil
}

func (z *htmlCompiler) Generate() error {
	err := z.generate()
	if err != nil {
		return efmt("%v: %v", z.fileName, err.Error())
	}

	return nil
}

func (z *htmlCompiler) generate() error {
	var buf bytes.Buffer
	da := newDeclArea(nil)
	err := z.elementGenerate(&buf, z.root, da)
	if err != nil {
		return err
	}

	renderFuncTpl.Execute(z.w, renderFuncTD{
		Return: &buf,
		Decls:  da.code(),
	})

	return nil
}
