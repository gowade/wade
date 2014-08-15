package wade

import (
	"fmt"

	"github.com/gopherjs/gopherjs/js"
)

type jsDep struct {
	url         string
	name        string
	checkSymbol string
	bowerpkg    string
	ainfo       string
}

var JsDepSymbols = []jsDep{
	jsDep{
		name:        "jQuery",
		url:         "jquery.com",
		checkSymbol: "jQuery",
		bowerpkg:    "jquery",
	},
	jsDep{
		name:        "history API",
		url:         "https://github.com/devote/HTML5-History-API",
		checkSymbol: "history",
		bowerpkg:    "html5-history-api",
	},
	jsDep{
		name:        "Watch.js",
		url:         "https://github.com/melanke/Watch.JS",
		checkSymbol: "watch",
		bowerpkg:    "wade-watch-js",
		ainfo: `Wade requires a modified version, which is "wade-watch-js" for bower,` +
			`and is at https://github.com/phaikawl/Watch.JS`,
	},
}

func jsDepCheck() {
	for _, dep := range JsDepSymbols {
		if js.Global.Get(dep.checkSymbol).IsUndefined() {
			panic(fmt.Sprintf(`The javascript dependency "%v" (%v) is not available. `+
				`It is in the bower package "%v", please install and use the required javascript file. `+
				`Additional info: "%v".`, dep.name, dep.url, dep.bowerpkg))
		}
	}
}
