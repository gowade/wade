package ctbinders

import (
	"fmt"

	"github.com/phaikawl/wade/compiler"
	"github.com/phaikawl/wade/core"
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
		fStr += d.Idt + "\t\t__node.Children = make([]*VNode, len(__data))\n"
		fStr += d.Idt + fmt.Sprintf("\t\tfor __index, __value := range __data { %v := __value \n", val)
		if key != "_" {
			fStr += d.Idt + fmt.Sprintf("\t\t\t%v := __index\n", key)
		}

		wrapper := core.VPrep(&core.VNode{
			Data:     "w_group",
			Type:     core.GroupNode,
			Children: d.Node.Children,
		})

		fStr += d.Idt + fmt.Sprintf("\t\t\t__node.Children[__index] = VPrep(&VNode%v)", d.Compiler.Process(wrapper, d.Depth+2, d.File))
		fStr += "\n" + d.Idt + "\t\t}"

		for _, c := range d.Node.Children {
			d.Compiler.PreventProcessing[c] = true
		}

		return
	}
}
