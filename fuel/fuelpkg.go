package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gowade/whtml"
)

type comDef struct {
	name   string
	markup *whtml.Node
}

type htmlFile struct {
	path    string
	imports map[string]importedPkg
	comDefs map[string]comDef //component definitions (top-level capitalized HTML elements)
}

type importedPkg struct {
	*fuelPkg
	importPath string
}

type parsedPkg struct {
	*ast.Package
	fset *token.FileSet
}

type pkgMap map[string]*parsedPkg

type comMap map[string]*htmlFile

type comStructInfo struct {
	stype *ast.StructType
	file  *ast.File
}

type comStructMap map[string]comStructInfo

type fuelPkg struct {
	pkg *parsedPkg
	dir string

	htmlFiles []*htmlFile
	imports   pkgMap

	comStructs comStructMap
	coms       comMap
}

// getFuelPkg builds a tree containing info about a package and its dependencies
func getFuelPkg(dir string) (*fuelPkg, error) {
	pkg, err := parsePkg(dir)
	if err != nil {
		return nil, err
	}

	if pkg == nil {
		return nil, nil
	}

	htmlFiles, err := pkgHTMLFiles(dir)
	if err != nil {
		return nil, err
	}

	coms, err := htmlComs(htmlFiles)
	if err != nil {
		return nil, err
	}

	comStructs := pkgComponents(pkg.Package, coms)

	imports := make(pkgMap)
	pkgDeps(imports, pkg.Package)

	return &fuelPkg{
		pkg:        pkg,
		htmlFiles:  htmlFiles,
		imports:    imports,
		coms:       coms,
		comStructs: comStructs,
	}, nil
}

// get dependencies imported from inside the package's source code
func pkgDeps(imports pkgMap, pkg *ast.Package) {
	for _, file := range pkg.Files {
		for _, imp := range file.Imports {
			importPath := importPath(imp)
			if _, ok := imports[importPath]; ok {
				return
			}

			pdir := importDir(importPath)
			if pdir != "" {
				pkg, err := parsePkg(pdir)

				imports[importPath] = pkg
				if err == nil {
					pkgDeps(imports, pkg.Package)
				}
			} else {
				imports[importPath] = nil
			}
		}
	}
}

func parsePkg(dir string) (*parsedPkg, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return !isFuelFile(fi.Name())
	}, 0)

	if err != nil {
		return nil, err
	}

	for _, pkg := range pkgs {
		if !strings.HasSuffix(pkg.Name, "_test") {
			return &parsedPkg{pkg, fset}, nil
		}
	}

	return nil, nil
}

func pkgHTMLFiles(dir string) ([]*htmlFile, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var htmlFiles []*htmlFile
	for _, fi := range files {
		if strings.HasSuffix(fi.Name(), ".html") {
			filePath := filepath.Join(dir, fi.Name())
			imports, comDefs, err := parseHTMLFile(filePath)
			if err != nil {
				return nil, err
			}

			htmlFiles = append(htmlFiles, &htmlFile{
				path:    filePath,
				imports: imports,
				comDefs: comDefs,
			})
		}
	}

	return htmlFiles, nil
}

// parse a component HTML markup file, returning its imports and component definitions (capitalized top-level elements)
func parseHTMLFile(filePath string) (
	imports map[string]importedPkg,
	comDefs map[string]comDef,
	err error) {

	imports = make(map[string]importedPkg)
	comDefs = make(map[string]comDef)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}

	nodes, err := whtml.Parse(file)
	if err != nil {
		return nil, nil, err
	}

	for _, node := range nodes {
		if node.Type == whtml.ElementNode {
			// import tag
			if node.Data == importSTag {
				impName, pkg, err := htmlImportTag(node)
				if err != nil {
					return nil, nil, err
				}

				imports[impName] = pkg
			}

			// component definition element
			if isCapitalized(node.Data) {
				cleanGarbageTextChildren(node)

				comDefs[node.Data] = comDef{
					name:   node.Data,
					markup: node.FirstChild,
				}
			}
		}
	}

	return imports, comDefs, nil
}

// process an import tag and the package it imports
func htmlImportTag(node *whtml.Node) (name string, pkg importedPkg, err error) {
	var path string
	for _, attr := range node.Attrs {
		switch attr.Key {
		case "from":
			if err = attrRequireNotEmpty(importSTag, attr); err != nil {
				return name, pkg, err
			}

			path = attr.Val

		case "as":
			name = attr.Val

		default:
			return name, pkg, invalidAttribute(importSTag, attr.Key)
		}
	}

	pkg.importPath = path
	if pdir := importDir(path); pdir != "" {
		pkg.fuelPkg, err = getFuelPkg(pdir)
		if err != nil {
			return
		}
	}

	return
}

func htmlComs(htmlFiles []*htmlFile) (comMap, error) {
	comDefs := make(comMap)
	for _, hf := range htmlFiles {
		for _, com := range hf.comDefs {
			if definedFile := comDefs[com.name]; definedFile != nil {
				return nil, efmt("%v:%v: duplicated component definition, "+
					"first defined here %v", hf.path, com.name, definedFile)
			}
			comDefs[com.name] = hf
		}
	}

	return comDefs, nil
}

// return a name -> *ast.StructType map of the package's components
func pkgComponents(pkg *ast.Package, comDefs comMap) comStructMap {
	coms := make(comStructMap)
	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			switch gdecl := decl.(type) {
			case *ast.GenDecl:
				if gdecl.Tok == token.TYPE {
					for _, ospec := range gdecl.Specs {
						spec := ospec.(*ast.TypeSpec)
						name := spec.Name.Name
						switch stype := spec.Type.(type) {
						case *ast.StructType:
							if comDefs[name] != nil {
								coms[name] = comStructInfo{
									stype: stype,
									file:  file,
								}
							}
						}
					}
				}
			}
		}
	}

	return coms
}
