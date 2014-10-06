package wade

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/phaikawl/wade/bind"
	"github.com/phaikawl/wade/custom"
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/libs/http"
)

var (
	WadeDevMode = true
)

func init() {
	if js.Global == nil {
		js.Global = NewStubJsValue(nil)
		ClientSide = false
		return
	}

	ClientSide = !js.Global.Get("window").IsUndefined()
}

type (
	RenderBackend struct {
		JsBackend   JsBackend
		Document    dom.Selection
		HttpBackend http.Backend
	}

	wade struct {
		errChan    chan error
		pm         *pageManager
		tm         *custom.TagManager
		tcontainer dom.Selection
		binding    *bind.Binding
		serverBase string
		customTags map[string]map[string]custom.TagPrototype
	}

	registry struct {
		w *wade
	}

	JsBackend interface {
		DepChecker
		History() History
		bind.JsWatcher
		WebStorages() (Storage, Storage)
	}
)

func (r registry) RegisterCustomTags(customTags ...custom.HtmlTag) {
	err := r.w.tm.RegisterTags(customTags)
	if err != nil {
		panic(err)
	}
}

// RegisterController adds a new controller function for the specified
// page / page group.
func (r registry) RegisterController(displayScope string, fn PageControllerFunc) {
	r.w.pm.registerController(displayScope, fn)
}

// RegisterDisplayScopes registers the pages and page groups
func (r registry) RegisterDisplayScopes(pages []PageDesc, pageGroups []PageGroupDesc) {
	r.w.pm.registerDisplayScopes(pages, pageGroups)
}

// RegisterNotFoundPage registers the page that is used when no page matches the url
func (r registry) RegisterNotFoundPage(pageid string) {
	r.w.pm.SetNotFoundPage(pageid)
}

func initServices(pm PageManager, rb RenderBackend) {
	AppServices.Http = http.NewClient(rb.HttpBackend)
	AppServices.LocalStorage, AppServices.SessionStorage = rb.JsBackend.WebStorages()
	AppServices.PageManager = pm
}

// loadHtml loads html from script[type='text/wadin'], performs html imports on it
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
	AppConf = config

	jsDepCheck(rb.JsBackend)
	http.SetDefaultClient(http.NewClient(rb.HttpBackend))
	document := rb.Document

	templateContainer, err := loadHtml(document, http.DefaultClient(), config.ServerBase)

	if err != nil {
		return err
	}

	tm := custom.NewTagManager()
	binding := bind.NewBindEngine(tm, rb.JsBackend)

	wd := &wade{
		errChan:    make(chan error),
		pm:         newPageManager(rb.JsBackend.History(), config, document, templateContainer, binding),
		tm:         tm,
		binding:    binding,
		tcontainer: templateContainer,
		serverBase: config.ServerBase,
		customTags: make(map[string]map[string]custom.TagPrototype),
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
			err = w.tm.RedefTag(tagname, d.Html())
			if err != nil {
				err = dom.ElementError(d, err.Error())
				return
			}
		}
	}

	return
}

// Start starts the real operation, meant to be called at the end of everything.
func (wd *wade) start() {
	wd.pm.prepare()
}
