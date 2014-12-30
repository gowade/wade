package app

import (
	"fmt"
	"path"

	"github.com/gopherjs/gopherjs/js"

	"github.com/phaikawl/wade/binders"
	"github.com/phaikawl/wade/core"
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/libs/http"
	"github.com/phaikawl/wade/markman"
	"github.com/phaikawl/wade/page"
)

var (
	ClientSide = js.Global != nil && !js.Global.Get("window").IsUndefined()
	DevMode    = true
	gApp       *Application
)

func App() *Application {
	if gApp == nil {
		panic("No appltication has been created.")
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
		bindEngine    *core.Binding
		markupMgr     *markman.MarkupManager
		renderBackend RenderBackend
		eventFinish   chan bool
	}

	fetcher struct {
		serverBase string
		http       *http.Client
	}

	ComponentProto struct {
		core.BaseProto
		App *Application
	}
)

func (app *Application) Document() dom.Selection {
	return app.markupMgr.Document()
}

func (app *Application) MarkupMgr() *markman.MarkupManager {
	return app.markupMgr
}

func (fetcher fetcher) FetchFile(file string) (data string, err error) {
	resp, err := fetcher.http.GET(path.Join(fetcher.serverBase, file))
	if resp.Failed() || err != nil {
		err = fmt.Errorf(`Failed to load HTML file "%v". Status code: %v. Error: %v.`, file, resp.StatusCode, err)
		return
	}

	data = resp.Data
	return
}

func (app *Application) Render() {
	app.markupMgr.Render()
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

	err = app.markupMgr.LoadView()
	if err != nil {
		return
	}

	app.PageMgr.Start()
	app.renderBackend.AfterReady(app)

	return
}

func (app *Application) AddComponent(cvList ...core.ComponentView) {
	for _, cv := range cvList {
		app.bindEngine.ComponentManager().Register(cv)
	}
}

func (app *Application) AddPageGroup(pageGroup page.PageGroup) {
	app.PageMgr.AddPageGroup(pageGroup)
}

// New creates the app
func New(config Config, rb RenderBackend) (app *Application) {
	httpClient := http.NewClient(rb.HttpBackend())
	http.SetDefaultClient(httpClient)

	markupMgr := markman.New(rb.Document(), fetcher{
		http:       httpClient,
		serverBase: config.ServerBase,
	})

	bindEngine := core.NewBindEngine(markupMgr.TemplateConverter(), defaultHelpers)
	binders.Install(bindEngine)

	app = &Application{
		Config:     config,
		Http:       httpClient,
		markupMgr:  markupMgr,
		bindEngine: bindEngine,
		PageMgr: page.NewPageManager(config.BasePath, rb.History(),
			markupMgr, bindEngine),
		renderBackend: rb,
		eventFinish:   make(chan bool),
	}

	defaultHelpers["url"] = app.PageMgr.PageUrl

	rb.Bootstrap(app)

	return
}
