package ctbinders

import (
	"fmt"

	"github.com/phaikawl/wade/compiler"
)

func init() {
	Binders["for"] = func(d compiler.TempComplData, args []string, expr string) (fStr string) {
		key, val := "_", "_"
		if len(args) >= 1 {
			key = args[0]
		}

		if len(args) >= 2 {
			val = args[1]
		}

		fStr += d.Idt + fmt.Sprintf("\t\t__data := %v\n", expr)
		fStr += d.Idt + fmt.Sprintf("\t\tfor __index, %v := range __data {\n", val)
		if key != "_" {
			fStr += d.Idt + fmt.Sprintf("\t\t\t%v := __index\n", key)
		}
		fStr += d.Idt + "\t\t\t__node.Children = make(*VNode, len(__data))\n"
		fStr += d.Idt + fmt.Sprintf("\t\t\t__node.Children[__index] = %v", d.Compiler.Process(d.Node.Children[0], d.Depth+2, d.File))
		fStr += "\n" + d.Idt + "\t\t}"

		d.Compiler.PreventProcessing[d.Node] = true

		return
	}
}
