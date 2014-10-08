package wade

import (
	"reflect"

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

	Application struct {
		Register    Registration
		Config      AppConfig
		Services    *AppServices
		wade        *wade
		main        AppFunc
		errChan     chan error
		baseCEProto *BaseProto
	}

	//Base custom element prototype
	BaseProto struct {
		App *Application
	}
)

func (b BaseProto) Init() error                                  { return nil }
func (b BaseProto) ProcessContents(ctl custom.ContentsCtl) error { return nil }
func (b BaseProto) Update(ctl custom.ElemCtl) error              { return nil }

func (app *Application) initServices(pm PageManager, rb RenderBackend, httpClient *http.Client) {
	app.Services.Http = httpClient
	app.Services.LocalStorage, app.Services.SessionStorage = rb.JsBackend.WebStorages()
	app.Services.PageManager = pm
}

func (app *Application) CurrentPage() *Scope {
	return app.Services.PageManager.CurrentPage()
}

func (app *Application) Start() (err error) {
	app.wade.start()

	select {
	case err = <-app.errChan:
		return err
	default:
	}

	return
}

func (app *Application) ErrChanPut(err error) {
	select {
	case app.errChan <- err:
	default:
		panic(err)
	}
}

func (app *Application) Http() *http.Client {
	return app.Services.Http
}

func (app *Application) CustomElemInit(proto custom.TagPrototype) {
	v := reflect.ValueOf(proto)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	bp := v.FieldByName("BaseProto")

	if bp.IsValid() {
		ap := reflect.ValueOf(app.baseCEProto)
		if bp.Kind() == reflect.Ptr && ap.Type().AssignableTo(bp.Type()) {
			bp.Set(ap)
		}

		if ap.Type().Elem().AssignableTo(bp.Type()) {
			bp.Set(ap.Elem())
		}
	}
}

func (r registry) CustomTags(customTags ...custom.HtmlTag) {
	err := r.w.tm.RegisterTags(customTags)
	if err != nil {
		panic(err)
	}
}

// RegisterController adds a new controller function for the specified
// page / page group.
func (r registry) Controller(displayScope string, fn PageControllerFunc) {
	r.w.pm.registerController(displayScope, fn)
}

// RegisterDisplayScopes registers the pages and page groups
func (r registry) DisplayScopes(pages []PageDesc, pageGroups []PageGroupDesc) {
	r.w.pm.registerDisplayScopes(pages, pageGroups)
}

// RegisterNotFoundPage registers the page that is used when no page matches the url
func (r registry) NotFoundPage(pageid string) {
	r.w.pm.SetNotFoundPage(pageid)
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

// StartApp initializes the app
//
// "appFn" is the main function for your app.
func NewApp(config AppConfig, appFn AppFunc, rb RenderBackend) (app *Application, err error) {
	jsDepCheck(rb.JsBackend)
	document := rb.Document

	httpClient := http.NewClient(rb.HttpBackend)
	http.SetDefaultClient(httpClient)
	templateContainer, err := loadHtml(document, httpClient, config.ServerBase)

	if err != nil {
		return
	}

	app = &Application{
		Config:   config,
		Services: &AppServices{},
		wade:     nil,
		main:     appFn,
		errChan:  make(chan error),
	}
	app.baseCEProto = &BaseProto{app}

	tm := custom.NewTagManager()
	binding := bind.NewBindEngine(app, tm, rb.JsBackend)

	wd := &wade{
		pm:         newPageManager(app, rb.JsBackend.History(), document, templateContainer, binding),
		tm:         tm,
		binding:    binding,
		tcontainer: templateContainer,
		serverBase: config.ServerBase,
		customTags: make(map[string]map[string]custom.TagPrototype),
	}

	app.wade = wd
	app.Register = registry{wd}
	wd.init()

	app.initServices(wd.pm, rb, httpClient)

	appFn(app)
	err = wd.loadCustomTagDefs()
	if err != nil {
		return
	}

	return
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

// Start starts the real operation
func (wd *wade) start() {
	wd.pm.prepare()
}
