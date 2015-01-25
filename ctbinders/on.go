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

		pdStr := "__event.PreventDefault();"
		spStr := ""
		if eventType == "click" {
			spStr = "__event.StopPropagation();"
		}

		if len(args) > 1 {
			for _, arg := range args[1:] {
				switch arg {
				case "OPT_NOPD":
					pdStr = ""
				case "OPT_NOSP":
					spStr = ""
				}
			}
		}

		fStr += d.Idt + "\t\t" + `__node.Attrs["on` + eventType +
			fmt.Sprintf(`"] = func(__event dom.Event) { %v %v %v }`, pdStr, spStr, expr)
		return
	}
}
