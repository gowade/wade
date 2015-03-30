package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gowade/wade/utils/htmlutils"
)

const (
	TestFile = "test/test.html"
)

func main() {
	targetFilename := TestFile
	flag.Parse()
	if flag.Arg(0) != "" {
		targetFilename = flag.Arg(0)
	}

	outputFileName := targetFilename + ".go"

	ifile, err := os.Open(targetFilename)
	if err != nil {
		panic(err)
	}

	ofile, err := os.Create(outputFileName)
	if err != nil {
		panic(err)
	}

	n, err := htmlutils.ParseFragment(ifile)
	if err != nil {
		panic(err)
	}

	c := NewCompiler()
	ctree := c.generate(n[0])
	writeCodeGofmt(ofile, outputFileName, ctree)

	if mess := c.Error(); mess != "" {
		fmt.Println(mess)
		os.Exit(2)
	}
}
