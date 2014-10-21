package wade

import (
	"reflect"

	"github.com/gopherjs/gopherjs/js"
	"github.com/phaikawl/wade/bind"
	"github.com/phaikawl/wade/com"
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/libs/http"
)

var (
	DevMode = true
)

func init() {
	if js.Global == nil {
		js.Global = newStubJsValue(nil)
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
		tm         *com.Manager
		tcontainer dom.Selection
		binding    *bind.Binding
		serverBase string
		customTags map[string]map[string]com.Prototype
	}

	Registration struct {
		w *wade
	}

	JsBackend interface {
		DepChecker
		History() History
		bind.WatchBackend
		WebStorages() (Storage, Storage)
	}

	Application struct {
		Register           Registration
		Router             *Router
		Config             AppConfig
		Services           AppServices
		wade               *wade
		main               AppFunc
		errChan            chan error
		baseComponentProto com.Prototype
	}

	BaseProto struct {
		com.BaseProto
		App *Application
	}
)

func (app *Application) initServices(pm PageManager, rb RenderBackend, httpClient *http.Client) {
	los, ses := rb.JsBackend.WebStorages()
	app.Services = AppServices{
		Http:           httpClient,
		LocalStorage:   los,
		SessionStorage: ses,
		PageManager:    pm,
	}
}

func (app *Application) Checkpoint() {
	app.wade.binding.Watcher().Checkpoint()
}

func (app *Application) CurrentPage() *PageScope {
	return app.Services.PageManager.CurrentPage()
}

func (app *Application) Start() (err error) {
	app.wade.start()

	select {
	case err = <-app.errChan:
		return err
	default:
	}

	go func() {
		for {
			<-http.ResponseChan
			app.Checkpoint()
		}
	}()

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

func (app *Application) ComponentInit(proto com.Prototype) {
	v := reflect.ValueOf(proto)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	bp := v.FieldByName("BaseProto")

	if bp.IsValid() {
		ap := reflect.ValueOf(app.baseComponentProto)
		if bp.Kind() == reflect.Ptr && ap.Type().AssignableTo(bp.Type()) {
			bp.Set(ap)
		}

		if ap.Type().Elem().AssignableTo(bp.Type()) {
			bp.Set(ap.Elem())
		}
	}
}

func (r Registration) Components(components ...com.Spec) {
	err := r.w.tm.RegisterComponents(components)
	if err != nil {
		panic(err)
	}
}

// RegisterController adds a new controller function for the specified
// page / page group.
func (r Registration) Controller(displayScope string, fn PageControllerFunc) {
	r.w.pm.registerController(displayScope, fn)
}

func (r Registration) PageGroup(id string, children []string) {
	r.w.pm.registerPageGroup(id, children)
}

// loadHtml loads html from script[type='text/wadin'], performs html imports on it
func loadHtml(document dom.Selection, httpClient *http.Client, serverBase string) (dom.Selection, error) {
	templateContainer := document.NewRootFragment()
	temp := document.Find("script[type='text/wadin']").First()
	templateContainer.Append(document.NewFragment(temp.Text()))

	err := htmlImport(httpClient, templateContainer, serverBase)
	temp.SetText(templateContainer.Html())
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
		Config:  config,
		wade:    nil,
		main:    appFn,
		errChan: make(chan error),
	}

	app.baseComponentProto = &BaseProto{com.BaseProto{}, app}

	tm := com.NewManager(templateContainer)
	binding := bind.NewBindEngine(app, tm, rb.JsBackend)

	wd := &wade{
		pm:         newPageManager(app, rb.JsBackend.History(), document, templateContainer, binding),
		tm:         tm,
		binding:    binding,
		tcontainer: templateContainer,
		serverBase: config.ServerBase,
	}

	app.wade = wd
	app.Register = Registration{wd}
	app.Router = app.wade.pm.router
	wd.init()

	app.initServices(wd.pm, rb, httpClient)

	appFn(app)

	return
}

func (wd *wade) init() {
	bind.RegisterInternalHelpers(wd.pm, wd.binding)
}

// Start starts the real operation
func (wd *wade) start() {
	wd.pm.prepare()
}
