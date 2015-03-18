package main

import (
	"os"
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

	n, err := parseFragment(ifile)
	if err != nil {
		panic(err)
	}

	ctree := generate(n[0])
	writeCodeGofmt(ofile, TestOut, ctree)
}
