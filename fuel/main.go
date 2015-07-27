package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/gowade/html"
	"github.com/gowade/wade/utils/htmlutils"
)

const (
	defaultIndexFile = "public/index.html"
)

func buildCmd(dir string, target string) {
	if target != "" {
		buildHtmlFile(target)
	} else {
		fuelBuild(dir, "")
	}
}

func serveCmd(dir string, args []string) {
	var (
		indexFile string
		port      string
		serveOnly bool
	)

	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	fs.StringVar(&indexFile, "i", defaultIndexFile, "HTML index file for your application. The compiled app.js file will be put into its directory")
	fs.StringVar(&port, "p", "8888", "HTTP port to serve the application")
	fs.BoolVar(&serveOnly, "serveonly", false, "Only serve and watch, no code generation")
	fs.Parse(args)

	if _, err := os.Stat(indexFile); err != nil {
		fatal(err.Error())
	}

	if serveOnly {
		fmt.Println("Running serve-only mode, fuel doesn't generate code..")
	}

	//NewFuel().Serve(dir, indexFile, port, serveOnly)
}

func cleanCmd(dir string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fatal(err.Error())
	}

	for _, file := range files {
		if isFuelFile(file.Name()) {
			os.Remove(file.Name())
		}
	}
}

func generateVDOMFile(htmlNode *html.Node, outputFileName string) error {
	ofile, err := os.Create(outputFileName)
	defer ofile.Close()
	checkFatal(err)

	preludeTpl.Execute(ofile, preludeTD{
		Pkg: "main",
	})

	return compileHTMLFile(outputFileName, ofile, htmlNode)
}

func buildHtmlFile(filename string) {
	outputFileName := filename + ".go"

	ifile, err := os.Open(filename)
	defer ifile.Close()
	checkFatal(err)

	n, err := htmlutils.ParseFragment(ifile)
	checkFatal(err)

	err = generateVDOMFile(n[0], outputFileName)
	checkFatal(err)

	runGofmt(outputFileName)
}

func runGofmt(file string) {
	cmd := exec.Command("go", "fmt", file)
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if err != nil {
		fmt.Fprintf(os.Stderr, "go fmt failed with %v\n", err.Error())
	}
}

func main() {
	flag.Parse()

	dir, err := os.Getwd()
	checkFatal(err)

	command := flag.Arg(0)
	switch command {
	case "build":
		buildCmd(dir, flag.Arg(1))

	case "serve":
		serveCmd(dir, flag.Args()[1:])

	case "clean":
		cleanCmd(dir)

	default:
		fatal("Please specify a command. Available commands: build, serve, clean")
	}
}
