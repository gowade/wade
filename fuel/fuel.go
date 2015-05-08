package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	//"unicode"

	"github.com/gowade/html"
	"github.com/gowade/wade/utils/htmlutils"
)

const (
	fuelSuffix = ".fuel.go"
)

type htmlInfo struct {
	file   string
	markup *html.Node
}

type stateInfo struct {
	field     string
	typ       string
	structTyp *ast.StructType
}

type componentInfo struct {
	htmlInfo  htmlInfo
	name      string
	argFields map[string]bool
	state     stateInfo
}

type componentMap map[string]componentInfo

type Fuel struct {
	dir        string
	components componentMap
}

func NewFuel(dir string) *Fuel {
	return &Fuel{
		dir:        dir,
		components: componentMap{},
	}
}

func (f *Fuel) BuildPackage() {
	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, f.dir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), fuelSuffix)
	}, 0)

	checkFatal(err)

	htmlComs, comList := f.getHtmlComponents()
	var pkgName string

	for _, pkg := range pkgs {
		if pkgName == "" && !strings.HasSuffix(pkg.Name, "_test") {
			pkgName = pkg.Name
		}

		ast.PackageExports(pkg)
		for _, file := range pkg.Files {
			f.getComponents(file, htmlComs)
		}
	}

	htmlCompiler := NewHTMLCompiler(f.components)
	for _, comName := range comList {
		f.buildComponent(htmlCompiler, f.components[comName], pkgName)
	}

	mfile, err := os.Create("autogen.fuel.go")
	if err != nil {
		fatal(err.Error())
	}

	write(mfile, Prelude(pkgName))
	for _, com := range f.components {
		if com.state.field != "" {
			write(mfile, stateMethsCode(com, fset))
		}

		write(mfile, comRefsDeclCode(com.name, htmlCompiler.comRefs[com.name]))
		write(mfile, comRefsMethsCode(com.name))
		write(mfile, fmt.Sprintf(`func (this %v) Rerender() {
	r := this.Render(nil)
	wade.DOM().PerformDiff(r, this.VNode.Element, this.VNode.DOMNode())
	this.VNode.Element = r
}

`, com.name))
	}
}

func comRefsMethsCode(comName string) string {
	return fmt.Sprintf(`func (this %v) Refs() *%vRefs {
	return this.Com.InternalRefsHolder.(*%vRefs)	
}

`, comName, comName, comName)
}

func comRefsDeclCode(comName string, refs []comRef) string {
	fields := make([]string, 0, len(refs))
	if refs != nil {
		fields = append(fields, "")
		for _, ref := range refs {
			elTp, _ := domElType(ref.elTag)
			fields = append(fields, fmt.Sprintf("\t%v %v", ref.name, elTp))
		}
		fields = append(fields, "")
	}
	return fmt.Sprintf(`type %vRefs struct {%v}
`, comName, strings.Join(fields, "\n"))
}

func stateMethsCode(com componentInfo, fset *token.FileSet) string {
	setters := ""
	for _, f := range com.state.structTyp.Fields.List {
		pos := fset.Position(f.Type.Pos())
		end := fset.Position(f.Type.End())
		file, err := os.Open(pos.Filename)
		if err != nil {
			fatal(err.Error())
		}

		buf := make([]byte, end.Offset-pos.Offset)
		_, err = file.ReadAt(buf, int64(pos.Offset))
		if err != nil {
			fatal(err.Error())
		}

		fname := fieldName(f)
		setters += fmt.Sprintf(`func (this %v) Set%v(v %v) {
	this.%v.%v = v
	this.Rerender()
}

`, com.name, fname, string(buf), com.state.field, fname)
	}

	return fmt.Sprintf(`func (this %v) InternalState() interface{} {
	return this.%v
}

`,
		//com.name, com.stateField, com.stateType,
		com.name, com.state.field) + setters
}

func (f *Fuel) buildComponent(compiler *HTMLCompiler, com componentInfo, pkgName string) {
	fileName := com.name + fuelSuffix
	err := writeGoDomFile(compiler, com.htmlInfo.markup, fileName, pkgName, &com)
	if err != nil {
		fatal("Error building component %v, HTML file %v:\n`%v`", com.name, com.htmlInfo.file, err.Error())
	}

	runGofmt(fileName)
}

func (f *Fuel) getHtmlComponents() (map[string]htmlInfo, []string) {
	files, err := ioutil.ReadDir(f.dir)
	if err != nil {
		checkFatal(err)
	}

	m := make(map[string]htmlInfo)
	comList := make([]string, 0)
	for _, fileInfo := range files {
		if !strings.HasSuffix(fileInfo.Name(), ".html") {
			continue
		}

		file, err := os.Open(fileInfo.Name())
		checkFatal(err)
		nodes, err := htmlutils.ParseFragment(file)
		checkFatal(err)

		for _, node := range nodes {
			if node.Type == html.ElementNode {
				if _, exists := m[node.Data]; exists {
					fatal(`Fatal Error: Found multiple definitions in HTML for component "%v".`, node.Data)
				}

				comList = append(comList, node.Data)
				m[node.Data] = htmlInfo{
					markup: node,
					file:   fileInfo.Name(),
				}
			}
		}
	}

	return m, comList
}

func anonFieldName(typ ast.Expr) string {
	switch t := typ.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return anonFieldName(t.X)
	case *ast.SelectorExpr:
		return anonFieldName(t.X)
	}

	panic(fmt.Sprintf("Unhandled ast expression type %T", typ))
	return ""
}

func fieldName(f *ast.Field) string {
	if len(f.Names) > 0 {
		return f.Names[0].Name
	}

	return anonFieldName(f.Type)
}

func extractFields(comName string, fields []*ast.Field) (map[string]bool, stateInfo) {
	argFields := make(map[string]bool)
	var state stateInfo
	for _, f := range fields {
		fname := fieldName(f)
		if f.Tag != nil {
			var stag = reflect.StructTag(f.Tag.Value[1 : len(f.Tag.Value)-1])
			if sf := stag.Get("fuel"); sf == "state" {
				if state.field != "" {
					fatal("Error processing component %v: component can only have 1 state field.", comName)
				}

				if pt, ok := f.Type.(*ast.StarExpr); ok {
					if ftype, ok := pt.X.(*ast.Ident); ok {
						state.field = fname
						state.typ = "*" + ftype.Name
						if spec, ok := ftype.Obj.Decl.(*ast.TypeSpec); ok {
							if st, ok := spec.Type.(*ast.StructType); ok {
								state.structTyp = st
							}
						}
						continue
					}
				}
				fatal(`Error processing field "%v" of component %v: state field's type must be pointer,
pointing to a named type (anonymous struct is forbidden).`, fname, comName)
			}
		}

		argFields[fname] = true
	}

	return argFields, state
}

func (f *Fuel) getComponents(file *ast.File, htmlComs map[string]htmlInfo) {
	for _, decl := range file.Decls {
		switch gdecl := decl.(type) {
		case *ast.GenDecl:
			if gdecl.Tok == token.TYPE {
				for _, ospec := range gdecl.Specs {
					spec := ospec.(*ast.TypeSpec)
					name := spec.Name.Name
					switch stype := spec.Type.(type) {
					case *ast.StructType:
						if hcom, ok := htmlComs[name]; ok {
							argFields, state := extractFields(name, stype.Fields.List)
							f.components[name] = componentInfo{
								name:      name,
								argFields: argFields,
								state:     state,
								htmlInfo:  hcom,
							}
						}
					}
				}
			}
		}
	}
}
