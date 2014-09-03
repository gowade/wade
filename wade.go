package wade

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/phaikawl/wade/bind"
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/libs/http"
)

var (
	WadeDevMode = true
)

type (
	RenderBackend struct {
		JsBackend   JsBackend
		Document    dom.Selection
		HttpBackend http.Backend
	}

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
		WebStorages() (Storage, Storage)
	}
)

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
func (r registry) RegisterCustomTags(customTags ...CustomTag) {
	r.w.tm.registerTags(customTags)
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

func initServices(pm PageManager, rb RenderBackend) {
	AppServices.Http = http.NewClient(rb.HttpBackend)
	AppServices.LocalStorage, AppServices.SessionStorage = rb.JsBackend.WebStorages()
	AppServices.PageManager = pm
}

// loadHtml loads html from script[type='text/wadin'], performs html imports
// and sets the resulting contents back to the script element
func loadHtml(document dom.Selection, httpClient *http.Client, serverBase string) (dom.Selection, error) {
	templateContainer := document.NewRootFragment()
	temp := document.Find("script[type='text/wadin']").First()
	templateContainer.Append(document.NewFragment(temp.Text()))

	err := htmlImport(httpClient, templateContainer, serverBase)
	temp.SetHtml(templateContainer.Html())
	return templateContainer, err
}

// StartApp initializes the app.
//
// "appFn" is the main function for your app.
func StartApp(config AppConfig, appFn AppFunc, rb RenderBackend) error {
	jsDepCheck(rb.JsBackend)
	http.SetDefaultClient(http.NewClient(rb.HttpBackend))
	document := rb.Document

	templateContainer, err := loadHtml(document, http.DefaultClient(), config.ServerBase)

	if err != nil {
		return err
	}

	tm := newCustagMan()
	binding := bind.NewBindEngine(tm, rb.JsBackend)

	wd := &wade{
		errChan:    make(chan error),
		pm:         newPageManager(rb.JsBackend.History(), config, document, templateContainer, binding),
		tm:         tm,
		binding:    binding,
		tcontainer: templateContainer,
		serverBase: config.ServerBase,
		customTags: make(map[string]map[string]CustomElemProto),
	}

	wd.init()
	initServices(wd.pm, rb)

	appFn(registry{wd})
	err = wd.loadCustomTagDefs()
	if err != nil {
		return err
	}

	wd.start()

	select {
	case err = <-wd.errChan:
		return err
	default:
	}

	return nil
}

func (wd *wade) init() {
	bind.RegisterInternalHelpers(wd.pm, wd.binding)
}

func (w *wade) loadCustomTagDefs() (err error) {
	for _, d := range w.tcontainer.Find("wdefine").Elements() {
		if tagname, ok := d.Attr("tagname"); ok {
			tag, ok := w.tm.custags[tagname]
			if !ok {
				err = dom.ElementError(d, fmt.Sprintf(`Custom tag "%v" has not been registered.`, tagname))
				return
			}

			tag.Html = d.Html()
			w.tm.custags[tagname] = tag
		}
	}

	return
}

// Start starts the real operation, meant to be called at the end of everything.
func (wd *wade) start() {
	wd.pm.prepare()
}
