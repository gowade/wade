package rt

import (
	"time"

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
		Setup(page.Router)
		Main()
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
		renderQ       chan (chan bool)
	}

	ComponentProto interface {
		Init()
		Update(*core.VNode)
	}

	BaseProto struct{}
)

func (m BaseProto) Init()                   {}
func (m BaseProto) Update(node *core.VNode) {}

func (app *Application) Document() dom.Selection {
	return app.PageMgr.Document()
}

func (app *Application) Render() chan bool {
	ch := make(chan bool, 1)
	app.renderQ <- ch
	return ch
}

func (app *Application) Router() page.Router {
	return app.PageMgr.Router()
}

func (app *Application) renderLoop() {
	for {
		ch := <-app.renderQ
		time.Sleep(70 * time.Millisecond)
		app.PageMgr.Render()
		cont := true
		for cont {
			select {
			case cc := <-app.renderQ:
				cc <- true
			default:
				cont = false
			}
		}
		ch <- true
	}
}

func (app *Application) Start(appMain Main) (err error) {
	SetApp(app)

	appMain.Setup(app.Router())
	appMain.Main()

	app.PageMgr.Start()
	go app.renderLoop()
	app.renderBackend.AfterReady(app)

	return
}

func (app *Application) AddPageGroup(pageGroup page.PageGroup) {
	app.PageMgr.AddPageGroup(pageGroup)
}

// New creates the app
func NewApp(config Config, rb RenderBackend) (app *Application) {
	httpClient := http.NewClient(rb.HttpBackend())
	http.SetDefaultClient(httpClient)

	app = &Application{
		Config: config,
		Http:   httpClient,
		PageMgr: page.NewPageManager(config.BasePath, rb.History(),
			rb.Document()),
		renderBackend: rb,
		renderQ:       make(chan (chan bool), 20),
	}

	rb.Bootstrap(app)

	return
}
