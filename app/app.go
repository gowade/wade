package app

import (
	"github.com/gopherjs/gopherjs/js"

	"github.com/phaikawl/wade/core"
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/libs/http"
	"github.com/phaikawl/wade/page"
)

var (
	ClientSide = js.Global != nil && js.Global.Get("window") != js.Undefined
	DevMode    = true
	gApp       *Application
)

func App() *Application {
	if gApp == nil {
		panic("No application has been created.")
	}

	return gApp
}

func SetApp(app *Application) {
	gApp = app
}

type (
	RenderBackend interface {
		History() page.History
		Bootstrap(*Application)
		AfterReady(*Application)
		HttpBackend() http.Backend
		Document() dom.Selection
	}

	Main interface {
		Main(app *Application)
	}

	// Config is app configurations, used at the start
	Config struct {
		BasePath   string
		ServerBase string
	}

	Registration struct {
		app *Application
	}

	// Application
	Application struct {
		Register      Registration
		Config        Config
		Http          *http.Client
		PageMgr       *page.PageManager
		renderBackend RenderBackend
		eventFinish   chan bool
	}

	ComponentProto interface {
		Init()
		Update(*core.VNode)
	}

	ComponentModel struct {
		App *Application
	}
)

func (m ComponentModel) Init()                   {}
func (m ComponentModel) Update(node *core.VNode) {}

func (app *Application) Document() dom.Selection {
	return app.PageMgr.Document()
}

func (app *Application) Render() {
	app.PageMgr.Render()
}

func (app *Application) Router() page.Router {
	return app.PageMgr.RouteMgr()
}

func (app *Application) NotifyEventFinish() {
	app.eventFinish <- true
}

func (app *Application) EventFinished() chan bool {
	return app.eventFinish
}

func (app *Application) Start(appMain Main) (err error) {
	SetApp(app)

	appMain.Main(app)

	app.PageMgr.Start()
	app.renderBackend.AfterReady(app)

	return
}

func (app *Application) AddPageGroup(pageGroup page.PageGroup) {
	app.PageMgr.AddPageGroup(pageGroup)
}

// New creates the app
func New(config Config, rb RenderBackend) (app *Application) {
	httpClient := http.NewClient(rb.HttpBackend())
	http.SetDefaultClient(httpClient)

	app = &Application{
		Config: config,
		Http:   httpClient,
		PageMgr: page.NewPageManager(config.BasePath, rb.History(),
			rb.Document()),
		renderBackend: rb,
		eventFinish:   make(chan bool),
	}

	rb.Bootstrap(app)

	return
}
