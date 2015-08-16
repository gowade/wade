package main

import (
	"bytes"
	"io"
	"strings"
	//"fmt"

	"github.com/gowade/html"
)

const (
	RefAttrName = "ref"
)

type refsMap map[string]string
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
	comDef comDef,
	pkg *fuelPkg,
	comSpec comSpecMap) *htmlCompiler {
	return &htmlCompiler{
		htmlFile: htmlFile,
		w:        w,
		root:     comDef.markup,
		comName:  comDef.name,
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
	comName  string
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

func refNameFromAttr(attrName string) string {
	var buf bytes.Buffer
	for i, c := range attrName {
		if ((c >= '0' && c <= '9') && i > 0) ||
			(c >= 'A' && c <= 'Z') ||
			(c >= 'a' && c <= 'z') {
			buf.WriteRune(c)
		}

		if c == '-' || c == ' ' {
			buf.WriteRune('_')
		}
	}

	return buf.String()
}

func (z *htmlCompiler) childrenGenerate(parent *html.Node, da *declArea, refs refsMap) (
	[]childCode, error) {

	var children []childCode
	for c := parent.FirstChild; c != nil; c = c.NextSibling {
		// clean pesky linebreaks and tabs in the HTML code
		if c.Type == html.TextNode && []rune(c.Data)[0] == '\n' &&
			justPeskySpaces(c.Data) &&
			strings.ToLower(parent.Data) != "pre" {
			continue
		}

		// generate
		var buf bytes.Buffer
		err := z.nodeGenerate(&buf, c, da, refs)
		if err != nil {
			return nil, err
		}

		// ref
		var refName string
		if refs != nil && c.Type == html.ElementNode {
			for _, attr := range c.Attr {
				if attr.Key == RefAttrName {
					refName = refNameFromAttr(attr.Val)
					refs[refName] = c.Data
				}
			}
		}

		children = append(children, childCode{
			Code:    &buf,
			RefName: refName,
			ElTag:   c.Data,
		})
	}

	return children, nil
}

func (z *htmlCompiler) elementGenerate(
	w io.Writer, el *html.Node,
	da *declArea, refs refsMap) error {

	key, htmlAttrs := extractKeyFromAttrs(el.Attr)

	children, err := z.childrenGenerate(el, da, refs)
	if err != nil {
		return err
	}

	return must(elementVDOMTpl.Execute(w, elementVDOMTD{
		Tag:      el.Data,
		Key:      strAttributeValueCode(parseTextMustache(key)),
		Attrs:    toTplAttrs(htmlAttrs),
		Children: children,
	}))
}

func (z *htmlCompiler) textNodeGenerate(w io.Writer, node *html.Node) error {
	parts := parseTextMustache(node.Data)

	return must(textNodeVDOMTpl.Execute(w, textNodeVDOMTD{
		Text: strAttributeValueCode(parts),
	}))
}

func (z *htmlCompiler) comInstGenerate(
	w io.Writer,
	node *html.Node,
	da *declArea, refs refsMap,
	info *comInfo) error {

	fieldsAss := make([]fieldAssTD, 0, len(node.Attr))
	for _, attr := range node.Attr {
		if isCapitalized(attr.Key) {
			fieldsAss = append(fieldsAss, fieldAssTD{
				Name:  attr.Key,
				Value: attributeValueCode(attr),
			})
		}
	}

	var childrenCode []childCode
	childrenCode, err := z.childrenGenerate(node, da, nil)
	if err != nil {
		return err
	}

	comType := info.name
	if info.importSelector != "" {
		comType = sfmt("%v.%v", info.importSelector, info.name)
	}

	return must(comCreateTpl.Execute(w, comCreateTD{
		ComName:      info.name,
		ComType:      comType,
		Decls:        da.code(),
		ChildrenCode: childrenCode,
		FieldsAss:    fieldsAss,
	}))
}

type comInfo struct {
	name, importSelector string
}

func (z *htmlCompiler) elComponent(node *html.Node) (
	*comInfo, error) {

	pkg := z.pkg
	var info comInfo

	csplit := strings.Split(node.Data, ":")
	if len(csplit) == 2 {
		// imported component
		info.name = csplit[1]
		if isCapitalized(info.name) {
			info.importSelector = csplit[0]
			impPkg, ok := z.htmlFile.imports[info.importSelector]
			if !ok {
				return nil, efmt("cannot create component instance for %v"+
					", %v has not been imported", info.name, info.importSelector)
			}

			pkg = impPkg.fuelPkg
		}
	} else {
		// local component
		if isCapitalized(node.Data) {
			info.name = node.Data
		}
	}

	if info.name != "" && pkg.coms != nil {
		if _, ok := pkg.coms[info.name]; ok {
			return &info, nil
		} else {
			return nil, efmt("unknown component %v", info.name)
		}
	}

	// element is not considered a component
	return nil, nil
}
func (z *htmlCompiler) nodeGenerate(w io.Writer, node *html.Node,
	da *declArea, refs refsMap) error {

	switch node.Type {
	case html.ElementNode:
		if fn := z.specialTag(node.Data); fn != nil {
			return fn(w, node, da, refs)
		}

		if z.pkg != nil {
			comInfo, err := z.elComponent(node)
			if err != nil {
				return err
			}

			if comInfo != nil {
				return z.comInstGenerate(w, node, da, refs, comInfo)
			}
		}

		return z.elementGenerate(w, node, da, refs)
	case html.TextNode:
		return z.textNodeGenerate(w, node)
	}

	return nil
}

func (z *htmlCompiler) newError(err error) error {
	return efmt("%v: %v", z.htmlFile.path, err.Error())
}

func (z *htmlCompiler) Generate() error {
	err := z.generate(z.root, nil)
	if err != nil {
		return z.newError(err)
	}

	return nil
}

func (z *htmlCompiler) generate(root *html.Node, refs refsMap) error {
	var buf bytes.Buffer
	var decls bytes.Buffer
	if root != nil {
		da := newDeclArea(nil)
		err := z.elementGenerate(&buf, root, da, refs)
		if err != nil {
			return err
		}

		decls = *da.code()
	} else {
		buf.WriteString("nil")
	}

	return must(renderFuncTpl.Execute(z.w, renderFuncTD{
		ComName: z.comName,
		Return:  &buf,
		Decls:   &decls,
		HasRefs: len(refs) > 0,
	}))
}

func (z *htmlCompiler) ComponentGenerate() error {
	err := z.componentGenerate()
	if err != nil {
		return z.newError(err)
	}

	return nil
}

func (z *htmlCompiler) componentGenerate() error {
	refs := make(refsMap)
	err := z.generate(z.root, refs)
	if err != nil {
		return err
	}

	if len(refs) > 0 {
		return must(refsTpl.Execute(z.w, refsTD{
			ComName: z.comName,
			Refs:    refs,
		}))
	}

	return nil
}
