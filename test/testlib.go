package test

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
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
		Html dom.Selection
	}

	TestApp struct {
		*wade.Application
		View    PageView
		started bool
	}
)

// Get the page's current title
func (v PageView) Title() string {
	return v.Html.Find("head title").Text()
}

// Find is simply a wrapper of v.Html.Find()
func (v PageView) Find(queryStr string) dom.Selection {
	return v.Html.Find(queryStr)
}

// TriggerEvent triggers a given event on the selected elements
func (v PageView) triggerEvent(selection dom.Selection, event Event) {
	for _, e := range selection.Elements() {
		event.Event().propaStopped = false
		event.Event().target = e
		triggerRec(e, event)
	}
}

// CheckText successively checks if each selected element has the corresponding
// list of text content.
//
// Returns an error if some content is not found.
func (v PageView) CheckText(selection dom.Selection, textLists [][]string) (err error) {
	elems := selection.Elements()
	for i, textList := range textLists {
		for _, text := range textList {
			if !strings.Contains(elems[i].Text(), text) {
				err = fmt.Errorf(`%vth element does not have text content "%v"`, i, text)
				return
			}
		}
	}

	return
}

// TriggerEvent triggers an event with watcher.Apply()
// this triggers a DigestAll() so that the view is updated with the changed data
func (app *TestApp) TriggerEvent(selection dom.Selection, event Event) {
	app.View.triggerEvent(selection, event)
}

// GoTo navigates the app to the given path
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

func (app *TestApp) Digest() {
	app.CurrentPage().Watcher.DigestAll()
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
			BasicWatchBackend: bind.BasicWatchBackend{},
			JsHistory:         wade.NewNoopHistory(conf.BasePath),
		},
		Document:    document,
		HttpBackend: httpMock,
	})

	app = &TestApp{
		Application: wapp,
		View: PageView{
			Html: document,
		},
	}

	return
}

func NewDummyApp(t *testing.T, httpMock http.Backend) (app *TestApp, err error) {
	return NewTestApp(t, wade.AppConfig{}, func(app *wade.Application) {}, "", httpMock)
}
