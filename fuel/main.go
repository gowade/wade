package main

import (
	"os"

	"github.com/gowade/wade/utils/htmlutils"
)

const (
	TestFile = "test/test.html"
	TestOut  = "test/test.go"
)

func main() {
	ifile, err := os.Open(TestFile)
	if err != nil {
		panic(err)
	}

	ofile, err := os.Create(TestOut)
	if err != nil {
		panic(err)
	}

	n, err := htmlutils.ParseFragment(ifile)
	if err != nil {
		panic(err)
	}

	ctree := generate(n[0])
	writeCodeGofmt(ofile, TestOut, ctree)
}
