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
	"unicode"

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

type componentInfo struct {
	htmlInfo   htmlInfo
	name       string
	argFields  map[string]bool
	stateField string
	stateType  string
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
	for _, pkg := range pkgs {
		ast.PackageExports(pkg)
		for _, file := range pkg.Files {
			f.getComponents(file, htmlComs)
		}
	}

	htmlCompiler := NewHTMLCompiler(f.components)
	for _, comName := range comList {
		f.buildComponent(htmlCompiler, f.components[comName])
	}

	mfile, err := os.Create("methods.fuel.go")
	if err != nil {
		fatal(err.Error())
	}

	write(mfile, `package main
`)

	for _, com := range f.components {
		if com.stateField != "" {
			write(mfile, stateMethsCode(com))
		}

		//write(mfile, fmt.Sprintf(`func (this *%v) InternalComPtr() *wade.Com {
		//return &this.Com
		//}

		//`, com.name))
	}
}

func stateMethsCode(com componentInfo) string {
	/*
		func (this *%v) InternalSetState(stateData interface{}) {
		this.%v = stateData.(%v)
	} */
	return fmt.Sprintf(`
func (this %v) InternalState() interface{} {
	return this.%v
}
`,
		//com.name, com.stateField, com.stateType,
		com.name, com.stateField)
}

func (f *Fuel) buildComponent(compiler *HTMLCompiler, com componentInfo) {
	fileName := com.name + fuelSuffix
	err := writeGoDomFile(compiler, com.htmlInfo.markup, fileName, &com)
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

func extractFields(comName string, fields []*ast.Field) (map[string]bool, string, string) {
	argFields := make(map[string]bool)
	var stateField, stateType string
	for _, f := range fields {
		var fname string
		if len(f.Names) > 0 {
			fname = f.Names[0].Name
		} else {
			fname = anonFieldName(f.Type)
		}

		if f.Tag != nil {
			var stag = reflect.StructTag(f.Tag.Value[1 : len(f.Tag.Value)-1])
			if sf := stag.Get("fuel"); sf == "state" {
				if stateField != "" {
					fatal("Error processing component %v: component can only have 1 state field.", comName)
				}

				if pt, ok := f.Type.(*ast.StarExpr); ok {
					if ftype, ok := pt.X.(*ast.Ident); ok {
						stateField = fname
						stateType = "*" + ftype.Name
						continue
					}
				}
				fatal(`Error processing field "%v" of component %v: state field's type must be pointer,
pointing to a named type (anonymous struct is forbidden).`, fname, comName)
			}
		}

		if unicode.IsUpper([]rune(fname)[0]) {
			argFields[fname] = true
		}
	}

	return argFields, stateField, stateType
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
							argFields, stateField, stateType := extractFields(name, stype.Fields.List)
							f.components[name] = componentInfo{
								name:       name,
								argFields:  argFields,
								stateField: stateField,
								stateType:  stateType,
								htmlInfo:   hcom,
							}
						}
					}
				}
			}
		}
	}
}
