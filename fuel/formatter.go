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

func runGofmt(file string) {
	cmd := exec.Command("go", "fmt", file)
	out, err := cmd.CombinedOutput()
	fmt.Printf(`Running "go fmt" on file "%v". Output: %v`+"\n", file, string(out))
	if err != nil {
		fmt.Println(`fuel requires that the standard go command is available. ` +
			`Please make sure it works.`)
		log.Fatal(err)
	}
}

func writeCodeNaive(w io.WriteCloser, file, pkgName string, root *codeNode) {
	write(w, Prelude(pkgName))
	emitCodeNaive(w, root)
	w.Close()
}

func emitCodeNaive(w io.Writer, node *codeNode) {
	if node == nil {
		write(w, "<<NIL>>")
		return
	}

	switch node.typ {
	case StringCodeNode:
		write(w, fmt.Sprintf(`%q`, node.code))

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
	opening := NodeListOpener + "{\n"
	closing := "}"

	//isAppend := node.typ == AppendListCodeNode

	// separate the list into parts to facilitate special constructs (for and if)
	rparts := [][]*codeNode{node.children}
	if len(node.children) > 1 {
		rparts = [][]*codeNode{make([]*codeNode, 0)}
		i := 0
		for _, c := range node.children {
			if c.typ == SliceVarCodeNode {
				rparts = append(rparts, []*codeNode{c})
				rparts = append(rparts, []*codeNode{})
				i = len(rparts) - 1
			} else {
				rparts[i] = append(rparts[i], c)
			}
		}
	}

	parts := make([][]*codeNode, 0, len(rparts))
	for _, part := range rparts {
		if len(part) > 0 {
			parts = append(parts, part)
		}
	}

	write(w, node.code)
	for i, part := range parts {
		if i < len(parts)-1 {
			write(w, "append(")
		}

		if len(part) == 1 && part[0].typ == SliceVarCodeNode {
			write(w, part[0].code)
		} else {
			write(w, opening)
			for i, c := range part {
				emitCodeNaive(w, c)
				if i < len(part)-1 {
					write(w, ",\n")
				}
			}
			write(w, closing)
		}

		if i < len(parts)-1 {
			write(w, ", ")
		}
	}

	if len(parts) > 1 {
		for range parts[1:] {
			write(w, "...)")
		}
	}
}
