package wade

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	jq "github.com/gopherjs/jquery"
	"github.com/phaikawl/wade/bind"
	"github.com/phaikawl/wade/libs/http"
	"github.com/phaikawl/wade/libs/pdata"
)

var (
	gHistory    js.Object
	gJQ         = jq.NewJQuery
	WadeDevMode = true
)

type wade struct {
	errChan    chan error
	appEnv     AppEnv
	pm         *pageManager
	tm         *custagMan
	tcontainer jq.JQuery
	binding    *bind.Binding
	serverbase string
	customTags map[string]map[string]interface{}
}

type registry struct {
	w *wade
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
func (r registry) RegisterCustomTags(srcFile string, protomap map[string]interface{}) {
	r.w.customTags[srcFile] = protomap
}

// ModuleInit calls the modules' Init method with an AppEnv
func (r registry) ModuleInit(modules ...NeedsInit) {
	for _, module := range modules {
		module.Init(r.w.appEnv)
	}
}

// RegisterController adds a new controller function for the specified
// page / page group.
func (r registry) RegisterController(displayScope string, fn PageControllerFunc) {
	r.w.pm.registerController(displayScope, fn)
}

// RegisterDisplayScopes registers the given map of pages and page groups
func (r registry) RegisterDisplayScopes(m map[string]DisplayScope) {
	r.w.pm.registerDisplayScopes(m)
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

type AppFunc func(Registration, AppEnv)

// StartApp gets and processes HTML source from script[type="text/wadin"]
// elements, performs HTML imports and initializes the app.
//
// "startPage" is the id of the page we redirect to on an access to /
//
// "appFn" is the main function for your app.
func StartApp(startPage, basePath string, appFn AppFunc) error {
	jsDepCheck()

	gHistory = js.Global.Get("history")
	//serverbase := js.Global.Get("document").Get("location").Get("origin").Str()
	serverbase := "/"
	tempContainers := gJQ("script[type='text/wadin']")
	jqParseHTML := func(src string) jq.JQuery {
		return gJQ(js.Global.Get(jq.JQ).Call("parseHTML", src))
	}
	tElem := jqParseHTML("<div></div>")
	tempContainers.Each(func(_ int, container jq.JQuery) {
		tElem.Append(container.Html())
	})

	err := htmlImport(tElem, serverbase)
	if err != nil {
		return err
	}
	tm := newCustagMan(tElem)
	binding := bind.NewBindEngine(tm)
	appEnv := AppEnv{
		Services: AppServices{
			Http:           http.NewClient(),
			SessionStorage: pdata.Service(pdata.SessionStorage),
			LocalStorage:   pdata.Service(pdata.LocalStorage),
		},
	}

	wd := &wade{
		errChan:    make(chan error),
		appEnv:     appEnv,
		pm:         newPageManager(appEnv, startPage, basePath, tElem, binding, tm),
		tm:         tm,
		binding:    binding,
		tcontainer: tElem,
		serverbase: serverbase,
		customTags: make(map[string]map[string]interface{}),
	}

	appEnv.PageManager = wd.pm
	wd.init()

	appFn(registry{wd}, appEnv)

	queueChan := make(chan bool, 100)
	finishChan := make(chan bool, len(wd.customTags))
	for srcFile, protoMap := range wd.customTags {
		go func(srcFile string, protoMap map[string]interface{}) {
			elems, err := wd.getHtml(srcFile)

			queueChan <- true

			if len(queueChan) == len(wd.customTags) {
				close(queueChan)
				finishChan <- true
			}

			if err != nil {
				wd.errChan <- fmt.Errorf(`Cannot load custom tag HTML file "%v".`, srcFile)
				return
			}

			tagElems := make([]jq.JQuery, 0)
			elems.Each(func(_ int, elem jq.JQuery) {
				if elem.Is("welement") {
					tagElems = append(tagElems, elem)
				}
			})

			wd.tm.registerTags(tagElems, protoMap)
		}(srcFile, protoMap)
	}
	<-finishChan

	wd.start()

	select {
	case err = <-wd.errChan:
		return err
	}

	return nil
}

// GetHtml makes a request and gets the HTML contents
func (wd *wade) getHtml(href string) (html jq.JQuery, err error) {
	html, err = getHtmlFile(wd.serverbase, href)
	return
}

func getHtmlFile(serverbase string, href string) (jq.JQuery, error) {
	resp, err := http.Do(http.NewRequest("GET", path.Join(serverbase, href)))
	if err != nil || resp.Failed() {
		return gJQ(), fmt.Errorf(`Failed to load HTML file "%v"`, href)
	}

	return gJQ(parseTemplate(resp.TextData)), nil
}

// htmlImport performs an HTML import
func htmlImport(parent jq.JQuery, serverbase string) error {
	for _, elem := range ToElemSlice(parent.Find("wimport")) {
		src := elem.Attr("src")
		var err error
		html := make(chan jq.JQuery)
		go func(elem jq.JQuery) {
			var ne jq.JQuery
			ne, err = getHtmlFile(serverbase, src)
			if err != nil {
				return
			}
			html <- ne
		}(elem)
		ne := <-html
		elem.ReplaceWith(ne)

		err = htmlImport(ne, serverbase)
		if err != nil {
			return err
		}
	}

	return nil
}

func (wd *wade) init() {
	bind.RegisterInternalHelpers(wd.pm, wd.binding)
}

// Start starts the real operation, meant to be called at the end of everything.
func (wd *wade) start() {
	gJQ(js.Global.Get("document")).Ready(func() {
		wd.pm.prepare()
	})
}
