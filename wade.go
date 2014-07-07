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
	TempReplaceRegexp = regexp.MustCompile(`<%([\(\)\.\-_` + "`" + `\w\s]+)%>`)
)

// parseTemplate replaces "<% bindstr %>" with <span bind-html="bindstr"></span>
func parseTemplate(source string) string {
	return TempReplaceRegexp.ReplaceAllStringFunc(source, func(m string) string {
		bindstr := strings.TrimSpace(TempReplaceRegexp.FindStringSubmatch(m)[1])
		return fmt.Sprintf(`<span bind-html="%v"></span>`, bindstr)
	})
}

// WadeUp initializes the Wade engine and performs HTML imports.
//
// "startPage" is the id of the page we redirect to on an access to /
//
// "tempcontainer" is the id of the <script type="text/wadin"> element where the HTML source code reside in.
// It usually contains <wimport> elements to import other HTML source files.
// It is never displayed and is ignored by the browser, screen readers, etc.
//
// "container" is the id of the HTML parent element where all the real page content is copied into and displayed.
//
// "initFn" is the callback that is run after initialization finishes.
func WadeUp(startPage, basePath string, tempcontainer, container string, initFn func(*Wade)) *Wade {
	gHistory = js.Global.Get("history")
	origin := js.Global.Get("document").Get("location").Get("origin").Str()
	tempContainer := gJQ("script[type='text/wadin']#" + tempcontainer)
	if tempContainer.Length == 0 {
		panic(fmt.Sprintf("Template container #%v not found or is wrong kind of element, must be script[type='text/wadin'].",
			tempContainer))
	}
	html := js.Global.Get(jq.JQ).Call("parseHTML", "<div>"+parseTemplate(tempContainer.Html())+"</div>")
	tElem := gJQ(html)
	htmlImport(tElem, origin)
	tm := newCustagMan(tElem)
	binding := bind.NewBindEngine(tm)
	wd := &Wade{
		pm:         newPageManager(startPage, basePath, container, tElem, binding, tm),
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

// htmlImport performs importing of HTML source
func htmlImport(parent jq.JQuery, origin string) {
	parent.Find("wimport").Each(func(i int, elem jq.JQuery) {
		src := elem.Attr("src")
		req := http.NewRequest(http.MethodGet, origin+src)
		html := req.DoSync().Data()
		elem.Append(parseTemplate(html))
		htmlImport(elem, origin)
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
