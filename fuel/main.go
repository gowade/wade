package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"strings"

	"golang.org/x/net/html"

	"github.com/gowade/wade/utils/htmlutils"
)

const (
	fuelSuffix = ".fuel.go"
)

func main() {
	flag.Parse()
	command := flag.Arg(0)
	switch command {
	case "build":
		bTarget := flag.Arg(1)
		if bTarget != "" {
			buildHtmlFile(bTarget)
		} else {
			buildPackage()
		}
	default:
		fatal("Please specify a command.")
	}
}

func fatal(msg string, fmtargs ...interface{}) {
	fmt.Fprintf(os.Stdout, msg+"\n", fmtargs...)
	os.Exit(2)
}

func checkFatal(err error) {
	if err != nil {
		fatal(err.Error())
	}
}

func getHtmlComponents(dir string) map[string]*html.Node {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		checkFatal(err)
	}

	m := make(map[string]*html.Node)
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

				m[node.Data] = node
			}
		}
	}

	return m
}

func buildComponent(compiler *HtmlCompiler, name string, htmlNode *html.Node, fields []*ast.Field) {
	writeGoDomFile(compiler, htmlNode, "c_"+name+fuelSuffix)
}

func buildComponents(dir string, file *ast.File) {
	components := getHtmlComponents(dir)
	compiler := NewHtmlCompiler()

	for _, decl := range file.Decls {
		switch gdecl := decl.(type) {
		case *ast.GenDecl:
			if gdecl.Tok == token.TYPE {
				for _, ospec := range gdecl.Specs {
					spec := ospec.(*ast.TypeSpec)
					name := spec.Name.Name
					switch stype := spec.Type.(type) {
					case *ast.StructType:
						if htmlnode, ok := components[strings.ToLower(name)]; ok {
							buildComponent(compiler, name, htmlnode, stype.Fields.List)
						}
					}
				}
			}
		}
	}
}

func buildPackage() {
	fset := token.NewFileSet()
	wd, err := os.Getwd()
	checkFatal(err)

	pkgs, err := parser.ParseDir(fset, wd, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), fuelSuffix)
	}, 0)

	checkFatal(err)

	for _, pkg := range pkgs {
		ast.PackageExports(pkg)
		for _, file := range pkg.Files {
			buildComponents(wd, file)
		}
	}
}

func writeGoDomFile(compiler *HtmlCompiler, htmlNode *html.Node, outputFileName string) {
	ofile, err := os.Create(outputFileName)
	defer ofile.Close()
	checkFatal(err)

	ctree := compiler.generate(htmlNode)
	writeCodeGofmt(ofile, outputFileName, ctree)

	if mess := compiler.Error(); mess != "" {
		fatal(mess)
	}
}

func buildHtmlFile(filename string) {
	outputFileName := filename + ".go"

	ifile, err := os.Open(filename)
	defer ifile.Close()
	checkFatal(err)

	n, err := htmlutils.ParseFragment(ifile)
	checkFatal(err)

	writeGoDomFile(NewHtmlCompiler(), n[0], outputFileName)
}
