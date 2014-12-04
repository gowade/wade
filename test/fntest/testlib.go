package test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/phaikawl/wade/app"
	//"github.com/phaikawl/wade/core"
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/libs/http"
	"github.com/phaikawl/wade/platform/serverside"
)

type (
	PageView struct {
		Document dom.Selection
	}

	TestApp struct {
		PageView
		*app.Application
		started bool
	}
)

// Get the page's current title
func (v PageView) Title() string {
	return v.Document.Find("head title").Text()
}

// Find is simply a wrapper of v.Html.Find()
func (v PageView) Find(queryStr string) dom.Selection {
	return v.Document.Find(queryStr)
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
	app.PageView.triggerEvent(selection, event)
}

// GoTo navigates the app to the given path
func (app *TestApp) GoTo(path string) {
	if !app.started {
		panic(fmt.Errorf("Application has not been started."))
	}

	found := app.PageMgr.GoToUrl(path)
	if !found {
		panic(fmt.Errorf(`Page not found for "%v"`, path))
	}
}

func (app *TestApp) Start(appMain app.Main) error {
	app.started = true
	return app.Application.Start(appMain)
}

func NewTestApp(conf app.Config,
	initialPath string,
	indexFile string,
	httpMock http.Backend) *TestApp {

	var document io.Reader
	if indexFile != "" {
		var err error
		document, err = os.Open(indexFile)
		if err != nil {
			panic(err)
		}
	} else {
		document = bytes.NewReader([]byte(`<html>
			<head></head>
			<body w-app-container="">
			</body>
		</html>`))
	}

	wapp := serverside.NewApp(conf, document, initialPath, httpMock)

	return &TestApp{
		Application: wapp,
		PageView: PageView{
			Document: wapp.Document(),
		},
	}
}

type dummyMain struct{}

func (dm dummyMain) Main(app *app.Application) {
}

func StartDummyApp(httpMock http.Backend) (*TestApp, error) {
	app := NewTestApp(app.Config{}, "/", "", httpMock)
	err := app.Start(dummyMain{})
	return app, err
}
