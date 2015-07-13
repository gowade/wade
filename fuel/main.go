package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gowade/html"
	"github.com/gowade/wade/utils/htmlutils"
)

const (
	defaultIndexFile = "public/index.html"
)

func fatal(msg string, fmtargs ...interface{}) {
	fmt.Fprintf(os.Stdout, msg+"\n", fmtargs...)
	os.Exit(2)
}

func checkFatal(err error) {
	if err != nil {
		fatal(err.Error())
	}
}

func main() {
	dir, err := os.Getwd()
	if err != nil {
		fatal(err.Error())
	}

	var indexFile string
	var port string

	flag.StringVar(&indexFile, "idx", defaultIndexFile, "HTML index file for your application. The compiled app.js file will be put into its directory")
	flag.StringVar(&port, "p", "8888", "HTTP port to serve the application")
	flag.Parse()

	command := flag.Arg(0)
	switch command {
	case "build":
		bTarget := flag.Arg(1)
		if bTarget != "" {
			buildHtmlFile(bTarget)
		} else {
			fuel := NewFuel()
			fuel.BuildPackage(dir, "", nil)
		}

	case "serve":
		if _, err := os.Stat(indexFile); err != nil {
			fatal(err.Error())
		}

		NewFuel().Serve(dir, indexFile, port)

	case "clean":
		dir, err := os.Getwd()
		if err != nil {
			fatal(err.Error())
		}

		files, err := ioutil.ReadDir(dir)
		if err != nil {
			fatal(err.Error())
		}

		for _, file := range files {
			if strings.HasSuffix(file.Name(), fuelSuffix) {
				os.Remove(file.Name())
			}
		}

	default:
		fatal("Please specify a command.")
	}
}

func compileDomFile(compiler *HTMLCompiler, htmlNode *html.Node, outputFileName, pkgName string, com *componentInfo) error {
	ofile, err := os.Create(outputFileName)
	defer ofile.Close()
	checkFatal(err)

	ctree, err := compiler.Generate(htmlNode, com)
	if err != nil {
		return err
	}

	write(ofile, prelude(pkgName, nil))
	emitDomCode(ofile, ctree)

	return nil
}

func buildHtmlFile(filename string) {
	outputFileName := filename + ".go"

	ifile, err := os.Open(filename)
	defer ifile.Close()
	checkFatal(err)

	n, err := htmlutils.ParseFragment(ifile)
	checkFatal(err)

	err = compileDomFile(NewHTMLCompiler(nil), n[0], outputFileName, "main", nil)
	if err != nil {
		fatal(err.Error())
	}

	runGofmt(outputFileName)
}
