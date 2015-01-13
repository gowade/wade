package ctbinders

import (
	"fmt"

	"github.com/phaikawl/wade/compiler"
)

func init() {
	Binders["for"] = func(d compiler.TempComplData) (fStr string) {
		key, val := "_", "_"
		if len(d.Args) >= 1 {
			key = d.Args[0]
		}

		if len(d.Args) >= 2 {
			val = d.Args[1]
		}

		fStr += d.Idt + fmt.Sprintf("\t\t__data := %v\n", d.Expr)
		fStr += d.Idt + fmt.Sprintf("\t\tfor __index, %v := range __data {\n", val)
		if key != "_" {
			fStr += d.Idt + fmt.Sprintf("\t\t\t%v := __index\n", key)
		}
		fStr += d.Idt + "\t\t\t__node.Children = make(*wc.VNode, len(__data))\n"
		fStr += d.Idt + fmt.Sprintf("\t\t\t__node.Children[__index] = %v", d.Compiler.Process(d.Node.ChildElems()[0], d.Depth+2, d.File))
		fStr += "\n" + d.Idt + "\t\t}"

		d.Compiler.PreventProcessing[d.Node] = true

		return
	}
}
