package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/gowade/whtml"
)

func must(err error) error {
	if err != nil {
		panic(err)
	}
	return err
}

func isFuelFile(fileName string) bool {
	return strings.HasSuffix(fileName, fuelSuffix)
}

func isCapitalized(name string) bool {
	c := []rune(name)[0]
	return c >= 'A' && c <= 'Z'
}

func whtmlParseElem(rd io.Reader) (*whtml.Node, error) {
	nodes, err := whtml.Parse(rd)
	var ret *whtml.Node
	if len(nodes) > 0 {
		ret = nodes[0]
	}

	return ret, err
}

func cleanGarbageTextChildren(node *whtml.Node) {
	prev := node.FirstChild
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == whtml.TextNode && strings.TrimSpace(c.Data) == "" {
			if c == node.FirstChild {
				node.FirstChild = c.NextSibling
				prev = node.FirstChild
			} else {
				prev.NextSibling = c.NextSibling
				if c == node.LastChild {
					node.LastChild = prev
				}
			}
		} else {
			prev = c
		}
	}
}

func fatal(msg string, fmtargs ...interface{}) {
	fmt.Fprintf(os.Stdout, msg+"\n", fmtargs...)
	os.Exit(2)
}

func printErr(err error) {
	fmt.Fprintf(os.Stdout, "%v\n", err)
}

func checkFatal(err error) {
	if err != nil {
		fatal(err.Error())
	}
}
func efmt(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

func sfmt(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

func execTplBuf(tpl *template.Template, data interface{}) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	err := tpl.Execute(&buf, data)
	if err != nil {
		return nil, err
	}

	return &buf, nil
}
