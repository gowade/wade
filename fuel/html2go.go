package main

import (
	"bytes"
	"io"
	"strings"
	//"fmt"

	"github.com/gowade/html"
)

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

func (z *HTMLCompiler) elementGenerate(w io.Writer, el *html.Node) error {
	key, htmlAttrs := extractKeyFromAttrs(el.Attr)

	var children []*bytes.Buffer
	for c := el.FirstChild; c != nil; c = c.NextSibling {
		//clean pesky linebreaks and tabs in the HTML code
		if c.Type == html.TextNode && []rune(c.Data)[0] == '\n' &&
			justPeskySpaces(c.Data) &&
			strings.ToLower(el.Data) != "pre" {
			continue
		}

		var buf bytes.Buffer
		err := z.nodeGenerate(&buf, c)
		if err != nil {
			return err
		}

		children = append(children, &buf)
	}

	return elementVDOMTpl.Execute(w, elementVDOMTD{
		Tag:      el.Data,
		Key:      strAttributeValueCode(parseTextMustache(key)),
		Attrs:    toTplAttrs(htmlAttrs),
		Children: children,
		LastIdx:  len(children) - 1,
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

func (z *HTMLCompiler) textNodeGenerate(w io.Writer, node *html.Node) error {
	return textNodeVDOMTpl.Execute(w, textNodeVDOMTD{
		Text: escapeNewlines(node.Data),
	})
}

func (z *HTMLCompiler) nodeGenerate(w io.Writer, node *html.Node) error {
	switch node.Type {
	case html.ElementNode:
		return z.elementGenerate(w, node)
	case html.TextNode:
		return z.textNodeGenerate(w, node)
	}

	return nil
}

func (c *HTMLCompiler) GenerateFile(w io.Writer, node *html.Node) error {
	var buf bytes.Buffer
	err := c.elementGenerate(&buf, node)
	if err != nil {
		return err
	}

	renderFuncTpl.Execute(w, renderFuncTD{
		Return: &buf,
	})

	return nil
}
