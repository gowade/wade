package ctbinders

import (
	"fmt"
	"strings"

	"github.com/phaikawl/wade/compiler"
)

func init() {
	Binders["on"] = func(d compiler.TempComplData, args []string, expr string) (fStr string) {
		eventType := "click"
		if len(args) > 0 {
			eventType = args[0]
		}

		pdStr := "__event.PreventDefault()"
		if len(args) > 1 && strings.ToLower(args[1]) == "false" {
			pdStr = ""
		}

		fStr += d.Idt + "\t\t" + `__node.Attrs["on` + eventType +
			fmt.Sprintf(`"] = func(__event dom.Event) { %v; %v }`, pdStr, expr)
		return
	}
}
