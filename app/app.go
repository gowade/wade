package app

import (
	"fmt"
	"path"
	"runtime"

	"github.com/gopherjs/gopherjs/js"

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
	}

	fetcher struct {
		serverBase string
		http       *http.Client
	}
)

func (app *Application) Document() dom.Selection {
	return app.markupMgr.Document()
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

func (app *Application) Start(appMain Main) (err error) {
	defer func() {
		if r := recover(); r != nil {
			trace := make([]byte, 1024)
			count := runtime.Stack(trace, true)
			err = fmt.Errorf("Error while starting and rendering the app: %s\nStack of %d bytes: %s\n", r, count, trace)
		}
	}()

	appMain.Main(app)
	app.PageMgr.Start()
	app.renderBackend.AfterReady(app)

	go func() {
		for {
			<-http.ResponseChan
			app.Render()
		}
	}()

	return
}

func (app *Application) AddComponent(cv core.ComponentView) {
	app.bindEngine.ComponentManager().Register(cv)
}

func (app *Application) AddPageGroup(pageGroup page.PageGroup) {
	app.PageMgr.AddPageGroup(pageGroup)
}

// New creates the app
func New(config Config, rb RenderBackend) (app *Application) {
	httpClient := http.NewClient(rb.HttpBackend())
	http.SetDefaultClient(httpClient)

	bindEngine := core.NewBindEngine(markman.TemplateConverter{rb.Document()}, DefaultHelpers)
	markupMgr := markman.New(rb.Document(), fetcher{
		http:       httpClient,
		serverBase: config.ServerBase,
	})

	app = &Application{
		Config:     config,
		Http:       httpClient,
		markupMgr:  markupMgr,
		bindEngine: bindEngine,
		PageMgr: page.NewPageManager(config.BasePath, rb.History(),
			markupMgr, bindEngine),
		renderBackend: rb,
	}

	rb.Bootstrap(app)

	return
}
