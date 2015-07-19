package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

func write(w io.Writer, content string) {
	w.Write([]byte(content))
}

func runGofmt(file string) {
	cmd := exec.Command("go", "fmt", file)
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if err != nil {
		fmt.Fprintf(os.Stderr, "go fmt failed with %v\n", err.Error())
	}
}

func emitDomCode(w io.Writer, node *codeNode, err error) {
	if node == nil {
		if err != nil {
			w.Write([]byte("!ERROR: " + err.Error()))
		}
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
			emitDomCode(w, c, err)
			if i < len(node.children)-1 {
				write(w, ",")
			}
		}
		write(w, ")")

	case SliceVarCodeNode:
		panic("Unexpected case, something is wrong, please report this.")

	case ElemListCodeNode, AppendListCodeNode:
		handleElemListCN(w, node, err)

	case VarDeclAreaCodeNode:
		write(w, "\n")
		for _, c := range node.children {
			emitDomCode(w, c, err)
			write(w, "\n")
		}

	case WrapperCodeNode:
		write(w, node.code)
		for _, c := range node.children {
			emitDomCode(w, c, err)
			write(w, "\n")
		}

	case CompositeCodeNode, BlockCodeNode:
		write(w, node.code+"{\n")
		for _, c := range node.children {
			emitDomCode(w, c, err)
			if node.typ == CompositeCodeNode {
				write(w, ",")
			}
			write(w, "\n")
		}
		write(w, "}")
	}
}

func handleElemListCN(w io.Writer, node *codeNode, err error) {
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
				emitDomCode(w, c, err)
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
