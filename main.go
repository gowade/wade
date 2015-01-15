package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/phaikawl/wade/compiler"
	"github.com/phaikawl/wade/core"
	"github.com/phaikawl/wade/ctbinders"
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/dom/gonet"
)

var (
	gd           = gonet.GetDom()
	groupNodeStr = fmt.Sprintf("<%v></%v>", core.GroupNodeTagName)
)

func parseHTML(filePath string) dom.Selection {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	file, err := os.Open(path.Join(wd, filePath))
	if err != nil {
		panic(err)
	}

	return gonet.NewFragment(file)
}

func importHtml(node dom.Selection) {
	for _, inclNode := range node.Find(core.IncludeTagName).Elements() {
		src, ok := inclNode.Attr("src")
		if !ok {
			continue
		}

		repl := gd.NewFragment(groupNodeStr)
		repl.SetAttr("src", src)

		if belongStr, hasBelong := inclNode.Attr("_belong"); hasBelong {
			repl.SetAttr("_belong", belongStr)
		}

		repl.Append(parseHTML(src))
		importHtml(repl)
		inclNode.ReplaceWith(repl)
	}
}

func main() {
	var (
		flagMasterFile = flag.String("f", "public/main.html", "main template file")
		flagOutputDir  = flag.String("o", "client", "output directory")

		masterFile = *flagMasterFile
		outputDir  = *flagOutputDir
	)

	root := gd.NewFragment(groupNodeStr)
	root.Append(parseHTML(masterFile))

	importHtml(root)
	pkgName := path.Base(outputDir)

	vRoot := root.ToVNode()
	c := compiler.NewCompiler(outputDir, pkgName, ctbinders.Binders)
	c.CompileRoot(masterFile, vRoot)
}
