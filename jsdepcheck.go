package wade

import (
	"fmt"
)

type DepChecker interface {
	CheckJsDep(dep string) bool
}

type JsDep struct {
	Url         string
	Name        string
	CheckSymbol string
	Bowerpkg    string
	Ainfo       string
}

var JsDepSymbols = []JsDep{
	JsDep{
		Name:        "jQuery",
		Url:         "jquery.com",
		CheckSymbol: "jQuery",
		Bowerpkg:    "jquery",
	},
	JsDep{
		Name:        "history API",
		Url:         "https://github.com/devote/HTML5-History-API",
		CheckSymbol: "history",
		Bowerpkg:    "html5-history-api",
	},
	JsDep{
		Name:        "polymer observe.js",
		Url:         "https://github.com/Polymer/observe-js",
		CheckSymbol: "PathObserver",
		Bowerpkg:    "observe-js",
	},
}

func jsDepCheck(depCheckImp DepChecker) {
	for _, dep := range JsDepSymbols {
		if !depCheckImp.CheckJsDep(dep.CheckSymbol) {
			panic(fmt.Sprintf(`The javascript dependency "%v" (%v) is not available. `+
				`It is in the bower package "%v", please install and use the required javascript file. `+
				`Additional info: "%v".`, dep.Name, dep.Url, dep.Bowerpkg, dep.Ainfo))
		}
	}
}
