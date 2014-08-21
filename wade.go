package wade

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/phaikawl/wade/bind"
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/dom/jquery"
	"github.com/phaikawl/wade/jsbackend"
	"github.com/phaikawl/wade/libs/http"
	"github.com/phaikawl/wade/libs/http/clientside"
	"github.com/phaikawl/wade/libs/pdata"
)

var (
	WadeDevMode = true
)

type (
	wade struct {
		errChan    chan error
		pm         *pageManager
		tm         *custagMan
		tcontainer dom.Selection
		binding    *bind.Binding
		serverBase string
		customTags map[string]map[string]CustomElemProto
	}

	registry struct {
		w *wade
	}

	JsBackend interface {
		DepChecker
		History() History
		Watch(modelRefl reflect.Value, field string, callback func())
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
func (r registry) RegisterCustomTags(srcFile string, protomap map[string]CustomElemProto) {
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

// RegisterDisplayScopes registers the given maps of pages and page groups
func (r registry) RegisterDisplayScopes(pages []PageDesc, pageGroups []PageGroupDesc) {
	r.w.pm.registerDisplayScopes(pages, pageGroups)
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
		templateContainer.Append(document.NewFragment(part.Html()))
	}

	return templateContainer
}

func (wd *wade) loadCustomElemsHTML(httpClient *http.Client, dombackend dom.Dom) {
	queueChan := make(chan bool, 100)
	finishChan := make(chan bool, len(wd.customTags))
	for srcFile, protoMap := range wd.customTags {
		go func(srcFile string, protoMap map[string]CustomElemProto) {
			src, err := wd.getHtml(httpClient, srcFile)

			queueChan <- true

			if len(queueChan) == len(wd.customTags) {
				close(queueChan)
				finishChan <- true
			}

			if err != nil {
				wd.errChan <- fmt.Errorf(`Cannot load custom tag HTML file "%v".`, srcFile)
				return
			}

			wd.tm.registerTags(dombackend.NewFragment(src).Filter("welement").Elements(), protoMap)
		}(srcFile, protoMap)
	}
	<-finishChan
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

	tm := newCustagMan()
	binding := bind.NewBindEngine(tm, jsb)

	wd := &wade{
		errChan:    make(chan error),
		pm:         newPageManager(jsb.History(), config, document, templateContainer, binding),
		tm:         tm,
		binding:    binding,
		tcontainer: templateContainer,
		serverBase: config.ServerBase,
		customTags: make(map[string]map[string]CustomElemProto),
	}

	wd.init()
	initServices(wd.pm)

	appFn(registry{wd})

	wd.loadCustomElemsHTML(http.DefaultClient(), dombackend)

	wd.start()

	select {
	case err = <-wd.errChan:
		return err
	}

	return nil
}

func (wd *wade) init() {
	bind.RegisterInternalHelpers(wd.pm, wd.binding)
}

// Start starts the real operation, meant to be called at the end of everything.
func (wd *wade) start() {
	wd.pm.prepare()
}
