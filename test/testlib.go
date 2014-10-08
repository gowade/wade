package test

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/phaikawl/wade"
	"github.com/phaikawl/wade/bind"
	"github.com/phaikawl/wade/dom"
	gqdom "github.com/phaikawl/wade/dom/goquery"
	"github.com/phaikawl/wade/libs/http"
	"github.com/phaikawl/wade/rbackend/serverside"
)

type (
	PageView struct {
		Document dom.Selection
	}

	TestApp struct {
		*wade.Application
		View    PageView
		started bool
	}
)

func (v PageView) Title() string {
	return v.Document.Find("head title").Text()
}

func (app *TestApp) GoTo(path string) {
	if !app.started {
		panic(fmt.Errorf("Application has not been started."))
	}

	found := app.Services.PageManager.GoToUrl(path)
	if !found {
		panic(fmt.Errorf(`Page not found for "%v"`, path))
	}
}

func (app *TestApp) Start() {
	app.started = true
	app.Application.Start()
}

func NewTestApp(t *testing.T, conf wade.AppConfig,
	appFn wade.AppFunc,
	indexFile string,
	httpMock http.Backend) (app *TestApp, err error) {

	sourcebytes := []byte{}
	if indexFile != "" {
		var iFile io.Reader
		iFile, err = os.Open(indexFile)
		if err != nil {
			return
		}

		sourcebytes, err = ioutil.ReadAll(iFile)
		if err != nil {
			return
		}
	}

	document := gqdom.GetDom().NewDocument(string(sourcebytes[:]))
	wapp, err := wade.NewApp(conf, appFn, wade.RenderBackend{
		JsBackend: &serverside.JsBackend{
			NoopJsWatcher: bind.NoopJsWatcher{},
			JsHistory:     wade.NewNoopHistory(conf.BasePath),
		},
		Document:    document,
		HttpBackend: httpMock,
	})

	app = &TestApp{
		Application: wapp,
		View: PageView{
			Document: document,
		},
	}

	return
}

func NewDummyApp(t *testing.T, httpMock http.Backend) (app *TestApp, err error) {
	return NewTestApp(t, wade.AppConfig{}, func(app *wade.Application) {}, "", httpMock)
}
