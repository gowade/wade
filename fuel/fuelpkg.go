package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gowade/html"
	"github.com/gowade/wade/utils/htmlutils"
)

type comDef struct {
	name   string
	markup *html.Node
}

type htmlFile struct {
	path    string
	imports map[string]*fuelPkg
	comDefs []comDef //component definitions (top-level capitalized HTML elements)
}

type fuelPkg struct {
	*ast.Package
	fset *token.FileSet
	dir  string

	htmlFiles []*htmlFile
	imports   []string

	coms comMap
}

type comMap map[string]*ast.StructType

// getFuelPkg builds a tree containing info about a package and its dependencies
func getFuelPkg(dir string) (*fuelPkg, error) {
	fset := token.NewFileSet()
	pkg, err := parsePkg(dir, fset)
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

	htmlComs, err := comDefsMap(htmlFiles)
	if err != nil {
		return nil, err
	}

	coms := pkgComponents(pkg, htmlComs)

	return &fuelPkg{
		fset:      fset,
		Package:   pkg,
		htmlFiles: htmlFiles,
		imports:   getPkgDeps(pkg, htmlFiles),
		coms:      coms,
	}, nil
}

// get dependencies imported from inside the package's source code
func getPkgDeps(pkg *ast.Package, htmlFiles []*htmlFile) []string {
	var deps []string

	for _, file := range pkg.Files {
		for _, imp := range file.Imports {
			pdir := importDir(importPath(imp))
			if pdir != "" {
				deps = append(deps, pdir)
			}
		}
	}

	return deps
}

func addPkgFromImport(path string, pkgs *[]*fuelPkg) error {
	pdir := importDir(path)
	if pdir != "" {
		cpkg, err := getFuelPkg(pdir)
		if err != nil {
			return err
		}

		*pkgs = append(*pkgs, cpkg)
	}

	return nil
}

func parsePkg(dir string, fset *token.FileSet) (*ast.Package, error) {
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return !isFuelFile(fi.Name())
	}, 0)

	if err != nil {
		return nil, err
	}

	for _, pkg := range pkgs {
		if !strings.HasSuffix(pkg.Name, "_test") {
			return pkg, nil
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
func parseHTMLFile(filePath string) (imports map[string]*fuelPkg, comDefs []comDef, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}

	nodes, err := htmlutils.ParseFragment(file)
	if err != nil {
		return nil, nil, err
	}

	for _, node := range nodes {
		if node.Type == html.ElementNode {
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
				comDefs = append(comDefs, comDef{
					name:   node.Data,
					markup: node,
				})
			}
		}
	}

	return imports, comDefs, nil
}

// process an import tag and the package it imports
func htmlImportTag(node *html.Node) (name string, pkg *fuelPkg, err error) {
	var path string
	for _, attr := range node.Attr {
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

	if pdir := importDir(path); pdir != "" {
		pkg, err = getFuelPkg(pdir)
		if err != nil {
			return name, pkg, err
		}
	}

	return name, pkg, nil
}

func comDefsMap(htmlFiles []*htmlFile) (map[string]string, error) {
	comDefs := make(map[string]string)
	for _, hf := range htmlFiles {
		for _, com := range hf.comDefs {
			if definedFile := comDefs[com.name]; definedFile != "" {
				return nil, efmt("%v:%v: duplicated component definition, "+
					"first defined here %v", hf.path, com.name, definedFile)
			}
			comDefs[com.name] = hf.path
		}
	}

	return comDefs, nil
}

// return a name -> *ast.StructType map of the package's components
func pkgComponents(pkg *ast.Package, comDefs map[string]string) comMap {
	coms := make(map[string]*ast.StructType)
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
							if comDefs[name] != "" {
								coms[name] = stype
							}
						}
					}
				}
			}
		}
	}

	return coms
}
