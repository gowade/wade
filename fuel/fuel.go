package main

import (
	"go/ast"
	"os"
	"path"
	"path/filepath"
)

const (
	fuelSuffix  = ".fuel.go"
	genComsFile = "~generatedComs~" + fuelSuffix
	methodsFile = "~methods~" + fuelSuffix

	importSTag = "import"
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

func fuelBuild(dir string, comPrefix string) {
	pkg, err := getFuelPkg(dir)
	checkFatal(err)

	for _, file := range pkg.htmlFiles {
		for _, com := range file.comDefs {
			_ = com
		}
	}
}
