package main

import (
	"bytes"
	"io"
	//"fmt"

	"github.com/gowade/html"
)

func (c *HTMLCompiler) elementGenerate(w io.Writer, el *html.Node) error {

}

func (c *HTMLCompiler) GenerateFile(w io.Writer, node *html.Node) error {
	var buf bytes.Buffer
	err := c.elementGenerate(&buf, node)
	if err != nil {
		return err
	}

	renderFuncTpl.Execute(w, renderFuncTD{
		Return: &buf,
	})

	return nil
}
