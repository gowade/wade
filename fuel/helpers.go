package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
)

func isFuelFile(fileName string) bool {
	return strings.HasSuffix(fileName, fuelSuffix)
}

func isCapitalized(name string) bool {
	c := []rune(name)[0]
	return c >= 'A' && c <= 'Z'
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
