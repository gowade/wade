package main

import (
	"fmt"
	"io"
	"log"
	"os/exec"
)

func write(w io.Writer, content string) {
	w.Write([]byte(content))
}

func writeCodeGofmt(w io.WriteCloser, file string, root *codeNode) {
	write(w, Prelude)
	write(w, "func Render() {\n return ")
	emitCodeNaive(w, root)
	write(w, "}")
	w.Close()

	cmd := exec.Command("go", "fmt", file)
	out, err := cmd.CombinedOutput()
	log.Println(string(out))
	if err != nil {
		log.Fatal(err)
	}
}

func emitCodeNaive(w io.Writer, node *codeNode) {
	switch node.typ {
	case stringCodeNode:
		write(w, fmt.Sprintf(`"%v"`, node.code))
	case nakedCodeNode:
		write(w, node.code)
	case funcCallCodeNode:
		write(w, node.code+"(")
		for i, c := range node.children {
			emitCodeNaive(w, c)
			if i < len(node.children)-1 {
				write(w, ",")
			}
		}
		write(w, ")")
	case compositeCodeNode, funcDeclCodeNode:
		write(w, node.code+"{\n")
		for _, c := range node.children {
			emitCodeNaive(w, c)
			if node.typ == compositeCodeNode {
				write(w, ",")
			}
			write(w, "\n")
		}
		write(w, "}")
	}
}
