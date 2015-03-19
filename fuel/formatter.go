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
	emitCodeNaive(w, root)
	w.Close()

	cmd := exec.Command("go", "fmt", file)
	out, err := cmd.CombinedOutput()
	fmt.Println(string(out))
	if err != nil {
		fmt.Println(`fuel requires that the standard go command is available.` +
			`Please make sure it works.`)
		log.Fatal(err)
	}
}

func emitCodeNaive(w io.Writer, node *codeNode) {
	switch node.typ {
	case StringCodeNode:
		write(w, fmt.Sprintf(`"%v"`, node.code))

	case NakedCodeNode:
		write(w, node.code)

	case FuncCallCodeNode:
		write(w, node.code+"(")
		for i, c := range node.children {
			emitCodeNaive(w, c)
			if i < len(node.children)-1 {
				write(w, ",")
			}
		}
		write(w, ")")

	case SliceVarCodeNode:
		panic("Unexpected case, something is wrong, please report this.")

	case ElemListCodeNode, AppendListCodeNode:
		handleElemListCN(w, node)

	case VarDeclAreaCodeNode:
		write(w, "\n")
		for _, c := range node.children {
			emitCodeNaive(w, c)
			write(w, "\n")
		}

	case CompositeCodeNode, BlockCodeNode:
		write(w, node.code+"{\n")
		for _, c := range node.children {
			emitCodeNaive(w, c)
			if node.typ == CompositeCodeNode {
				write(w, ",")
			}
			write(w, "\n")
		}
		write(w, "}")
	}
}

func handleElemListCN(w io.Writer, node *codeNode) {
	opening := ElementListOpener + "{\n"
	closing := "}"
	svEnding := ""
	if node.typ == AppendListCodeNode {
		opening = "append(" + node.children[0].code + ", "
		closing = ")"
		svEnding = "..."
		node.children = node.children[1:]
	}

	// separate the list into parts to facilitate special constructs (for and if)
	parts := [][]*codeNode{node.children}
	if len(node.children) > 1 {
		parts = [][]*codeNode{make([]*codeNode, 0)}
		i := 0
		for _, c := range node.children {
			if c.typ == SliceVarCodeNode {
				parts = append(parts, []*codeNode{c})
				parts = append(parts, []*codeNode{})
				i = len(parts) - 1
			} else {
				parts[i] = append(parts[i], c)
			}
		}
	}

	write(w, node.code)
	for i, part := range parts {
		if i > 0 && len(parts[i-1]) > 0 {
			write(w, " + \n")
		}

		if len(part) == 1 && part[0].typ == SliceVarCodeNode {
			write(w, opening+part[0].code+svEnding+closing)
			continue
		}

		if len(part) == 0 {
			continue
		}

		write(w, opening)
		for i, c := range part {
			emitCodeNaive(w, c)
			if node.typ != AppendListCodeNode || i < len(part)-1 {
				write(w, ",\n")
			}
		}
		write(w, closing)
	}
}
