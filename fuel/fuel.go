package main

import (
	"go/ast"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	genPrefix  = "~"
	fuelSuffix = ".fuel.go"
	importSTag = "import"
)

func generatedFileName(name string) string {
	return genPrefix + name + fuelSuffix
}

var (
	methodsFile = generatedFileName("spice")
)

var (
	gSrcPath string
)

func srcPath() string {
	if gSrcPath == "" {
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			fatal("GOPATH environment variable has not been set, please set it to a correct value.")
		}

		gSrcPath = filepath.Join(gopath, "src")
	}

	return gSrcPath
}

func importPath(imp *ast.ImportSpec) string {
	return imp.Path.Value[1 : len(imp.Path.Value)-1]
}

func importDir(impPath string) string {
	pdir := filepath.Join(srcPath(), filepath.FromSlash(impPath))
	if _, err := os.Stat(pdir); err == nil {
		return pdir
	}

	return ""
}

func importName(imp *ast.ImportSpec) string {
	if imp.Name == nil {
		return path.Base(importPath(imp))
	}

	return imp.Name.String()
}

func componentVDOMFilePath(file *htmlFile) string {
	odir, obase := filepath.Dir(file.path), filepath.Base(file.path)
	return filepath.Join(odir, generatedFileName(obase))
}

func fieldNameFromPath(fieldPath string) string {
	tc := strings.Title(fieldPath)
	frags := strings.Split(tc, ".")
	return strings.Join(frags, "")
}

func toTplStateFields(fields []*fieldInfo) []stateFieldTD {
	//detect name clashes
	m := make(map[string]int)
	for _, field := range fields {
		m[field.name]++
	}

	ret := make([]stateFieldTD, 0, len(fields))
	for _, field := range fields {
		fieldName := field.name
		if m[field.name] > 1 {
			fieldName = fieldNameFromPath(field.path)
		}

		ret = append(ret, stateFieldTD{
			Name: fieldName,
			Path: field.path,
			Type: field.typeName,
		})
	}

	return ret
}

func toImportList(imports map[string]string) []importTD {
	ret := make([]importTD, 0, len(imports))
	for name, path := range imports {
		ret = append(ret, importTD{
			Name: name,
			Path: path,
		})
	}

	return ret
}

func defaultImports(m map[string]string) map[string]string {
	m["fmt"] = "fmt"
	m["vdom"] = "github.com/gowade/vdom"
	m["wade"] = "github.com/gowade/wade"
	m["dom"] = "github.com/gowade/wade/dom"

	return m
}

func htmlFileVDOMGenerate(pkg *fuelPkg, file *htmlFile) error {
	// create the file
	filePath := componentVDOMFilePath(file)
	ofile, err := os.Create(filePath)
	if err != nil {
		checkFatal(err)
	}

	apkg := newAstPkg(pkg.pkg, pkg.imports)
	defaultImports(apkg.genImports)

	// process the components's struct, get their state fields
	// additional imports required for those fields are put into apkg.genImports
	comSfMap := make(map[string][]*fieldInfo)
	for _, com := range file.comDefs {
		if cs, ok := pkg.comStructs[com.name]; ok {
			stateFields, err := apkg.getStateFields("", cs.stype.Fields.List, cs.file)
			if err != nil {
				return efmt("Error when processing %v struct: %v", com.name, err)
			}

			comSfMap[com.name] = stateFields
		}
	}

	// add the imports from state fields to prelude
	imports := make([]importTD, 0, len(file.imports))
	for impName, impPath := range apkg.genImports {
		imports = append(imports, importTD{
			Name: impName,
			Path: impPath,
		})
	}

	// add imports from HTML file to prelude
	for name, impPkg := range file.imports {
		imports = append(imports, importTD{
			Name: name,
			Path: impPkg.importPath,
		})
	}

	// generate prelude
	preludeTpl.Execute(ofile, preludeTD{
		Pkg:     pkg.pkg.Name,
		Imports: imports,
	})

	for _, com := range file.comDefs {
		// generate render method
		compiler := newComponentHTMLCompiler(file, ofile, com, pkg, nil)
		err = compiler.componentGenerate()
		if err != nil {
			return err
		}

		// generate other methods
		stateFields := toTplStateFields(comSfMap[com.name])
		comMethodsTpl.Execute(ofile, comMethodsTD{
			Receiver:    "*" + com.name,
			StateFields: stateFields,
		})
	}

	runGofmt(filePath)
	return nil
}

func fuelBuild(dir string) error {
	pkg, err := getFuelPkg(dir)
	checkFatal(err)

	return fuelBuildRec(pkg)
}

func fuelBuildRec(pkg *fuelPkg) error {
	for _, file := range pkg.htmlFiles {
		err := htmlFileVDOMGenerate(pkg, file)
		if err != nil {
			return err
		}

		for _, imp := range file.imports {
			if imp.fuelPkg != nil {
				err := fuelBuildRec(imp.fuelPkg)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
