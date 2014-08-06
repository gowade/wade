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
	serverbase string
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
	jsDepCheck()

	gHistory = js.Global.Get("history")
	serverbase := js.Global.Get("document").Get("location").Get("origin").Str()
	tempContainers := gJQ("script[type='text/wadin']")
	jqParseHTML := func(src string) jq.JQuery {
		return gJQ(js.Global.Get(jq.JQ).Call("parseHTML", src))
	}
	tElem := jqParseHTML("<div></div>")
	tempContainers.Each(func(_ int, container jq.JQuery) {
		tElem.Append(container.Html())
	})

	htmlImport(tElem, serverbase)
	tm := newCustagMan(tElem)
	binding := bind.NewBindEngine(tm)
	wd := &Wade{
		pm:         newPageManager(startPage, basePath, tElem, binding, tm),
		tm:         tm,
		binding:    binding,
		tcontainer: tElem,
		serverbase: serverbase,
	}
	wd.init()
	initFn(wd)
	return wd
}

// Pager returns the Page Manager
func (wd *Wade) Pager() *PageManager {
	return wd.pm
}

// RegisterCustomTags registers custom element tags declared inside a given html file
// srcFile and associate them with given model prototypes. srcFile is used
// like when using <wimport>.
//
// For reach <welement> inside the file, it registers a new tag with the name
// specified by its "tagname" attribute.
// The template content and specifications of the tag is taken from the <welement>.
// For example, if there's a <welement> inside "public/elements.html":
//	<welement tagname="errorlist" attributes="Errors Subject">
//		<p>errors for <% Subject %></p>
//		<ul>
//			<li bind-each="Errors -> _, msg"><% msg %></li>
//		</ul>
//	</welement>
// If wade.RegisterNew("public/elements.html", prototype)
// is called, a new html tag "errorlist" will be registered.
//
// This new tag may be used like this
//	<errorlist attr-subject="Username" bind="Errors: Username.Errors"></errorlist>
// And if "Username.Errors" is {"Invalid.", "Not enough chars."}, something like this will
// be put in place of the above:
//	<p>errors for Username</p>
//	<ul>
//		<li>Invalid.</li>
//		<li>Not enough chars.</li>
//	</ul>
// The prototype is a struct which specifies datatypes for the custom element's
// attributes. If it's pointer version has a method "Init" which satisfies the
// CustomElementInit interface, Init will be called
// when the custom element is processed.
func (wd *Wade) RegisterCustomTags(srcFile string, protomap map[string]interface{}) {
	elems := wd.GetHtml(srcFile)
	tagElems := make([]jq.JQuery, 0)
	elems.Each(func(_ int, elem jq.JQuery) {
		if elem.Is("welement") {
			tagElems = append(tagElems, elem)
		}
	})
	wd.tm.registerTags(tagElems, protomap)
}

// Binding returns the binding engine
func (wd *Wade) Binding() *bind.Binding {
	return wd.binding
}

// GetHtml makes a request and gets the HTML contents
func (wd *Wade) GetHtml(href string) jq.JQuery {
	return getHtmlFile(wd.serverbase, href)
}

func getHtmlFile(serverbase string, href string) jq.JQuery {
	req := http.NewRequest(http.MethodGet, serverbase+href)
	html := req.DoSync().Data()
	return gJQ(parseTemplate(html))
}

// htmlImport performs an HTML import
func htmlImport(parent jq.JQuery, serverbase string) {
	parent.Find("wimport").Each(func(i int, elem jq.JQuery) {
		src := elem.Attr("src")
		ne := getHtmlFile(serverbase, src)
		elem.ReplaceWith(ne)
		htmlImport(ne, serverbase)
	})
}

func (wd *Wade) init() {
	bind.RegisterInternalHelpers(wd.pm, wd.binding)
}

// Start starts the real operation, meant to be called at the end of everything.
func (wd *Wade) Start() {
	gJQ(js.Global.Get("document")).Ready(func() {
		wd.pm.prepare()
	})
}
