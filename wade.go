package wade

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	jq "github.com/gopherjs/jquery"
	"github.com/phaikawl/wade/bind"
	"github.com/phaikawl/wade/services/http"
)

var (
	gHistory    js.Object
	gJQ         = jq.NewJQuery
	WadeDevMode = true
)

type Wade struct {
	pm         *PageManager
	tm         *CustagMan
	tcontainer jq.JQuery
	binding    *bind.Binding
}

var (
	TempReplaceRegexp = regexp.MustCompile(`<%([^"<>]+)%>`)
)

// parseTemplate replaces "<% bindstr %>" with <span bind-html="bindstr"></span>
func parseTemplate(source string) string {
	return TempReplaceRegexp.ReplaceAllStringFunc(source, func(m string) string {
		bindstr := strings.TrimSpace(TempReplaceRegexp.FindStringSubmatch(m)[1])
		return fmt.Sprintf(`<span bind-html="%v"></span>`, bindstr)
	})
}

// WadeUp gets and processes HTML source from script[type="text/wadin"]
// elements, performs HTML imports and initializes the app.
//
// "startPage" is the id of the page we redirect to on an access to /
//
// "initFn" is the callback that is run after initialization finishes.
func WadeUp(startPage, basePath string, initFn func(*Wade)) *Wade {
	gHistory = js.Global.Get("history")
	origin := js.Global.Get("document").Get("location").Get("origin").Str()
	tempContainers := gJQ("script[type='text/wadin']")
	jqParseHTML := func(src string) jq.JQuery {
		return gJQ(js.Global.Get(jq.JQ).Call("parseHTML", src))
	}
	tElem := jqParseHTML("<div></div>")
	tempContainers.Each(func(_ int, container jq.JQuery) {
		tElem.Append(container.Html())
	})

	htmlImport(tElem, origin)
	tm := newCustagMan(tElem)
	binding := bind.NewBindEngine(tm)
	wd := &Wade{
		pm:         newPageManager(startPage, basePath, tElem, binding, tm),
		tm:         tm,
		binding:    binding,
		tcontainer: tElem,
	}
	wd.init()
	initFn(wd)
	return wd
}

// Pager returns the Page Manager
func (wd *Wade) Pager() *PageManager {
	return wd.pm
}

// Custags returns the Custom Tags Manager
func (wd *Wade) Custags() *CustagMan {
	return wd.tm
}

// Binding returns the binding engine
func (wd *Wade) Binding() *bind.Binding {
	return wd.binding
}

// htmlImport performs an HTML import
func htmlImport(parent jq.JQuery, origin string) {
	parent.Find("wimport").Each(func(i int, elem jq.JQuery) {
		src := elem.Attr("src")
		req := http.NewRequest(http.MethodGet, origin+src)
		html := req.DoSync().Data()
		ne := gJQ(parseTemplate(html))
		elem.ReplaceWith(ne)
		htmlImport(ne, origin)
	})
}

func (wd *Wade) init() {
	bind.RegisterUrlHelper(wd.pm, wd.binding)
}

// Start starts the real operation, meant to be called at the end of everything.
func (wd *Wade) Start() {
	gJQ(js.Global.Get("document")).Ready(func() {
		wd.tm.prepare()
		wd.pm.getReady()
	})
}
