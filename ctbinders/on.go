package ctbinders

import (
	"fmt"

	"github.com/phaikawl/wade/compiler"
)

func init() {
	Binders["on"] = func(d compiler.TempComplData, args []string, expr string) (fStr string) {
		eventType := "click"
		if len(args) > 0 {
			eventType = args[0]
		}

		fStr += d.Idt + "\t\t" + `__node.Attrs["on` + eventType +
			fmt.Sprintf(`"] = func(__event dom.Event) { %v }`, expr)
		return
	}
}
