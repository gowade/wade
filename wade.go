package wade

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	jq "github.com/gopherjs/jquery"
	"github.com/phaikawl/wade/bind"
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/dom/jquery"
	"github.com/phaikawl/wade/jsbackend"
	"github.com/phaikawl/wade/libs/http"
	"github.com/phaikawl/wade/libs/http/clientside"
	"github.com/phaikawl/wade/libs/pdata"
)

var (
	gJQ         = jq.NewJQuery
	WadeDevMode = true
)

type (
	wade struct {
		errChan    chan error
		pm         *pageManager
		tm         *custagMan
		tcontainer jq.JQuery
		binding    *bind.Binding
		serverBase string
		customTags map[string]map[string]interface{}
	}

	registry struct {
		w *wade
	}

	JsBackend interface {
		DepChecker
		History() History
	}

	jsBackend struct {
		*jsbackend.BackendImp
	}
)

func (b jsBackend) History() History {
	return b.BackendImp.History
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
		module.Init(AppServices)
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

func initServices(pm PageManager) {
	AppServices.Http = http.NewClient(clientside.XhrBackend{})
	AppServices.LocalStorage = pdata.Service(pdata.LocalStorage)
	AppServices.SessionStorage = pdata.Service(pdata.SessionStorage)
	AppServices.PageManager = pm
}

func loadHtml(document dom.Selection) dom.Selection {
	templateContainer := document.NewRootFragment("<div></div>")
	temps := document.Find("script[type='text/wadin']")
	for _, part := range temps.Elements() {
		templateContainer.Append(part)
	}

	return templateContainer
}

// StartApp gets and processes HTML source from script[type="text/wadin"]
// elements, performs HTML imports and initializes the app.
//
// "appFn" is the main function for your app.
func StartApp(config AppConfig, appFn AppFunc) error {
	var jsb JsBackend = jsBackend{jsbackend.Get()}
	jsDepCheck(jsb)
	http.SetDefaultClient(http.NewClient(clientside.XhrBackend{}))

	dombackend := jquery.GetDom()
	var document dom.Selection = jquery.Document()

	templateContainer := loadHtml(document)

	err := htmlImport(http.DefaultClient(), templateContainer, config.ServerBase)
	if err != nil {
		return err
	}

	jqTemp := templateContainer.(jquery.Selection).JQuery

	tm := newCustagMan(jqTemp)
	binding := bind.NewBindEngine(tm)

	wd := &wade{
		errChan:    make(chan error),
		pm:         newPageManager(jsb.History(), config, jqTemp, binding, tm),
		tm:         tm,
		binding:    binding,
		tcontainer: jqTemp,
		serverBase: config.ServerBase,
		customTags: make(map[string]map[string]interface{}),
	}

	wd.init()
	initServices(wd.pm)

	appFn(registry{wd})

	queueChan := make(chan bool, 100)
	finishChan := make(chan bool, len(wd.customTags))
	for srcFile, protoMap := range wd.customTags {
		go func(srcFile string, protoMap map[string]interface{}) {
			src, err := wd.getHtml(http.DefaultClient(), srcFile)
			elems := dombackend.NewFragment(src).Elements()

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
			for _, elem := range elems {
				if elem.Is("welement") {
					tagElems = append(tagElems, elem.(jquery.Selection).JQuery)
				}
			}

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
func (wd *wade) getHtml(httpClient *http.Client, href string) (string, error) {
	return getHtmlFile(httpClient, wd.serverBase, href)
}

func getHtmlFile(httpClient *http.Client, serverbase string, href string) (string, error) {
	resp, err := httpClient.GET(path.Join(serverbase, href))
	if err != nil || resp.Failed() {
		return "", fmt.Errorf(`Failed to load HTML file "%v"`, href)
	}

	return parseTemplate(resp.Data), nil
}

// htmlImport performs an HTML import
func htmlImport(httpClient *http.Client, parent dom.Selection, serverbase string) error {
	imports := parent.Find("wimport").Elements()
	if len(imports) == 0 {
		return nil
	}

	queueChan := make(chan bool, len(imports))
	finishChan := make(chan bool, 1)

	for _, elem := range imports {
		src, ok := elem.Attr("src")
		if !ok {
			return dom.ElementError(elem, `wimport element has no "src" attribute`)
		}

		var err error
		go func(elem dom.Selection) {
			var html string
			html, err = getHtmlFile(httpClient, serverbase, src)
			if err != nil {
				return
			}

			// the go html parser will refuse to work if the content is only text, so
			// we put a wrapper here
			ne := parent.NewFragment("<pendingimport>" + html + "</pendingimport>")
			elem.ReplaceWith(ne)

			err = htmlImport(httpClient, ne, serverbase)
			if err != nil {
				return
			}

			ne.Unwrap()

			queueChan <- true
			if len(queueChan) == len(imports) {
				finishChan <- true
			}
		}(elem)

		if err != nil {
			finishChan <- true
			return err
		}
	}
	<-finishChan

	return nil
}

func (wd *wade) init() {
	bind.RegisterInternalHelpers(wd.pm, wd.binding)
}

// Start starts the real operation, meant to be called at the end of everything.
func (wd *wade) start() {
	wd.pm.prepare()
}
