package app

import (
	"fmt"
	"path"

	"github.com/gopherjs/gopherjs/js"

	"github.com/phaikawl/wade"
	"github.com/phaikawl/wade/core"
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/libs/http"
	"github.com/phaikawl/wade/markman"
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
		History() wade.History
		Bootstrap(*Application)
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
		Register   Registration
		Config     Config
		Http       *http.Client
		PageMgr    *wade.PageManager
		bindEngine *core.Binding
		markupMgr  *markman.MarkupManager
	}

	fetcher struct {
		serverBase string
		http       *http.Client
	}
)

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

func (app *Application) Router() wade.Router {
	return app.PageMgr.RouteMgr()
}

func (app *Application) Start(appMain Main) (err error) {
	appMain.Main(app)
	app.PageMgr.Start()

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

func (app *Application) AddPageGroup(pageGroup wade.PageGroup) {
	app.PageMgr.AddPageGroup(pageGroup)
}

// New creates the app
func New(config Config, rb RenderBackend) (app *Application) {
	httpClient := http.NewClient(rb.HttpBackend())
	http.SetDefaultClient(httpClient)

	bindEngine := core.NewBindEngine(DefaultHelpers)
	markupMgr := markman.New(rb.Document(), fetcher{
		http:       httpClient,
		serverBase: config.ServerBase,
	})

	app = &Application{
		Config:     config,
		Http:       httpClient,
		markupMgr:  markupMgr,
		bindEngine: bindEngine,
		PageMgr: wade.NewPageManager(config.BasePath, rb.History(),
			markupMgr, bindEngine),
	}

	rb.Bootstrap(app)

	return
}
