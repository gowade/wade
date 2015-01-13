package ctbinders

import (
	"fmt"
	"strings"

	"github.com/phaikawl/wade/compiler"
)

func init() {
	Binders["on"] = func(d compiler.TempComplData) (fStr string) {
		eventType := "click"
		if len(d.Args) > 0 {
			eventType = d.Args[0]
		}

		expr := strings.Replace(d.Expr, "$event", "__evt", 1)
		fStr += d.Idt + "\t\t" + `__node.Attrs["on` + eventType +
			fmt.Sprintf(`"] = func(__evt dom.Event) { %v }`, expr)
		return
	}
}
