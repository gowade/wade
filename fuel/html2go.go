package main

import (
	"bytes"
	"io"
	"strings"
	//"fmt"

	"github.com/gowade/html"
)

type comSpec struct {
	childrenField string
}

type comSpecMap map[string]comSpec

func newHTMLCompiler(htmlFileName string, w io.Writer, root *html.Node) *htmlCompiler {
	return &htmlCompiler{
		htmlFile: &htmlFile{
			path: htmlFileName,
		},
		w:    w,
		root: root,
	}
}

func newComponentHTMLCompiler(
	htmlFile *htmlFile,
	w io.Writer,
	root *html.Node,
	pkg *fuelPkg,
	comSpec comSpecMap) *htmlCompiler {
	return &htmlCompiler{
		htmlFile: htmlFile,
		w:        w,
		root:     root,
		pkg:      pkg,
		comSpec:  comSpec,
	}
}

func compileHTMLFile(fileName string, w io.Writer, root *html.Node) error {
	compiler := newHTMLCompiler(fileName, w, root)
	return compiler.Generate()
}

type htmlCompiler struct {
	htmlFile *htmlFile
	w        io.Writer
	root     *html.Node
	pkg      *fuelPkg
	comSpec  comSpecMap
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

func (z *htmlCompiler) comInstGenerate(
	w io.Writer,
	node *html.Node,
	da *declArea,
	impSel string, comName string) error {

	fieldsAss := make([]fieldAssTD, 0, len(node.Attr))
	for _, attr := range node.Attr {
		if isCapitalized(attr.Key) {
			fieldsAss = append(fieldsAss, fieldAssTD{
				Name:  attr.Key,
				Value: attributeValueCode(attr),
			})
		}
	}

	//var childrenField string
	//if cs, ok := z.comSpec[comName]; ok {
	//childrenField = cs.childrenField
	//}

	var childrenCode []*bytes.Buffer
	//if childrenField != "" {
	childrenCode, err := z.childrenGenerate(node, da)
	if err != nil {
		return err
	}
	//}

	comType := comName
	if impSel != "" {
		comType = sfmt("%v.%v", impSel, comType)
	}

	return comCreateTpl.Execute(w, comCreateTD{
		ComName:       comName,
		ComType:       comType,
		ChildrenField: "Children",
		Decls:         da.code(),
		ChildrenCode:  childrenCode,
		FieldsAss:     fieldsAss,
	})
}

func (z *htmlCompiler) nodeGenerate(w io.Writer, node *html.Node, da *declArea) error {
	switch node.Type {
	case html.ElementNode:
		if fn := z.specialTag(node.Data); fn != nil {
			return fn(w, node, da)
		}

		if z.pkg != nil {
			if z.pkg.coms != nil {
				// imported component
				csplit := strings.Split(node.Data, ":")
				if len(csplit) == 2 {
					comName := csplit[1]
					if isCapitalized(comName) {
						impSel := csplit[0]
						impPkg := z.htmlFile.imports[impSel]
						if impPkg == nil {
							return efmt("cannot create component instance for %v"+
								", %v has not been imported", node.Data, impSel)
						}

						if impPkg.coms[comName] != nil {
							return z.comInstGenerate(w, node, da, impSel, comName)
						}
					}
				}

				// component
				if isCapitalized(node.Data) {
					if z.pkg.coms[node.Data] != nil {
						return z.comInstGenerate(w, node, da, "", node.Data)
					} else {
						return efmt("unknown component %v", node.Data)
					}
				}
			}
		}

		return z.elementGenerate(w, node, da)
	case html.TextNode:
		return z.textNodeGenerate(w, node)
	}

	return nil
}

func (z *htmlCompiler) newError(err error) error {
	return efmt("%v: %v", z.htmlFile.path, err.Error())
}

func (z *htmlCompiler) Generate() error {
	err := z.generate(z.root)
	if err != nil {
		return z.newError(err)
	}

	return nil
}

func (z *htmlCompiler) generate(root *html.Node) error {
	var buf bytes.Buffer
	da := newDeclArea(nil)
	err := z.elementGenerate(&buf, root, da)
	if err != nil {
		return err
	}

	renderFuncTpl.Execute(z.w, renderFuncTD{
		ComName: z.root.Data,
		Return:  &buf,
		Decls:   da.code(),
	})

	return nil
}

func (z *htmlCompiler) ComponentGenerate() error {
	err := z.componentGenerate()
	if err != nil {
		return z.newError(err)
	}

	return nil
}

func (z *htmlCompiler) componentGenerate() error {
	cleanGarbageTextChildren(z.root)
	if z.root != nil && z.root.FirstChild != z.root.LastChild {
		return efmt("%v: component definition cannot have more than 1 direct children",
			z.root.Data)
	}

	return z.generate(z.root.FirstChild)
}
